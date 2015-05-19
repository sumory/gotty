package session

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sumory/gotty/buffer"
	"github.com/sumory/gotty/codec"
	"github.com/sumory/gotty/config"
	log "github.com/sumory/log4go"
	"io"
	"net"
	"syscall"
	"time"
)

//会话
type Session struct {
	conn       *net.TCPConn //tcp的session
	remoteAddr string

	//消息传输
	bReader      *bufio.Reader
	bWriter      *bufio.Writer
	ReadChannel  chan []byte //request的channel
	WriteChannel chan []byte //response的channel
	inBuffer     *buffer.Buffer
	outBuffer    *buffer.Buffer

	isClose  bool
	lastTime time.Time
	config   *config.GottyConfig
	attrs    map[string]interface{}

	//编解码
	codec codec.Codec
}

func NewSession(conn *net.TCPConn, codec codec.Codec, config *config.GottyConfig) *Session {

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(config.IdleTime * 2)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(config.ReadBufSize)
	conn.SetWriteBuffer(config.WriteBufSize)

	session := &Session{
		conn:       conn,
		remoteAddr: conn.RemoteAddr().String(),

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

func (self *Session) SetAttribute(name string, v interface{}) {
	self.attrs[name] = v
}

func (self *Session) GetAttribute(name string) interface{} {
	return self.attrs[name]
}

func (self *Session) RemotingAddr() string {
	return self.remoteAddr
}

//空闲逻辑
func (self *Session) Idle() bool {
	return time.Now().After(self.lastTime.Add(self.config.IdleTime))
}

//读取
func (self *Session) ReadPacket() {
	defer func() {
		if err := recover(); nil != err {
			log.Info("session read packet panic recover failed, remoteAddr: %s, err: %s", self.remoteAddr, err)
		}
	}()

	//编解码接口调用
	e:=self.codec.ReadPacket(self.conn, self.inBuffer)
	if e!=nil{
		log.Error("read packet error, ",e)
		self.Close()
	}

	tmpData:=make([]byte, self.inBuffer.Length())

	copy(self.inBuffer.Data[:],tmpData)
	//写入缓冲
	self.ReadChannel <- tmpData
}

//写入响应
func (self *Session) WritePacket() {
	var p []byte
	for !self.isClose {
		p = <-self.WriteChannel
		if nil != p {

log.Info("写出报。。")
			e:=self.codec.WritePacket(self.conn,self.outBuffer, p)
			if e!=nil {
				log.Error("写出包错误", e)
			}

			//self.write0(p)
			self.lastTime = time.Now()
		} else {
			log.Warn("the packet to write is nil")
		}
	}
}

//真正写入网络的流
func (self *Session) write0(d []byte) {
	length, err := self.conn.Write(d)
	if nil != err {
		log.Error("session write0 error: remoteAddr %s, writeLength %d, fullLength %d, err %s", self.remoteAddr, length, len(d), err)
		//链接是关闭的
		if err == io.EOF || err == syscall.EPIPE || err == syscall.ECONNRESET {
			log.Info("to close session")
			self.Close()
			return
		}

		//如果没有写够则再写一次,是否能够写完？
		if err == io.ErrShortWrite {
			self.conn.Write(d[length:])
		}
	}
}

//写出数据
func (self *Session) Write(d []byte) error {
	defer func() {
		if err := recover(); nil != err {
			log.Error("session write revoer faild: %s, %s", self.remoteAddr, err)
		}
	}()

	if !self.isClose {
		select {
		case self.WriteChannel <- d:
			return nil
		default:
			return errors.New(fmt.Sprintf("write channel is full: %s", self.remoteAddr))
		}
	}
	return errors.New(fmt.Sprintf("session closed: %s", self.remoteAddr))
}

//当前连接是否关闭
func (self *Session) Closed() bool {
	return self.isClose
}

func (self *Session) Close() error {

	if !self.isClose {
		self.isClose = true
		self.conn.Close()
		close(self.WriteChannel)
		close(self.ReadChannel)
		log.Info("session close, remoteAddr: %s", self.remoteAddr)
	}
	return nil
}
