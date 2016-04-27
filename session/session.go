package session

import (
	"bufio"
	"fmt"
	"github.com/sumory/gotty/buffer"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	log "github.com/sumory/log4go"
	"net"
	"sync/atomic"
	"time"
)

//GlobalSessionID session标识
var GlobalSessionID uint64

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
	ReadChannel  chan []byte //传输请求体的channel
	WriteChannel chan []byte //传输响应体的channel
	inBuffer     *buffer.Buffer
	outBuffer    *buffer.Buffer

	isClose  bool
	lastTime time.Time              //最后活跃时间
	attrs    map[string]interface{} //其他属性数据

	codec codec.Codec //编解码器
}

//NewSession 创建新的session对话
func NewSession(conn *net.TCPConn, codec codec.Codec, config *config.GottyConfig) *Session {
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
		ReadChannel:  make(chan []byte, config.ReadChanSize),
		WriteChannel: make(chan []byte, config.WriteChanSize),
		inBuffer:     buffer.NewBuffer(0, 1024),
		outBuffer:    buffer.NewBuffer(0, 1024),

		isClose: false,
		config:  config,

		codec: codec,
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
		e := session.codec.Read(session.conn, session.inBuffer)
		if e != nil {
			log.Error("read packet error, ", e)
			session.Close()
		}

		tmpData := make([]byte, session.inBuffer.Length())
		copy(tmpData, session.inBuffer.Data)
		session.ReadChannel <- tmpData
	}
}

//WritePacket 从channel中取出包并写出
func (session *Session) WritePacket() {
	var p []byte
	for !session.isClose {
		p = <-session.WriteChannel
		if nil != p {
			log.Debug("WritePacket写出包, %s", p)
			e := session.codec.Write(session.conn, session.outBuffer, p)
			if e != nil {
				log.Error("写出包错误", e)
			}

			session.lastTime = time.Now()
		} else {
			log.Warn("the packet from WriteChannel is nil")
		}
	}
}

//写出数据
func (session *Session) Write(d []byte) error {
	defer func() {
		if err := recover(); nil != err {
			log.Warn("session write packet failed, localAddr: %s, remoteAddr: %s, err: %s",
				session.localAddr, session.remoteAddr, err)
		}
	}()

	if !session.isClose {
		select {
		case session.WriteChannel <- d:
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
