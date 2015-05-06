package server

import (
	"github.com/sumory/gotty"
	"github.com/sumory/gotty/client"
	"github.com/sumory/gotty/utils"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/packet"
	log "github.com/sumory/log4go"
	"net"
	"time"
)

type GottyServer struct {
	addr             string
	keepalive        time.Duration
	stopChan         chan bool
	isShutdown       bool
	config           *config.GottyConfig
	context          *gotty.Context
	packetDispatcher func(client *client.GottyClient, p *packet.Packet) //包处理函数
}

func NewGottyServer( //
	addr string, // 服务地址
	keepaliveTime int, // keepalive时间，秒
	config *config.GottyConfig, // 配置信息
	maxOpaque int, // 最大id标识
	concurrent int, //缓冲器的并发因子
	packetDispatcher func(client *client.GottyClient, p *packet.Packet), //包处理函数
) *GottyServer {
	reqHolder := gotty.NewReqHolder(concurrent, maxOpaque)
	timeWheel := utils.NewTimeWheel(1*time.Second, 6, 10)
	context := gotty.NewContext(reqHolder, timeWheel)

	server := &GottyServer{
		addr:             addr,
		keepalive:        5 * time.Second,
		stopChan:         make(chan bool, 1),
		isShutdown:       false,
		config:           config,
		context:          context,
		packetDispatcher: packetDispatcher, //包处理函数
	}
	return server
}

func (self *GottyServer) ListenAndServe() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", self.addr)
	if nil != err {
		log.Info("cannot user tcp addr: %s", self.addr)
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
			log.Info("listner accept new connection: %s", conn.RemoteAddr())
			gottyClient := client.NewGottyClient(conn, self.config, self.context, self.packetDispatcher)
			gottyClient.Start()
		}
	}
	return nil
}

func (self *GottyServer) Shutdown() {
	self.isShutdown = true
	close(self.stopChan)
	log.Info("server shutdown")
}
