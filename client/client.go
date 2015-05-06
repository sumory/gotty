package client

import (
	"errors"
	"fmt"
	"github.com/sumory/gotty"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/packet"
	"github.com/sumory/gotty/session"
	log "github.com/sumory/log4go"
	"net"
	"time"
)

type GottyClient struct {
	conn             *net.TCPConn
	localAddr        string
	remoteAddr       string
	heartbeat        int64
	session          *session.Session
	config           *config.GottyConfig
	context          *gotty.Context
	packetDispatcher func(client *GottyClient, p *packet.Packet) //包处理函数
}

func NewGottyClient(conn *net.TCPConn, //
	config *config.GottyConfig, //
	context *gotty.Context, //
	packetDispatcher func(client *GottyClient, p *packet.Packet), //
) *GottyClient {

	session := session.NewSession(conn, config)

	client := &GottyClient{
		heartbeat:        0,
		conn:             conn,
		session:          session,
		config:           config,
		context:          context,
		packetDispatcher: packetDispatcher}

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

			//self.do(p) //处理包
			self.packetDispatcher(self, p)
		}()
	}
}

//处理包发出
func (self *GottyClient) do(p *packet.Packet) {
	resp := packet.NewRespPacket(p.Opaque, p.CmdType, p.Data)
	self.Write(*resp)
}

func (self *GottyClient) Write(p packet.Packet) (chan interface{}, error) {

	pp := &p
	opaque, future := self.fillOpaque(pp)
	self.context.ReqHolder.Attach(opaque, future)
	return future, self.session.Write(pp)
}

//attach到当前的等待回调chan
func (self *GottyClient) Attach(opaque int32, obj interface{}) {
	defer func() {
		if err := recover(); nil != err {
			log.Error("attach failed %s %s", err, obj)
		}
	}()

	self.context.ReqHolder.Detach(opaque, obj)

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
	self.session = session.NewSession(self.conn, self.config)
	self.Start()
	return true, nil
}

func (self *GottyClient) Ping(heartbeat *packet.Packet, timeout time.Duration) error {
	pong, err := self.WriteAndGet(*heartbeat, timeout)
	if nil != err {
		return err
	}
	version, ok := pong.(int64)
	if !ok {
		log.Error("ping pong error type: %s", pong)
		return errors.New("ERROR PONG TYPE !")
	}
	self.updateHeartBeat(version)
	return nil
}

func (self *GottyClient) WriteAndGet(p packet.Packet, timeout time.Duration) (interface{}, error) {

	pp := &p
	opaque, future := self.fillOpaque(pp)
	self.context.ReqHolder.Attach(opaque, future)
	err := self.session.Write(pp)

	if nil != err {
		return nil, err
	}

	tid, ch := self.context.TimeWheel.After(timeout, func() {
		log.Warn("timeout!!!!")
	})

	var resp interface{}
	select {
	case <-ch:
		// 	//删除掉当前holder
		return nil, errors.New("WAIT RESPONSE TIMEOUT")
	case resp = <-future:
		self.context.TimeWheel.Remove(tid)
		return resp, nil
	}

}

func (self *GottyClient) fillOpaque(p *packet.Packet) (int32, chan interface{}) {
	tid := p.Opaque
	//只有在默认值没有赋值的时候才去赋值
	if tid < 0 {
		id := self.context.ReqHolder.CurrentOpaque()
		p.Opaque = id
		tid = id
	}

	return tid, make(chan interface{}, 1)
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
