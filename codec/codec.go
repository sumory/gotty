package codec

import (
	"bufio"
)

//具体业务实体需要实现此接口
type Encoder interface {
	Encode(p *Packet) (interface{}, error)
}

type Decoder interface {
	Decode(m interface{}) (*Packet, error)
}

//编解码器接口
type Codec interface {
	Read(bReader *bufio.Reader) (*Packet, error)
	Write(bWriter *bufio.Writer, p *Packet) error
	//编码, 实体 -> 数据包
	Marshal(m interface{}) (*Packet, error)
	//解码，数据包 -> 实体
	Unmarshal(p *Packet) (interface{}, error)
}
