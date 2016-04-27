package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/sumory/gotty/buffer"
	log "github.com/sumory/log4go"
	"io"
	"net"
)

//LengthBasedCodec 定长编解码器
type LengthBasedCodec struct {
	byteOrder binary.ByteOrder //大小端
	totalLen  []byte           //包第一部分,一般指包总大小
	headerLen []byte           //包第二部分,一般指包头大小
	maxSize   int              //包最大长度
	encoder   Encoder
	decoder   Decoder
}

//NewLengthBasedCodec 新建定长编解码器
func NewLengthBasedCodec(byteOrder binary.ByteOrder) *LengthBasedCodec {
	return &LengthBasedCodec{
		byteOrder: byteOrder,
		totalLen:  make([]byte, 4),
		headerLen: make([]byte, 4),
		maxSize:   32 * 1024, //32k
		encoder:   nil,
		decoder:   nil,
	}
}

//ReadPacket 从连接中读取报文放入buffer中
func (lbc *LengthBasedCodec) Read(conn net.Conn, bf *buffer.Buffer) error {
	//读包总大小
	if _, err := io.ReadFull(conn, lbc.totalLen); err != nil {
		return err
	}

	tLen := int(lbc.byteOrder.Uint32(lbc.totalLen)) //包总长度
	if lbc.maxSize > 0 && tLen > lbc.maxSize {
		return PacketTooLargeError
	}
	if tLen == 0 {
		return nil
	}

	//读包头大小
	if _, err := io.ReadFull(conn, lbc.headerLen); err != nil {
		return err
	}
	hLen := int(lbc.byteOrder.Uint32(lbc.headerLen))

	if hLen > tLen {
		return HeaderLargerThanTotal
	}

	log.Debug("totallen: %d, headerLen: %d", tLen, hLen)
	bf.Reset(0, tLen)
	bf.Write(lbc.totalLen)  //将包总长度写入buffer
	bf.Write(lbc.headerLen) //将包头长度写入buffer

	tmp := make([]byte, tLen-8)
	if _, err := io.ReadFull(conn, tmp); err != nil {
		return err
	}
	bf.Write(tmp)

	log.Debug("inbuffer data: %v", bf.Data[:])
	return nil
}

//WritePacket 将包写出
func (lbc *LengthBasedCodec) Write(conn net.Conn, bf *buffer.Buffer, p []byte) error {
	//fmt.Println("to write", p)
	tLen := 8 + len(p)
	bf.Reset(0, tLen)
	hLen := 0

	if lbc.byteOrder == binary.BigEndian {
		bf.WriteUint32BE(uint32(tLen))
		bf.WriteUint32BE(uint32(hLen))
	}
	if lbc.byteOrder == binary.LittleEndian {
		bf.WriteUint32LE(uint32(tLen))
		bf.WriteUint32LE(uint32(hLen))
	}
	bf.Write(p)

	if lbc.maxSize > 0 && bf.Length() > lbc.maxSize {
		return PacketTooLargeError
	}

	//fmt.Println("outbuffer data:", bf.Data)
	if _, err := conn.Write(bf.Data); err != nil {
		fmt.Printf("写出数据错误: %s\n", err)
		return err
	}

	return nil
}

func (lbc *LengthBasedCodec) Marshal(m interface{}) (*Packet, error) {
	return lbc.decoder.Decode(m)
}

func (lbc *LengthBasedCodec) Unmarshal(p *Packet) (interface{}, error) {
	return lbc.encoder.Encode(p)
}
