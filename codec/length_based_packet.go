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
type LengthBasedPacketMeta struct {
	TotalLen  uint32
	HeaderLen uint32
}

func (packetMeta *LengthBasedPacketMeta) Len() int {
	return 8
}

//packet的包头部分
type LengthBasedPacketHeader struct {
	Sequence  uint32 //请求的Sequence
	Operation uint16 //操作
	Version   uint16 //协议的版本号
	Extra     []byte //扩展数据
}

func (packetHeader *LengthBasedPacketHeader) Len() int {
	l := 4 + 2 + 2
	if packetHeader.Extra != nil {
		l += len(packetHeader.Extra)
	}
	return l
}

//packet的包体部分
type LengthBasedPacketBody struct {
	Data []byte
}

func (packetBody *LengthBasedPacketBody) Len() int {
	l := 0
	if packetBody.Data != nil {
		l += len(packetBody.Data)
	}
	return l
}

//请求的packet
type LengthBasedPacket struct {
	Meta   *LengthBasedPacketMeta
	Header *LengthBasedPacketHeader
	Body   *LengthBasedPacketBody
}

//NewPacket 新建packet
func NewLengthBasedPacket(totalLen, headerLen uint32, sequence uint32, operation, version uint16, extra, data []byte) *LengthBasedPacket {
	meta := &LengthBasedPacketMeta{
		TotalLen:  totalLen,
		HeaderLen: headerLen,
	}
	header := &LengthBasedPacketHeader{
		Sequence:  sequence,
		Operation: operation,
		Version:   version,
		Extra:     extra,
	}
	body := &LengthBasedPacketBody{
		Data: data,
	}
	return &LengthBasedPacket{
		Meta:   meta,
		Header: header,
		Body:   body,
	}
}

//NewPacketFromBinary 从字节数组构建packet
func NewLengthBasedPacketFromBinary(bo binary.ByteOrder, data []byte) *LengthBasedPacket {
	if data == nil || len(data) == 0 {
		return nil
	}

	var p *LengthBasedPacket = &LengthBasedPacket{}
	meta := &LengthBasedPacketMeta{
		TotalLen:  bo.Uint32(data[0:4]),
		HeaderLen: bo.Uint32(data[4:8]),
	}
	p.Meta = meta

	if meta.HeaderLen > 0 {
		header := &LengthBasedPacketHeader{
			Sequence:  bo.Uint32(data[8:12]),
			Operation: bo.Uint16(data[12:14]),
			Version:   bo.Uint16(data[14:16]),
			Extra:     nil,
		}

		if meta.HeaderLen > packetMinHeaderLen {
			header.Extra = make([]byte, meta.HeaderLen-packetMinHeaderLen)
			copy(header.Extra[:], data[packetMinHeaderLen+packetMetaLen:])
		}
		p.Header = header
	}

	body := &LengthBasedPacketBody{
		Data: make([]byte, meta.TotalLen-packetMetaLen-meta.HeaderLen),
	}
	copy(body.Data[:], data[packetMetaLen+meta.HeaderLen:])
	p.Body = body
	return p
}

func (packet LengthBasedPacket) Decode(bo binary.ByteOrder, totalLen, headerLen uint32, headerAndBody []byte) error {
	meta := &LengthBasedPacketMeta{
		TotalLen:  totalLen,
		HeaderLen: headerLen,
	}
	header := &LengthBasedPacketHeader{
		Extra: make([]byte, headerLen-packetMinHeaderLen),
	}
	header.Sequence = bo.Uint32(headerAndBody[0:4])
	header.Operation = bo.Uint16(headerAndBody[4:6])
	header.Version = bo.Uint16(headerAndBody[6:8])
	if headerLen > packetMinHeaderLen { //存在附加属性
		copy(header.Extra[:], headerAndBody[8:])
	}
	body := &LengthBasedPacketBody{
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

func (packet LengthBasedPacket) Encode(bo binary.ByteOrder) ([]byte, error) {
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

func (packet LengthBasedPacket) Transform(m Message) error {
	return m.FromPacket(packet)
}
