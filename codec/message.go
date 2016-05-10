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
	ToPacket() (Packet, error)
	Bytes() ([]byte, error)
}

// ~================= simple message =======================

type SimpleMessage struct {
	id  uint32
	msg string
}

func NewEmptySimpleMessage() *SimpleMessage {
	return &SimpleMessage{}
}

func NewSimpleMessage(id uint32, msg string) *SimpleMessage {
	return &SimpleMessage{
		id:  id,
		msg: msg,
	}
}

func (sm *SimpleMessage) getId() uint32 {
	return sm.id
}

func (sm *SimpleMessage) getMsg() string {
	return sm.msg
}

func (sm *SimpleMessage) Length() int {
	return len([]byte(sm.msg))
}

func (sm *SimpleMessage) FromPacket(p Packet) error {
	pp, _ := p.(LengthBasedPacket)
	sm.id = pp.Header.Sequence
	sm.msg = string(pp.Body.Data)
	return nil
}

func (sm *SimpleMessage) ToPacket() (Packet, error) {
	//not implemented
	return nil, UnImplementedError
}

func (sm *SimpleMessage) Bytes() ([]byte, error) {
	return []byte(sm.msg), nil
}

// ~================= func message =======================
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

func (msgFunc MessageFunc) ToPacket() (Packet, error) {
	//not implemented
	return nil, UnImplementedError
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
