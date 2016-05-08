package session

import (
	"bufio"
	"fmt"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	log "github.com/sumory/log4go"
	"net"
	"sync/atomic"
	"time"
)

//GlobalSessionID session标识
var GlobalSessionID uint64

type handlerFunc func(session *Session, p codec.Packet)

//Session 服务端与客户端间对话，对应一条物理连接
type Session struct {
	config *config.GottyConfig //配置

	id         uint64 //id标识
	conn       *net.TCPConn
	remoteAddr string //远程地址
	localAddr  string //本地地址

	//消息传输
	bReader      *bufio.Reader
	bWriter      *bufio.Writer
	ReadChannel  chan codec.Packet //传输请求体的channel
	WriteChannel chan codec.Packet //传输响应体的channel

	isClose  bool
	lastTime time.Time              //最后活跃时间
	attrs    map[string]interface{} //其他属性数据

	codec   codec.Codec //编解码器
	handler handlerFunc //包处理函数
}

//NewSession 创建新的session对话
func NewSession(conn *net.TCPConn, sessionCodec codec.Codec, config *config.GottyConfig, handler handlerFunc) *Session {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(config.IdleTime * 2)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(config.ReadBufSize)
	conn.SetWriteBuffer(config.WriteBufSize)

	session := &Session{
		id:         atomic.AddUint64(&GlobalSessionID, 1),
		conn:       conn,
		remoteAddr: conn.RemoteAddr().String(),
		localAddr:  conn.LocalAddr().String(),

		bReader:      bufio.NewReaderSize(conn, config.ReadBufSize),
		bWriter:      bufio.NewWriterSize(conn, config.WriteBufSize),
		ReadChannel:  make(chan codec.Packet, config.ReadChanSize),
		WriteChannel: make(chan codec.Packet, config.WriteChanSize),

		isClose: false,
		config:  config,

		codec:   sessionCodec,
		handler: handler,
	}
	return session
}

//Set 保存自定义的kv数据
func (session *Session) Set(name string, v interface{}) {
	session.attrs[name] = v
}

//Get 获取自定义的kv数据
func (session *Session) Get(name string) interface{} {
	return session.attrs[name]
}

//RemoteAddr 获取连接的远程地址
func (session *Session) RemoteAddr() string {
	return session.remoteAddr
}

//LocalAddr 获取连接的本地地址
func (session *Session) LocalAddr() string {
	return session.localAddr
}

//Idle 是否空闲
func (session *Session) Idle() bool {
	return time.Now().After(session.lastTime.Add(session.config.IdleTime))
}

//ReadPacket 读取
func (session *Session) ReadPacket() {
	defer func() {
		if err := recover(); nil != err {
			log.Warn("session read packet failed, localAddr: %s, remoteAddr: %s, err: %s",
				session.localAddr, session.remoteAddr, err)
		}
	}()
	for !session.isClose {
		packet, err := session.codec.Read(session.bReader)
		if err != nil {
			log.Error("read packet error, ", err)
			session.Close()
		}

		session.ReadChannel <- packet
	}
}

//WritePacket 从channel中取出包并写出
func (session *Session) WritePacket() {
	var p codec.Packet
	for !session.isClose {
		p = <-session.WriteChannel
		if nil != p {
			log.Debug("WritePacket写出包, Header.Extra:%v  Body:%v", p.Header.Extra, p.Body.Data)
			err := session.codec.Write(session.bWriter, p)
			if err != nil {
				log.Error("写出包错误", err)
			}

			session.lastTime = time.Now()
		} else {
			log.Warn("the packet from WriteChannel is nil")
		}
	}
}

//dispatchPacket 包分发
func (session *Session) dispatchPacket() {
	//解析
	for !session.Closed() {
		p := <-session.ReadChannel
		if nil == p {
			continue
		}

		//模拟queue/pool
		session.config.DispatcherQueueSize <- 1
		go func() {
			defer func() {
				<-session.config.DispatcherQueueSize
			}()

			session.handler(session, p)
		}()
	}
}

//ReadMessage 读取
func (session *Session) ReadMessage() {
	defer func() {
		if err := recover(); nil != err {
			log.Warn("session read packet failed, localAddr: %s, remoteAddr: %s, err: %s",
				session.localAddr, session.remoteAddr, err)
		}
	}()
	for !session.isClose {
		packet, err := session.codec.Read(session.bReader)
		if err != nil {
			log.Error("read packet error, ", err)
			session.Close()
		}

		session.ReadChannel <- packet
	}
}

//WriteMessage 从channel中取出包并写出
func (session *Session) WriteMessage() {
	var p codec.Packet
	for !session.isClose {
		p = <-session.WriteChannel
		if nil != p {
			log.Debug("WritePacket写出包, Header.Extra:%v  Body:%v", p.Header.Extra, p.Body.Data)
			err := session.codec.Write(session.bWriter, p)
			if err != nil {
				log.Error("写出包错误", err)
			}

			session.lastTime = time.Now()
		} else {
			log.Warn("the packet from WriteChannel is nil")
		}
	}
}

//dispatchMessage 包分发
func (session *Session) dispatchMessage() {
	//解析
	for !session.Closed() {
		p := <-session.ReadChannel
		if nil == p {
			continue
		}

		//模拟queue/pool
		session.config.DispatcherQueueSize <- 1
		go func() {
			defer func() {
				<-session.config.DispatcherQueueSize
			}()

			session.handler(session, p)
		}()
	}
}

//Start 开启session，开始收发包
func (session *Session) Start() {
	go session.WritePacket()
	go session.dispatchPacket()
	go session.ReadPacket()

	laddr := session.conn.LocalAddr().(*net.TCPAddr)
	raddr := session.conn.RemoteAddr().(*net.TCPAddr)
	session.localAddr = fmt.Sprintf("%s:%d", laddr.IP, laddr.Port)
	session.remoteAddr = fmt.Sprintf("%s:%d", raddr.IP, raddr.Port)

	log.Info("session start: %s <-> %s", session.localAddr, session.remoteAddr)
}

//写出数据
func (session *Session) Write(p codec.Packet) error {
	defer func() {
		if err := recover(); nil != err {
			log.Warn("session write packet failed, localAddr: %s, remoteAddr: %s, err: %s",
				session.localAddr, session.remoteAddr, err)
		}
	}()

	if !session.isClose {
		log.Debug("client write packet: %+v", p.Body.Data)

		select {
		case session.WriteChannel <- p:
			return nil
		default:
			return fmt.Errorf("write channel is full: %s", session.remoteAddr)
		}
	}
	return fmt.Errorf("session closed: %s", session.remoteAddr)
}

//Closed 当前连接是否关闭
func (session *Session) Closed() bool {
	return session.isClose
}

//Close 关闭当前对话：关闭连接、channel及其他善后处理
func (session *Session) Close() error {
	if !session.isClose {
		session.isClose = true
		session.conn.Close()
		close(session.WriteChannel)
		close(session.ReadChannel)
		log.Info("session close, remoteAddr: %s", session.remoteAddr)
	}
	return nil
}
