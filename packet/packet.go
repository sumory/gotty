package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	//"log"
)

const (
	//最大packet的字节数
	MAX_PACKET_BYTES = 32 * 1024
	MIN_PACKET_BYTES = 128         //128个字节
	PACKET_HEAD_LEN  = (4 + 1 + 4) //请求头部长度	 4 字节+ 1字节 + 4字节 + data + \r + \n
)

var CMD_CRLF = []byte{'\r', '\n'}
var CMD_STR_CRLF = "\r\n"
var ERROR_PACKET_TYPE = errors.New("unmatches packet type ")

//按协议封装的Packet
type Packet struct {
	Opaque  int32 //标识
	CmdType uint8 //类型
	Data    []byte
}

func NewPacket(cmdtype uint8, data []byte) *Packet {
	return &Packet{
		Opaque:  -1,
		CmdType: cmdtype,
		Data:    data}
}

func NewRespPacket(opaque int32, cmdtype uint8, data []byte) *Packet {
	p := NewPacket(cmdtype, data)
	p.Opaque = opaque
	return p
}

func (self *Packet) Reset() {
	self.Opaque = -1
}

//TODO:
//限制包大小&优化buffer使用
func (self *Packet) marshal() []byte {
	//总长度	4 字节+ 1字节 + 4字节 + var + \r + \n
	dl := 0
	if nil != self.Data {
		dl = len(self.Data)
	}
	length := PACKET_HEAD_LEN + dl + 2
	buffer := make([]byte, length)

	binary.BigEndian.PutUint32(buffer[0:4], uint32(self.Opaque))    // 请求id
	buffer[4] = self.CmdType                                        //数据类型
	binary.BigEndian.PutUint32(buffer[5:9], uint32(len(self.Data))) //总数据包长度
	copy(buffer[9:9+dl], self.Data)
	copy(buffer[9+dl:], CMD_CRLF) //添加结束标记

	return buffer
}

func (self *Packet) unmarshal(b []byte) error {
	var paramLength = len(b)
	if paramLength < PACKET_HEAD_LEN {
		return errors.New(fmt.Sprintf("packet length error: %d/%d", paramLength, PACKET_HEAD_LEN))
	}

	// if paramLength < MIN_PACKET_BYTES || paramLength > MAX_PACKET_BYTES {
	// 	log.Println(fmt.Sprintf("packet length illegal: %d", paramLength))
	// 	//return errors.New(fmt.Sprintf("packet length illegal: %d", paramLength))
	// }

	self.Opaque = int32(binary.BigEndian.Uint32(b[:4]))
	self.CmdType = b[4]
	dataLength := binary.BigEndian.Uint32(b[5:9])

	if dataLength > 0 {
		if int(dataLength) == len(b[9:]) && dataLength <= MAX_PACKET_BYTES {
			self.Data = make([]byte, dataLength)
			copy(self.Data, b[9:])
		} else {
			if dataLength > MAX_PACKET_BYTES {
				return errors.New(fmt.Sprintf("paket is to large: %d > %d", dataLength, MAX_PACKET_BYTES))
			}
			return errors.New("Corrupt PacketData ")
		}
	} else {
		return errors.New("param has no data")
	}

	return nil
}

func MarshalPacket(packet *Packet) []byte {
	return packet.marshal()
}

func UnmarshalTLV(packet []byte) (*Packet, error) {
	packet = bytes.TrimSuffix(packet, CMD_CRLF)

	tlv := &Packet{}
	err := tlv.unmarshal(packet)
	if nil != err {
		return tlv, err
	} else {
		return tlv, nil
	}
}
