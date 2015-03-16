package db

import (
	"github.com/garyburd/redigo/redis"
	"sandswind/marble/log"
	"sync"
	"time"
)

const REQ_CHAN_SIZE = 4000

type redisResp struct {
	reply interface{}
	err   error
}

type redisReq struct {
	ch   chan redisResp
	cmd  string
	args []interface{}
}

type RedisWorker struct {
	redisReqChan    chan redisReq
	newAddrChan     chan []string
	retryResultChan chan redis.Conn
	retryCancelChan chan bool
	addr            []string
	pwd             string
	conn            redis.Conn
	currAddr        int
	addrCount       int
	err             error
	reqs            []chan redisResp
	newAddr         bool
}

var redisChanPool = sync.Pool{
	New: func() interface{} {
		return make(chan redisResp, 1)
	},
}

func NewRedisWorker(addr []string, password string) *RedisWorker {
	return &RedisWorker{
		make(chan redisReq, REQ_CHAN_SIZE),
		make(chan []string, 1),
		make(chan redis.Conn, 1),
		nil,
		addr,
		password,
		nil,
		0,
		len(addr),
		nil,
		make([]chan redisResp, 0, 100),
		false,
	}
}

func (worker *RedisWorker) NewAddr(addr []string) {
	worker.newAddrChan <- addr
}

func (worker *RedisWorker) RedisExec(cmd string, args ...interface{}) (interface{}, error) {
	ch := redisChanPool.Get().(chan redisResp)
	worker.redisReqChan <- redisReq{ch, cmd, args}
	resp := <-ch
	redisChanPool.Put(ch)
	return resp.reply, resp.err
}

func (worker *RedisWorker) closer() {
	if worker.conn != nil {
		worker.conn.Close()
		worker.conn = nil
	}
}

// drain request when disconnect
func (worker *RedisWorker) drain() {
	log.Debug("drain")
	for _, r := range worker.reqs {
		if r != nil {
			log.Debug("drain|resp|%v", worker.err)
			r <- redisResp{nil, worker.err}
		}
	}
	worker.reqs = worker.reqs[0:0]
	count := 0
	for {
		if count > REQ_CHAN_SIZE {
			// too many to drain
			return
		}
		count++
		select {
		case req := <-worker.redisReqChan:
			log.Debug("drain|resp|%v", worker.err)
			req.ch <- redisResp{nil, worker.err}
		case req := <-worker.newAddrChan:
			worker.receiveAddr(req)
			return
		default:
			// empty
			return
		}
	}
}

func (worker *RedisWorker) reopen() {
	worker.closer()
	log.Info("connect|%v|%v", worker.currAddr, worker.addr[worker.currAddr])
	addr := worker.addr[worker.currAddr]
	if worker.currAddr != 0 && worker.retryCancelChan == nil {
		worker.retryCancelChan = make(chan bool, 10)
		worker.retryResultChan = make(chan redis.Conn, 1)
		go tryConn(worker.addr[0], worker.pwd, worker.retryResultChan, worker.retryCancelChan)
	}
	timeout := 5 * time.Second
	worker.conn, worker.err = redis.DialTimeout("tcp", addr, 2*time.Second, timeout, timeout)
	if worker.err != nil {
		log.Error("connect|%v", worker.err)
		time.Sleep(1 * time.Second)
		worker.currAddr++
		worker.currAddr = worker.currAddr % worker.addrCount
		return
	}
	if worker.pwd != "" {
		if _, worker.err = worker.conn.Do("AUTH", worker.pwd); worker.err != nil {
			worker.closer()
			log.Error("auth|%v", worker.err)
			time.Sleep(1 * time.Second)
			worker.currAddr++
			worker.currAddr = worker.currAddr % worker.addrCount
			return
		} else {
			log.Debug("reopen|nopass")
		}
	}
	log.Info("connect|ok|%v", worker.addr[worker.currAddr])
}

func (worker *RedisWorker) receiveReq(req redisReq) bool {
	log.Debug("receiveReq|%v|%v", req.cmd, req.args)
	worker.reqs = append(worker.reqs, req.ch)
	worker.err = worker.conn.Send(req.cmd, req.args...)
	if worker.err != nil {
		log.Error("send|%v", worker.err)
		return true
	}
	if len(worker.reqs) == cap(worker.reqs) {
		// full
		return true
	}
	return false
}

