package test

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

type Header struct {
	Id  uint64
	Ext []byte
}

type Packet struct {
	TLen   uint32
	HLen   uint32
	Header Header
	Body   []byte
}

type Person struct {
	Id   int64
	Name string
	Aage int8
	Desc string
}

func FromBytes(buffer []byte) *Packet {
	p := &Packet{
		Header: Header{},
	}
	p.TLen = binary.BigEndian.Uint32(buffer[:4])
	p.HLen = binary.BigEndian.Uint32(buffer[4:8])

	headerBuffer := buffer[8 : 8+p.HLen]
	p.Header.Id = binary.BigEndian.Uint64(headerBuffer[:8])
	p.Header.Ext = headerBuffer[8:]

	p.Body = buffer[8+p.HLen:]

	return p
}

func ToBytes(packet *Packet) []byte {
	buffer := make([]byte, packet.TLen)
	binary.BigEndian.PutUint32(buffer[0:4], packet.TLen)
	binary.BigEndian.PutUint32(buffer[4:8], packet.HLen)

	headerBuffer := make([]byte, packet.HLen)
	binary.BigEndian.PutUint64(headerBuffer[0:8], packet.Header.Id)
	copy(headerBuffer[8:], packet.Header.Ext[:])

	copy(buffer[8:8+packet.HLen], headerBuffer[:])
	copy(buffer[8+packet.HLen:], packet.Body)
	return buffer
}

func Encode(ps *Person) *Packet {
	p := &Packet{
		TLen:   0,
		HLen:   0,
		Header: Header{},
		Body:   []byte{},
	}

	buffer := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(ps)
	if err != nil {
		return nil
	}

	p.Body = buffer.Bytes()
	p.Header.Id = 100
	p.Header.Ext = []byte{}
	p.HLen = 8
	p.TLen = 4 + 4 + p.HLen + uint32(len(p.Body))

	return p
}

func Decode(p *Packet) *Person {
	pData := p.Body

	var person Person
	dec := gob.NewDecoder(bytes.NewBuffer(pData))

	if e := dec.Decode(&person); e == nil {
		return &person
	}

	return nil
}
