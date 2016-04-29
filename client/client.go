package client

import (
	"fmt"
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
	handler    func(session *session.Session, p *codec.Packet) //包处理函数
}

func NewGottyClient(conn *net.TCPConn, //
	codec codec.Codec,
	config *config.GottyConfig, //
	handler func(session *session.Session, p *codec.Packet), //
) *GottyClient {

	session := session.NewSession(conn, codec, config, handler)

	client := &GottyClient{
		heartbeat: 0,
		conn:      conn,
		session:   session,
		config:    config,
		handler:   handler,
	}

	return client
}

func (client *GottyClient) RemoteAddr() string {
	return client.remoteAddr
}

func (client *GottyClient) LocalAddr() string {
	return client.localAddr
}

func (client *GottyClient) Idle() bool {
	return client.session.Idle()
}

func (client *GottyClient) Start() {

	//重新初始化
	laddr := client.conn.LocalAddr().(*net.TCPAddr)
	raddr := client.conn.RemoteAddr().(*net.TCPAddr)
	client.localAddr = fmt.Sprintf("%s:%d", laddr.IP, laddr.Port)
	client.remoteAddr = fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	go client.session.WritePacket()
	go client.dispatchPacket()
	go client.session.ReadPacket()

	log.Info("client start: %s <-> %s", client.localAddr, client.remoteAddr)
}

//dispatchPacket 包分发
func (client *GottyClient) dispatchPacket() {
	//解析
	for nil != client.session && !client.session.Closed() {
		p := <-client.session.ReadChannel
		if nil == p {
			continue
		}

		//模拟queue/pool
		client.config.DispatcherQueueSize <- 1
		go func() {
			defer func() {
				<-client.config.DispatcherQueueSize
			}()

			client.handler(client.session, p)
		}()
	}
}

func (client *GottyClient) Write(p *codec.Packet) error {
	return client.session.Write(p)
}

func (client *GottyClient) reconnect() (bool, error) {
	conn, err := net.DialTCP("tcp4", nil, client.conn.RemoteAddr().(*net.TCPAddr))
	if nil != err {
		log.Info("client reconnect failed, remoteAddr: %s, err: %s", client.RemoteAddr(), err)
		return false, err
	}

	//重置
	client.conn = conn
	client.session = session.NewSession(client.conn, client.codec, client.config, client.handler)
	client.Start()
	return true, nil
}

func (client *GottyClient) IsClosed() bool {
	return client.session.Closed()
}

//Shutdown 关闭客户端
func (client *GottyClient) Shutdown() {
	client.session.Close()
	log.Debug("client shutdown: %s", client.remoteAddr)
}
