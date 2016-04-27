package codec

import (
	"github.com/sumory/gotty/buffer"
	"net"
)

//packet元信息
type PacketMeta struct {
	totalLen  uint32
	headerLen uint32
}

//packet的包头部分
type PacketHeader struct {
	Sequence  uint32 //请求的Sequence
	Operation uint16 //操作
	Version   uint16 //协议的版本号
	Extra     []byte //扩展数据
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
type Encoder interface {
	Encode(p *Packet) (interface{}, error)
}

type Decoder interface {
	Decode(m interface{}) (*Packet, error)
}

//编解码器接口
type Codec interface {
	Read(conn net.Conn, buffer *buffer.Buffer) error
	Write(conn net.Conn, buffer *buffer.Buffer, p []byte) error
	//编码, 实体 -> 数据包
	Marshal(m interface{}) (*Packet, error)
	//解码，数据包 -> 实体
	Unmarshal(p *Packet) (interface{}, error)
}
