package codec

import (
	"encoding/binary"
	"github.com/sumory/gotty/buffer"
	"io"
	"net"
	"fmt"
)

type LengthBasedProtocol struct {
	byteOrder     binary.ByteOrder
	bufferFactory BufferFactory
	codec         LengthBasedCodec
	nBit          int //描述消息总大小和head大小的整数类型，1，2，4，8
}

func NewLengthBasedProtocol(nBit int, bo binary.ByteOrder, bf BufferFactory) *LengthBasedProtocol {
	return &LengthBasedProtocol{
		byteOrder:     bo,
		bufferFactory: bf,
		nBit:          nBit,
	}
}

func (lbp *LengthBasedProtocol) BufferFactory() BufferFactory {
	return lbp.bufferFactory
}

func (lbp *LengthBasedProtocol) NewCodec() Codec {
	return NewLengthBasedCodec(lbp.nBit, lbp.byteOrder)
}

type LengthBasedCodec struct {
	nBit      int //描述包总大小和head大小的整数类型，1，2，4，8
	byteOrder binary.ByteOrder
	T         []byte //包第一部分,一般指包总大小
	H         []byte //包第二部分,一般指包头大小
	maxSize   int    //包最大长度
}

func NewLengthBasedCodec(nBit int, byteOrder binary.ByteOrder) *LengthBasedCodec {
	return &LengthBasedCodec{
		nBit:      nBit,
		byteOrder: byteOrder,
		T:         make([]byte, nBit),
		H:         make([]byte, nBit),
		maxSize:   32 * 1024, //32k
	}
}

func (lbc *LengthBasedCodec) GetNbit() int{
	return lbc.nBit
}

//读包
func (lbc *LengthBasedCodec) ReadPacket(conn net.Conn, bf *buffer.Buffer) error {
	//读包总大小
	if _, err := io.ReadFull(conn, lbc.T); err != nil {
		return err
	}

	tLen := 0

	switch lbc.nBit {
	case 1:
		tLen = int(lbc.T[0])
	case 2:
		tLen = int(lbc.byteOrder.Uint16(lbc.T))
	case 4:
		tLen = int(lbc.byteOrder.Uint32(lbc.T))
	case 8:
		tLen = int(lbc.byteOrder.Uint64(lbc.T))
	}


	if lbc.maxSize > 0 && tLen > lbc.maxSize {
		return PacketTooLargeError
	}
	if tLen == 0 {
		return nil
	}

	//读包头大小
	if _, err := io.ReadFull(conn, lbc.H); err != nil {
		return err
	}
	hLen := 0

	switch lbc.nBit {
	case 1:
		hLen = int(lbc.H[0])
	case 2:
		hLen = int(lbc.byteOrder.Uint16(lbc.H))
	case 4:
		hLen = int(lbc.byteOrder.Uint32(lbc.H))
	case 8:
		hLen = int(lbc.byteOrder.Uint64(lbc.H))
	}

	if hLen > tLen {
		return HeaderLargerThanTotal
	}

	fmt.Println("tlen:", tLen, "hLen:", hLen)
	bf.Reset(0, tLen)
	bf.Write(lbc.T) //将包总长度写入buffer
	bf.Write(lbc.H) //将包头长度写入buffer

	tmp:=make([]byte,tLen-lbc.nBit*2)

	if _, err := io.ReadFull(conn, tmp); err != nil {
		return err
	}

	bf.Write(tmp)
	fmt.Println("inbuffer data:",bf.Data[:])

	return nil
}

//写包
func (lbc *LengthBasedCodec) WritePacket(conn net.Conn, bf *buffer.Buffer, p []byte) error {
fmt.Println("to write", p)
	tLen:=lbc.nBit*2 + len(p)
	bf.Reset(0, tLen)
	hLen:=0

	if lbc.byteOrder==binary.BigEndian{
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

	if lbc.byteOrder==binary.LittleEndian{
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


	fmt.Println("outbuffer data:", bf.Data)
	if _, err := conn.Write(bf.Data); err != nil {
		fmt.Errorf("写出数据错误", err)
		return err
	}

	return nil
}

func (lbc *LengthBasedCodec) Encode(obj *interface{}) []byte {
	return nil
}

func (lbc *LengthBasedCodec) Decode(data []byte) *interface{} {
	return nil
}
