package codec

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
)

//Message 消息体接口
type Message interface {
	Length() int
	FromPacket(p Packet) error
	Bytes() ([]byte, error)
}

type MessageFunc func() ([]byte, error)

func (msgFunc MessageFunc) Length() int {
	bytes, err := msgFunc.Bytes()
	if err != nil {
		return 0
	}

	return len(bytes)
}

func (msgFunc MessageFunc) FromPacket(p Packet) error {
	//not implemented
	return nil
}

func (msgFunc MessageFunc) Bytes() ([]byte, error) {
	return msgFunc()
}

func ByteMsg(b []byte) Message {
	return MessageFunc(func() ([]byte, error) {
		return b, nil
	})
}

func StringMsg(s string) Message {
	return MessageFunc(func() ([]byte, error) {
		return []byte(s), nil
	})
}

func JsonMsg(v interface{}) Message {
	return MessageFunc(func() ([]byte, error) {
		return json.Marshal(v)
	})
}

func XmlMsg(v interface{}) Message {
	return MessageFunc(func() ([]byte, error) {
		return xml.Marshal(v)
	})
}

func PacketMsg(bo binary.ByteOrder, p Packet) Message {
	return MessageFunc(func() ([]byte, error) {
		return p.Encode(bo)
	})
}
