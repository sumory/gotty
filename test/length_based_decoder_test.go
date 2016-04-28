package test

import (
	"bufio"
	"bytes"
	"github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestBufio(t *testing.T) {

	convey.Convey("Test bufio", t, func() {
		data := "12345678901234567890字母ABCDEFGHIJKLMNOPQRS我\r\nTUVWXYZ1234567\r\n890"

		s := strings.NewReader(data)
		br := bufio.NewReader(s)
		b := make([]byte, len([]byte(data)))

		line, isPrefix, err := br.ReadLine()
		convey.So(bytes.Equal(line, []byte("12345678901234567890字母ABCDEFGHIJKLMNOPQRS我")), convey.ShouldBeTrue)

		line, isPrefix, err = br.ReadLine()
		convey.So(bytes.Equal(line, []byte("TUVWXYZ1234567")), convey.ShouldBeTrue)

		line, isPrefix, err = br.ReadLine()
		convey.So(bytes.Equal(line, []byte("890")), convey.ShouldBeTrue)

		line, isPrefix, err = br.ReadLine()
		convey.So(line, convey.ShouldBeEmpty)
		convey.So(isPrefix, convey.ShouldBeFalse)
		convey.So(err, convey.ShouldNotBeNil)

		n, err := br.Read(b)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(n, convey.ShouldEqual, 0)

		convey.So(err, convey.ShouldNotBeNil)

	})
}

func TestBytes(t *testing.T) {
	convey.Convey("Testing bytes", t, func() {
		h := Header{
			Id:  10,
			Ext: []byte("abcdefgh=header"),
		}
		hLen := 8 + len(h.Ext)

		data := []byte("1234567890abcdefghijklmnopqrstuvwxyz中文繁體ａｃ\\r\\n$%^&(_+../@")
		tLen := 4 + 4 + hLen + len(data)

		convey.ShouldBeLessThan(hLen, tLen)

		p := &Packet{
			TLen:   uint32(tLen),
			HLen:   uint32(hLen),
			Header: h,
			Body:   data,
		}

		b := ToBytes(p)
		//fmt.Printf("bytes: %v %d\n", b, len(b))

		pp := FromBytes(b)
		convey.So(p.TLen, convey.ShouldEqual, pp.TLen)
		convey.So(p.HLen, convey.ShouldEqual, pp.HLen)
		convey.So(p.Header.Id, convey.ShouldEqual, pp.Header.Id)
		convey.So(bytes.Equal(p.Header.Ext, pp.Header.Ext), convey.ShouldBeTrue)
		convey.So(bytes.Equal(p.Body, pp.Body), convey.ShouldBeTrue)
		convey.So(bytes.Equal(p.Body, data), convey.ShouldBeTrue)
	})
}

func TestDecoder(t *testing.T) {
	convey.Convey("Testing encoder/decoder", t, func() {
		ps := new(Person)
		ps.Id = 1
		ps.Aage = 123
		ps.Name = "sumory"
		ps.Desc = "this is me!"

		packet := Encode(ps)
		person := Decode(packet)

		convey.So(person.Id, convey.ShouldEqual, ps.Id)
		convey.So(person.Aage, convey.ShouldEqual, ps.Aage)
		convey.So(person.Name, convey.ShouldEqual, ps.Name)
		convey.So(person.Desc, convey.ShouldEqual, ps.Desc)

	})
}

func TestCodec(t *testing.T) {
	convey.Convey("Testing codec, Person <-> Packet <-> bytes <-> Packet <-> Person", t, func() {
		ps := new(Person)
		ps.Id = 1
		ps.Aage = 123
		ps.Name = "sumory"
		ps.Desc = "this is me!"

		packet := Encode(ps)
		bs := ToBytes(packet)
		pp := FromBytes(bs)
		person := Decode(pp)

		convey.So(person.Id, convey.ShouldEqual, ps.Id)
		convey.So(person.Aage, convey.ShouldEqual, ps.Aage)
		convey.So(person.Name, convey.ShouldEqual, ps.Name)
		convey.So(person.Desc, convey.ShouldEqual, ps.Desc)

	})
}
