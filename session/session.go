package session

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sumory/gotty/config"
	"github.com/sumory/gotty/packet"
	log "github.com/sumory/log4go"
	"io"
	"net"
	"syscall"
	"time"
)

//会话
type Session struct {
	conn         *net.TCPConn //tcp的session
	remoteAddr   string
	bReader      *bufio.Reader
	bWriter      *bufio.Writer
	ReadChannel  chan *packet.Packet //request的channel
	WriteChannel chan *packet.Packet //response的channel
	isClose      bool
	lastTime     time.Time
	config       *config.GottyConfig
	attrs        map[string]interface{}
}

func NewSession(conn *net.TCPConn, config *config.GottyConfig) *Session {

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(config.IdleTime * 2)
	conn.SetNoDelay(true)
	conn.SetReadBuffer(config.ReadBufSize)
	conn.SetWriteBuffer(config.WriteBufSize)

	session := &Session{
		conn:         conn,
		bReader:      bufio.NewReaderSize(conn, config.ReadBufSize),
		bWriter:      bufio.NewWriterSize(conn, config.WriteBufSize),
		ReadChannel:  make(chan *packet.Packet, config.ReadChanSize),
		WriteChannel: make(chan *packet.Packet, config.WriteChanSize),
		isClose:      false,
		remoteAddr:   conn.RemoteAddr().String(),
		config:       config}
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

	//缓存本次包的数据
	buff := make([]byte, 0, self.config.ReadBufSize)

	for !self.isClose {
		line, err := self.bReader.ReadSlice(packet.CMD_CRLF[1])
		//如果没有达到请求头的最小长度则继续读取
		if nil != err {
			buff = buff[:0]
			// buff.Reset()
			//链接是关闭的
			if err == io.EOF ||
				err == syscall.EPIPE ||
				err == syscall.ECONNRESET {
				self.Close()
				log.Error("session read packet failed, close session, remoteAddr: %s, err: %s", self.remoteAddr, err)
			}
			continue
		}

		l := len(buff) + len(line)
		//如果是\n那么就是一个完整的包
		if l >= packet.MAX_PACKET_BYTES {
			log.Error("session read packet too large, close session, remoteAddr: %s, err: %s", self.remoteAddr, err)
			self.Close()
			return
		} else {
			buff = append(buff, line...)
		}

		//complete packet
		if l > packet.PACKET_HEAD_LEN && buff[len(buff)-2] == packet.CMD_CRLF[0] {
			packet, err := packet.UnmarshalTLV(buff)
			if nil != err || nil == packet {
				log.Error("session read packet unmarshal failed, err: %s, buff length:%d, buff: %s", err, len(buff), buff)
				buff = buff[:0]
				continue
			}

			//写入缓冲
			self.ReadChannel <- packet
			//重置buffer
			buff = buff[:0]
		}
	}
}

//写入响应
func (self *Session) WritePacket() {
	var p *packet.Packet
	for !self.isClose {
		p = <-self.WriteChannel
		if nil != p {
			self.write0(p)
			self.lastTime = time.Now()
		} else {
			log.Warn("the packet to write is nil")
		}
	}
}

//真正写入网络的流
func (self *Session) write0(tlv *packet.Packet) {

	p := packet.MarshalPacket(tlv)
	if nil == p || len(p) <= 0 {
		log.Warn("packet after marshal to bytes is empty: %s", tlv)
		//如果是同步写出
		return
	}

	length, err := self.conn.Write(p)
	if nil != err {
		log.Error("session write0 error: remoteAddr %s, writeLength %d, fullLength %d, err %s", self.remoteAddr, length, len(p), err)
		//链接是关闭的
		if err == io.EOF || err == syscall.EPIPE || err == syscall.ECONNRESET {
			log.Info("to close session")
			self.Close()
			return
		}

		//如果没有写够则再写一次,是否能够写完？
		if err == io.ErrShortWrite {
			self.conn.Write(p[length:])
		}
	}
}

//写出数据
func (self *Session) Write(p *packet.Packet) error {
	defer func() {
		if err := recover(); nil != err {
			log.Error("session write revoer faild: %s, %s", self.remoteAddr, err)
		}
	}()

	if !self.isClose {
		select {
		case self.WriteChannel <- p:
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
