package server

import (
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/session"
	log "github.com/sumory/log4go"
	"net"
	"time"
)

type GottyServer struct {
	addr       string
	keepalive  time.Duration
	stopChan   chan bool
	isShutdown bool
	config     *config.GottyConfig
	handler    func(session *session.Session, p codec.Packet) //包处理函数
	//编解码
	codec codec.Codec
}

func NewGottyServer( //
	addr string, // 服务地址
	keepalive time.Duration, // keepalive时间，秒
	config *config.GottyConfig, // 配置信息
	handler func(session *session.Session, p codec.Packet), //包处理函数
	codec codec.Codec, //编解码器
) *GottyServer {
	server := &GottyServer{
		addr:       addr,
		keepalive:  keepalive,
		stopChan:   make(chan bool, 1),
		isShutdown: false,
		config:     config,
		handler:    handler, //包处理函数
		codec:      codec,
	}
	return server
}

func (self *GottyServer) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.addr)
	if nil != err {
		log.Error("cannot use tcp addr: %s", self.addr)
		return err
	}

	listener, err := net.ListenTCP("tcp4", tcpAddr)
	if nil != err {
		log.Info("server listen failed: %s", self.addr)
		return err
	}

	stopListener := &StoppedListener{listener, self.stopChan, self.keepalive}
	go self.serve(stopListener)

	return nil
}

func (self *GottyServer) serve(listener *StoppedListener) error {
	for !self.isShutdown {
		conn, err := listener.Accept()
		if nil != err {
			log.Info("listener accept failed: %s", err)
			continue
		} else {
			log.Info("listner accept new connection, server <--- %s", conn.RemoteAddr())
			// gottyClient := client.NewGottyClient(conn, self.codec, self.config, self.handler)
			// gottyClient.Start()

			s := session.NewSession(conn, self.codec, self.config, self.handler)
			s.Start()
		}
	}
	return nil
}

func (self *GottyServer) Shutdown() {
	self.isShutdown = true
	close(self.stopChan)
	log.Info("server shutdown")
}
