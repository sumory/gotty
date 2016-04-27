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
	nBit      uint8            //描述包总大小和head大小的整数类型字节数，1，2，4，8
	byteOrder binary.ByteOrder //大小端
	totalLen  []byte           //包第一部分,一般指包总大小
	headerLen []byte           //包第二部分,一般指包头大小
	maxSize   int              //包最大长度
}

//NewLengthBasedCodec 新建定长编解码器
func NewLengthBasedCodec(nBit uint8, byteOrder binary.ByteOrder) *LengthBasedCodec {
	return &LengthBasedCodec{
		nBit:      nBit,
		byteOrder: byteOrder,
		totalLen:  make([]byte, nBit),
		headerLen: make([]byte, nBit),
		maxSize:   32 * 1024, //32k
	}
}

//ReadPacket 从连接中读取报文放入buffer中
func (lbc *LengthBasedCodec) Read(conn net.Conn, bf *buffer.Buffer) error {
	//读包总大小
	if _, err := io.ReadFull(conn, lbc.totalLen); err != nil {
		return err
	}

	tLen := 0 //包总长度
	switch lbc.nBit {
	case 1:
		tLen = int(lbc.totalLen[0])
	case 2:
		tLen = int(lbc.byteOrder.Uint16(lbc.totalLen))
	case 4:
		tLen = int(lbc.byteOrder.Uint32(lbc.totalLen))
	case 8:
		tLen = int(lbc.byteOrder.Uint64(lbc.totalLen))
	}
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
	hLen := 0
	switch lbc.nBit {
	case 1:
		hLen = int(lbc.headerLen[0])
	case 2:
		hLen = int(lbc.byteOrder.Uint16(lbc.headerLen))
	case 4:
		hLen = int(lbc.byteOrder.Uint32(lbc.headerLen))
	case 8:
		hLen = int(lbc.byteOrder.Uint64(lbc.headerLen))
	}

	if hLen > tLen {
		return HeaderLargerThanTotal
	}

	log.Debug("totallen: %d, headerLen: %d", tLen, hLen)
	bf.Reset(0, tLen)
	bf.Write(lbc.totalLen)  //将包总长度写入buffer
	bf.Write(lbc.headerLen) //将包头长度写入buffer

	tmp := make([]byte, tLen-int(lbc.nBit*2))
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
	tLen := int(lbc.nBit*2) + len(p)
	bf.Reset(0, tLen)
	hLen := 0

	if lbc.byteOrder == binary.BigEndian {
		switch lbc.nBit {
		case 1:
			bf.WriteUint8(uint8(tLen))
			bf.WriteUint8(uint8(hLen))
		case 2:
			bf.WriteUint16BE(uint16(tLen))
			bf.WriteUint16BE(uint16(hLen))
		case 4:
			bf.WriteUint32BE(uint32(tLen))
			bf.WriteUint32BE(uint32(hLen))
		case 8:
			bf.WriteUint64BE(uint64(tLen))
			bf.WriteUint64BE(uint64(hLen))
		}
	}
	if lbc.byteOrder == binary.LittleEndian {
		switch lbc.nBit {
		case 1:
			bf.WriteUint8(uint8(tLen))
			bf.WriteUint8(uint8(hLen))
		case 2:
			bf.WriteUint16LE(uint16(tLen))
			bf.WriteUint16LE(uint16(hLen))
		case 4:
			bf.WriteUint32LE(uint32(tLen))
			bf.WriteUint32LE(uint32(hLen))
		case 8:
			bf.WriteUint64LE(uint64(tLen))
			bf.WriteUint64LE(uint64(hLen))
		}
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

func (lbc *LengthBasedCodec) Marshal(m Message) *Packet {
	return nil
}

func (lbc *LengthBasedCodec) Unmarshal(p *Packet) (Message, error) {
	return nil, nil
}
