package codec

import (
	"github.com/sumory/gotty/buffer"
	"net"
)

//pacakge元信息
type PacketMeta struct {
	nBit      uint8
	totalLen  []byte
	headerLen []byte
}

//packet的包头部分
type PacketHeader struct {
	Sequence  int32  //请求的Sequence
	Operation uint16 //操作
	Version   int16  //协议的版本号
}

//packet的包体部分
type PacketBody struct {
	data []byte
}

//请求的packet
type Packet struct {
	Meta   *PacketMeta
	Header *PacketHeader
	Body   *PacketBody
}

//具体业务实体需要实现此接口
type Message interface {
	Encode(p *Packet) interface{}
	Decode(m interface{}) *Packet
}

//编解码器接口
type Codec interface {
	Read(conn net.Conn, buffer *buffer.Buffer) error
	Write(conn net.Conn, buffer *buffer.Buffer, p []byte) error
	//编码, 包 -> 字节
	Marshal(m Message) *Packet
	//解码，字节 -> 包
	Unmarshal(p *Packet) (Message, error)
}
