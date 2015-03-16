package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sandswind/marble/config"
	"sandswind/marble/log"
	"sandswind/marble/server"
	"sandswind/marble/web"
	"syscall"
	"time"
)

var svr *server.Server

var (
	certFile = flag.String("tlscert", "", "TLS public certificate")
	keyFile  = flag.String("tlskey", "", "TLS private key")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s[OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "others")
}

func init() {
	// 日志初始化
	log.Init("")

	// 配置文件
	config.Init("./config.ini")

	// 初始化db

	// 设置cpu数量
	gomax := runtime.NumCPU() * 2
	runtime.GOMAXPROCS(gomax)
	rand.Seed(time.Now().UnixNano())
	//runtime.GOMAXPROCS(config.LocalConfig.CpuNum)

	// 服务器初始化
	svr = server.NewServer()
	svr.SetAcceptTimeout(time.Duration(config.LocalConfig.AccTimeout) * time.Second)
	svr.SetReadTimeout(time.Duration(config.LocalConfig.ReadTimeout) * time.Second)
	svr.SetWriteTimeout(time.Duration(config.LocalConfig.WriteTimeout) * time.Second)
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", config.LocalConfig.Tcp)
	checkError(err)

	tsock, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// Web服务器开始监听
	//go web.Start(config.LocalConfig.Http)
	go web.Start(config.LocalConfig.Http)

	// TCP服务器开始监听
	go svr.Start(tsock)

	// 处理中断信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Signal:%v", <-ch)

	// 结束服务
	svr.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Error("Error:%v", err)
		os.Exit(1)
	}
}
