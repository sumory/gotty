package codec

import (
	"encoding/binary"
	"github.com/sumory/gotty/buffer"
	log "github.com/sumory/log4go"
)

const (
	packetBytesLen     = 4 //package包的长度字段字节长度,对应uint32类型
	packetMetaLen      = packetBytesLen * 2
	packetMinHeaderLen = 4 + 2 + 2
)

//packet元信息
type PacketMeta struct {
	TotalLen  uint32
	HeaderLen uint32
}

func (packetMeta *PacketMeta) Len() int {
	return 8
}

//packet的包头部分
type PacketHeader struct {
	Sequence  uint32 //请求的Sequence
	Operation uint16 //操作
	Version   uint16 //协议的版本号
	Extra     []byte //扩展数据
}

func (packetHeader *PacketHeader) Len() int {
	l := 4 + 2 + 2
	if packetHeader.Extra != nil {
		l += len(packetHeader.Extra)
	}
	return l
}

//packet的包体部分
type PacketBody struct {
	Data []byte
}

func (packetBody *PacketBody) Len() int {
	l := 0
	if packetBody.Data != nil {
		l += len(packetBody.Data)
	}
	return l
}

//请求的packet
type Packet struct {
	Meta   *PacketMeta
	Header *PacketHeader
	Body   *PacketBody
}

//NewPacket 新建packet
func NewPacket(totalLen, headerLen uint32, sequence uint32, operation, version uint16, extra, data []byte) *Packet {
	meta := &PacketMeta{
		TotalLen:  totalLen,
		HeaderLen: headerLen,
	}
	header := &PacketHeader{
		Sequence:  sequence,
		Operation: operation,
		Version:   version,
		Extra:     extra,
	}
	body := &PacketBody{
		Data: data,
	}
	return &Packet{
		Meta:   meta,
		Header: header,
		Body:   body,
	}
}

func (packet *Packet) Decode(bo binary.ByteOrder, totalLen, headerLen uint32, headerAndBody []byte) error {
	meta := &PacketMeta{
		TotalLen:  totalLen,
		HeaderLen: headerLen,
	}
	header := &PacketHeader{
		Extra: make([]byte, headerLen-packetMinHeaderLen),
	}
	header.Sequence = bo.Uint32(headerAndBody[0:4])
	header.Operation = bo.Uint16(headerAndBody[4:6])
	header.Version = bo.Uint16(headerAndBody[6:8])
	if headerLen > packetMinHeaderLen { //存在附加属性
		copy(header.Extra[:], headerAndBody[8:])
	}
	body := &PacketBody{
		Data: make([]byte, totalLen-headerLen-packetMetaLen),
	}
	if totalLen-packetMetaLen > headerLen { //存在包体
		copy(body.Data[:], headerAndBody[headerLen:])
	}
	packet.Meta = meta
	packet.Header = header
	packet.Body = body

	return nil
}

func (packet *Packet) Encode(bo binary.ByteOrder) ([]byte, error) {
	tLen := packet.Meta.TotalLen
	bf := buffer.NewBuffer(0, int(tLen))

	if bo == binary.BigEndian {
		bf.WriteUint32BE(tLen)
		bf.WriteUint32BE(packet.Meta.HeaderLen)
		bf.WriteUint32BE(packet.Header.Sequence)
		bf.WriteUint16BE(packet.Header.Operation)
		bf.WriteUint16BE(packet.Header.Version)
	}
	if bo == binary.LittleEndian {
		bf.WriteUint32LE(tLen)
		bf.WriteUint32LE(packet.Meta.HeaderLen)
		bf.WriteUint32LE(packet.Header.Sequence)
		bf.WriteUint16LE(packet.Header.Operation)
		bf.WriteUint16LE(packet.Header.Version)
	}

	log.Debug("packet.Encode, packet.len: %d  Header.Extra.len: %d  Body.Data.len: %d",
		packet.Meta.TotalLen, len(packet.Header.Extra), len(packet.Body.Data))
	//写header extra
	bf.Write(packet.Header.Extra)
	//写body
	bf.Write(packet.Body.Data)

	return bf.Data[:], nil
}
