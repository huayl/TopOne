package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	lPort = flag.String("l", "6379", "listen port<port>")
	rHost = flag.String("h", "172.16.1.10", "remote addr<ip>")
	rPort = flag.String("p", "7826", "remote addr<port>")
)

func copy(wc io.WriteCloser, r io.Reader) {
	defer wc.Close()
	io.Copy(wc, r)
}

func handleConnection(us net.Conn, addr string) {
	if addr == "" {
		log.Printf("no addr for connection from %s", us.RemoteAddr())
		us.Close()
		return
	}

	ds, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("failed to dial %s: %s", addr, err)
		us.Close()
		return
	}

	// Ignore errors
	go copy(ds, us)
	go copy(us, ds)
}

func tcpBalance(bind string, addr string) error {
	log.Printf("using %v:%v\n", bind, addr)
	ln, err := net.Listen("tcp", bind)
	if err != nil {
		return fmt.Errorf("failed to bind: %s", err)
	}

	log.Printf("listening on %s", bind)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %s", err)
			continue
		}
		go handleConnection(conn, addr)
	}

	return err
}

func main() {
	flag.Parse()
	mPort := fmt.Sprintf(":%v", *lPort)
	mAddr := fmt.Sprintf("%v:%v", *rHost, *rPort)
	emsg := tcpBalance(mPort, mAddr)
	if emsg != nil {
		log.Printf("%v", emsg)
	}
}
