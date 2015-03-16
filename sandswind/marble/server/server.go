package server

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"runtime"
	"sandswind/marble/log"
	"sandswind/marble/stats"
	"sandswind/marble/utils"
	"syscall"
	"time"
)

type Server struct {
	exitCh        chan bool     // 结束信号
	acceptTimeout time.Duration // 连接超时时间
	readTimeout   time.Duration // 读超时时间,其实也就是心跳维持时间
	writeTimeout  time.Duration // 写超时时间
	maxPkgLen     uint32        // 包长限制
	closed        bool          // 关闭服务标志
}

type asyncWork struct {
	conn   net.Conn
	header *Header
	body   *bytes.Buffer
}

var connMap *ConnMap = NewConnMap()

func NewServer() *Server {
	return &Server{
		exitCh:        make(chan bool),
		acceptTimeout: 60,
		readTimeout:   60,
		writeTimeout:  60,
		maxPkgLen:     10240,
		closed:        false,
	}
}

func (srv *Server) SetAcceptTimeout(acceptTimeout time.Duration) {
	srv.acceptTimeout = acceptTimeout
}

func (srv *Server) SetReadTimeout(readTimeout time.Duration) {
	srv.readTimeout = readTimeout
}

func (srv *Server) SetWriteTimeout(writeTimeout time.Duration) {
	srv.writeTimeout = writeTimeout
}

func (srv *Server) SetMaxPkgLen(maxPkgLen uint32) {
	srv.maxPkgLen = maxPkgLen
}

func (srv *Server) Start(listener *net.TCPListener) {
	log.Info("Start listen on:%v", listener.Addr())
	defer func() {
		listener.Close()
	}()

	// 防止恶意连接
	go srv.checkConnTimeout()

	for {
		select {
		case <-srv.exitCh:
			log.Warn("Stop listen on:%v", listener.Addr())
			return
		default:
		}

		listener.SetDeadline(time.Now().Add(srv.acceptTimeout))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if err == syscall.EINVAL {
				return
			}
			if e, ok := err.(*net.OpError); ok && !e.Temporary() {
				return
			}
			log.Warn("Accept error:%v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}
		stats.AddConn()

		// 连接后等待登陆验证
		connMap.Set(conn, time.Now())
		log.Info("Accept addr:%v", conn.RemoteAddr())

		go srv.handleClient(conn)
	}
}

// 处理连接超时
func (srv *Server) checkConnTimeout() {
	limitTime := 60 * time.Second
	ticker := time.NewTicker(limitTime)
	for _ = range ticker.C {
		items := connMap.Items()
		for conn, status := range items {
			if status != nil {
				deadline := status.(time.Time).Add(limitTime)
				if time.Now().After(deadline) {
					conn.(*net.TCPConn).Close()
					connMap.Delete(conn.(*net.TCPConn))
				}
			}
		}

		log.Info("当前连接数: %v", connMap.Size())
		log.Info("当前goroutine数: %v", runtime.NumGoroutine())
	}
}

func (srv *Server) Stop() {
	// 所有的exitCh都返回false
	if srv.closed {
		return
	}
	srv.closed = true
	close(srv.exitCh)
}

func (srv *Server) handleClient(conn *net.TCPConn) {
	var asyncWorkChan = make(chan *asyncWork, 1000)
	chStop := make(chan bool) // 通知停止消息处理
	addr := conn.RemoteAddr().String()

	defer func() {
		defer func() {
			if e := recover(); e != nil {
				log.Info("Panic:%v", e)
			}
		}()

		conn.Close()
		connMap.Delete(conn)
		log.Info("Disconnect:%v", addr)
		chStop <- true
	}()

	// 处理接收到的包
	go srv.handlePackets(conn, asyncWorkChan, chStop)

	// 接收数据
	log.Info("HandleClient:%v", addr)
	var hdata Header

	for {
		select {
		case <-srv.exitCh:
			log.Info("Server Stop Client")
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(srv.readTimeout))
		if err := binary.Read(conn, binary.LittleEndian, &hdata); err != nil {
			if err != io.EOF {
				log.Error("Read header:%v", err)
			}
			return
		}

		if hdata.Bodylen > srv.maxPkgLen {
			log.Warn("body too long:%v", hdata.Bodylen)
			return
		}

		body := utils.GetBuff()
		if hdata.Bodylen > 0 {
			body.Grow(int(hdata.Bodylen))
			b := body.Bytes()[0:hdata.Bodylen]
			if _, err := io.ReadFull(conn, b); err != nil {
				if err != io.EOF {
					log.Error("Read body:%v", err)
				}
				utils.PutBuff(body)
				return
			}
		}

		asyncWorkChan <- &asyncWork{
			conn:   conn,
			header: &hdata,
			body:   body,
		}
	}
}

func (srv *Server) handlePackets(conn *net.TCPConn, asyncWorkChan <-chan *asyncWork, chStop <-chan bool) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("Panic: %v", e)
		}
	}()

	for {
		select {
		case <-chStop:
			log.Warn("Stop handle receivePackets.")
			return

		// 消息包处理
		case w := <-asyncWorkChan:
			ModCall(w.conn, w.header, w.body.Bytes()[0:w.header.Bodylen])
			w.body.Reset()
			utils.PutBuff(w.body)
		}
	}
}
