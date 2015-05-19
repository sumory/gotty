package client

import (
	"fmt"
	"github.com/sumory/gotty"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/session"
	log "github.com/sumory/log4go"
	"net"
)

type GottyClient struct {
	conn       *net.TCPConn
	codec      codec.Codec
	localAddr  string
	remoteAddr string
	heartbeat  int64
	session    *session.Session
	config     *config.GottyConfig
	context    *gotty.Context
	handler    func(client *GottyClient, d []byte) //包处理函数
}

func NewGottyClient(conn *net.TCPConn, //
	codec codec.Codec,
	config *config.GottyConfig, //
	context *gotty.Context, //
	handler func(client *GottyClient, d []byte), //
) *GottyClient {

	session := session.NewSession(conn, codec, config)

	client := &GottyClient{
		heartbeat: 0,
		conn:      conn,
		session:   session,
		config:    config,
		context:   context,
		handler:   handler,
	}

	return client
}

func (self *GottyClient) RemoteAddr() string {
	return self.remoteAddr
}

func (self *GottyClient) LocalAddr() string {
	return self.localAddr
}

func (self *GottyClient) Idle() bool {
	return self.session.Idle()
}

func (self *GottyClient) Start() {

	//重新初始化
	laddr := self.conn.LocalAddr().(*net.TCPAddr)
	raddr := self.conn.RemoteAddr().(*net.TCPAddr)
	self.localAddr = fmt.Sprintf("%s:%d", laddr.IP, laddr.Port)
	self.remoteAddr = fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	go self.session.WritePacket()
	go self.dispatchPacket()
	go self.session.ReadPacket()

	log.Info("client start: %s <-> %s", self.localAddr, self.remoteAddr)
}

//包分发
func (self *GottyClient) dispatchPacket() {
	//解析
	for nil != self.session && !self.session.Closed() {
		p := <-self.session.ReadChannel
		if nil == p {
			continue
		}

		//模拟queue/pool
		self.config.DispatcherQueueSize <- 1
		go func() {
			defer func() {
				<-self.config.DispatcherQueueSize
			}()


			fmt.Println("处理包",p)
			//self.do(p) //处理包
			self.handler(self, p)
		}()
	}
}

func (self *GottyClient) Write(d []byte) error {
	return self.session.Write(d)
}


//////// 网络维持 ////////////////////////////////
///

//重连
func (self *GottyClient) reconnect() (bool, error) {

	conn, err := net.DialTCP("tcp4", nil, self.conn.RemoteAddr().(*net.TCPAddr))
	if nil != err {
		log.Info("client reconnect failed, remoteAddr: %s, err: %s", self.RemoteAddr(), err)
		return false, err
	}

	//重置
	self.conn = conn
	self.session = session.NewSession(self.conn, self.codec, self.config)
	self.Start()
	return true, nil
}




func (self *GottyClient) updateHeartBeat(version int64) {
	if version > self.heartbeat {
		self.heartbeat = version
	}
}

func (self *GottyClient) Pong(opaque int32, version int64) {
	self.updateHeartBeat(version)
}

func (self *GottyClient) IsClosed() bool {
	return self.session.Closed()
}

func (self *GottyClient) Shutdown() {
	self.session.Close()
	log.Info("client shutdown: %s", self.remoteAddr)
}