func (worker *RedisWorker) receiveAddr(addr []string) {
	worker.addr = addr
	worker.currAddr = 0
	log.Debug("receiveAddr|%v", addr)
	worker.newAddr = true
	if worker.retryCancelChan != nil {
		worker.retryCancelChan <- true
		worker.retryCancelChan = nil
	}
}

func (worker *RedisWorker) receive() {
	for {
		select {
		case req := <-worker.redisReqChan:
			if worker.receiveReq(req) {
				return
			}
		case req := <-worker.newAddrChan:
			worker.receiveAddr(req)
			return
		default:
			if len(worker.reqs) != 0 {
				return
			}
			t := time.NewTimer(5 * time.Second)
			select {
			case <-t.C:
				if _, worker.err = worker.conn.Do("get", "__ping"); worker.err != nil {
					log.Error("ping|%v", worker.err)
					return
				}
			case req := <-worker.newAddrChan:
				t.Stop()
				worker.receiveAddr(req)
				return
			case req := <-worker.redisReqChan:
				t.Stop()
				if worker.receiveReq(req) {
					return
				}
			case conn := <-worker.retryResultChan:
				t.Stop()
				if worker.retryCancelChan != nil && conn != nil {
					worker.switchToFirst(conn)
				}
				return
			}
		}
	}
}

func (worker *RedisWorker) process() {
	if worker.err = worker.conn.Flush(); worker.err != nil {
		log.Error("process|Flush|%v", worker.err)
		return
	}
	for i := 0; i < len(worker.reqs); i++ {
		var v interface{}
		v, worker.err = worker.conn.Receive()
		if worker.err != nil {
			if _, ok := worker.err.(redis.Error); ok {
				log.Error("receive|redis_err|%v", worker.err)
				worker.reqs[i] <- redisResp{nil, worker.err}
				worker.reqs[i] = nil
				// redis data error, send to user
				worker.err = nil
			} else {
				log.Error("receive|net_err|%v", worker.err)
				return
			}
		} else {
			log.Debug("process|Receive|%v", v)
			worker.reqs[i] <- redisResp{v, nil}
			worker.reqs[i] = nil
		}
	}
	worker.reqs = worker.reqs[0:0]
}

func tryConn(addr string, pwd string, ch chan redis.Conn, cancel chan bool) {
	log.Info("tryConn|%v", addr)
	// slave addr, try switch back every 5 minute
	timeout := 5 * time.Second
	for {
		select {
		case <-cancel:
			log.Info("tryConn|Cancel|%v", addr)
			ch <- nil
			return
		default:
			time.Sleep(30 * time.Second)
			if c, err := redis.DialTimeout("tcp", addr, 2*time.Second, timeout, timeout); err != nil {
				log.Error("tryConn|%v", err)
			} else {
				if pwd != "" {
					if _, err = c.Do("AUTH", pwd); err != nil {
						log.Error("tryConn|auth|%v", err)
						c.Close()
						continue
					}
				}
				if _, err = c.Do("get", "__test"); err != nil {
					log.Error("try_conn|get|%v", err)
					c.Close()
					continue
				}
				ch <- c
				log.Info("tryConn|ok|%v", addr)
				return
			}
		}
	}
}

func (worker *RedisWorker) switchToFirst(conn redis.Conn) {
	log.Info("switchToFirst")
	worker.retryCancelChan = nil
	worker.closer()
	worker.currAddr = 0
	worker.conn = conn
}

func (worker *RedisWorker) Start() {
	defer worker.closer()
	for {
		if worker.err != nil {
			worker.drain()
			worker.closer()
			worker.err = nil
		}
		select {
		case conn := <-worker.retryResultChan:
			if worker.retryCancelChan != nil && conn != nil {
				worker.switchToFirst(conn)
			}
		default:
		}
		if worker.conn == nil || worker.newAddr {
			worker.newAddr = false
			if worker.reopen(); worker.err != nil {
				continue
			}
		}
		if worker.receive(); worker.err != nil {
			continue
		}
		if len(worker.reqs) == 0 {
			continue
		}
		worker.process()
	}
}
