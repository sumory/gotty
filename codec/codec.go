package codec

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/sumory/gotty/buffer"
	"net"
)

// Errors
var (
	SendToClosedError     = errors.New("Send to closed session")
	BlockingError         = errors.New("Blocking happened")
	PacketTooLargeError   = errors.New("Packet too large")
	NilBufferError        = errors.New("Buffer is nil")
	HeaderLargerThanTotal = errors.New("Header is larger than total packet")
)

//协议接口
type Protocol interface {
	BufferFactory() BufferFactory
	NewCodec() Codec
}

type BufferFactory interface {
	NewInBuffer() buffer.Buffer
	NewOutBuffer() buffer.Buffer
}

//编解码器接口
type Codec interface {
	GetNbit() int
	ReadPacket(conn net.Conn, buffer *buffer.Buffer) error
	WritePacket(conn net.Conn, buffer *buffer.Buffer, p []byte) error
	Encode(obj *interface{}) []byte
	Decode(data []byte) *interface{}
}

//实体消息接口
type Message interface {
	BufferSize() int
	WriteBuffer(v []byte, bf *buffer.Buffer) error
	ReadBuffer(bf *buffer.Buffer) interface{}
}

type BytesMessage struct {
}

func (e BytesMessage) BufferSize() int {
	return 1024
}

func (e BytesMessage) WriteBuffer(v []byte, bf *buffer.Buffer) error {
	bf.WriteBytes(v)
	return nil
}

func (e BytesMessage) ReadBuffer(n int, bf *buffer.Buffer) interface{} {
	v := bf.ReadBytes(n)
	return v
}

type JsonMessage struct {
}

func (e JsonMessage) BufferSize() int {
	return 1024
}

func (e JsonMessage) WriteBuffer(v []byte, bf *buffer.Buffer) error {
	return json.NewEncoder(bf).Encode(v)
}

func (e JsonMessage) ReadBuffer(n int, bf *buffer.Buffer) interface{} {
	v := bf.ReadBytes(n)
	return v
}

type GobMessage struct {
}

func (e GobMessage) BufferSize() int {
	return 1024
}

func (e GobMessage) WriteBuffer(v []byte, bf *buffer.Buffer) error {
	return gob.NewEncoder(bf).Encode(v)
}

func (e GobMessage) ReadBuffer(n int, bf *buffer.Buffer) interface{} {
	v := bf.ReadBytes(n)
	return v
}

type XmlMessage struct {
}

func (e XmlMessage) BufferSize() int {
	return 1024
}

func (e XmlMessage) WriteBuffer(v []byte, bf *buffer.Buffer) error {
	return xml.NewEncoder(bf).Encode(v)
}

func (e XmlMessage) ReadBuffer(n int, bf *buffer.Buffer) interface{} {
	v := bf.ReadBytes(n)
	return v
}
