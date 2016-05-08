package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

type TestMessage struct {
	ID   int
	Name string
}

func Test_Message(t *testing.T) {
	convey.Convey("Test message.length()", t, func() {
		var data = "abc"
		bs, err := StringMsg(data).Bytes()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(bs), convey.ShouldEqual, StringMsg(data).Length())
		convey.So(len([]byte(data)), convey.ShouldEqual, StringMsg(data).Length())
	})

	convey.Convey("Test byte message", t, func() {
		var data = []byte{1, 2, 3, 4, 5, 6}
		bs, err := ByteMsg(data).Bytes()
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes.Equal(bs, data), convey.ShouldEqual, true)
	})

	convey.Convey("Test string message", t, func() {
		var data = "abc"
		bs, err := StringMsg(data).Bytes()
		convey.So(err, convey.ShouldBeNil)
		convey.So(bytes.Equal(bs, []byte(data)), convey.ShouldEqual, true)
	})

	convey.Convey("Test json message", t, func() {
		bs, err := JsonMsg(TestMessage{
			ID:   123,
			Name: "sumory",
		}).Bytes()
		convey.So(err, convey.ShouldBeNil)

		var decodeJsonMessage TestMessage
		json.Unmarshal(bs, &decodeJsonMessage)
		convey.So(decodeJsonMessage.ID, convey.ShouldEqual, 123)
		convey.So(decodeJsonMessage.Name, convey.ShouldEqual, "sumory")
	})

	convey.Convey("Test xml message", t, func() {
		bs, err := XmlMsg(TestMessage{
			ID:   123,
			Name: "sumory",
		}).Bytes()
		convey.So(err, convey.ShouldBeNil)

		var decodeXmlMessage TestMessage
		xml.Unmarshal(bs, &decodeXmlMessage)
		convey.So(decodeXmlMessage.ID, convey.ShouldEqual, 123)
		convey.So(decodeXmlMessage.Name, convey.ShouldEqual, "sumory")
	})

	convey.Convey("Test Packet message", t, func() {

		header := &LengthBasedPacketHeader{
			Sequence:  1,
			Operation: 1,
			Version:   0,
			Extra:     []byte("头部扩展信息header extra"),
		}
		body := &LengthBasedPacketBody{
			Data: []byte("包体body"),
		}
		meta := &LengthBasedPacketMeta{
			TotalLen:  uint32(8 + header.Len() + body.Len()),
			HeaderLen: uint32(header.Len()),
		}

		p := LengthBasedPacket{
			Meta:   meta,
			Header: header,
			Body:   body,
		}

		bs, err := PacketMsg(binary.BigEndian, p).Bytes()
		convey.So(err, convey.ShouldBeNil)

		newPacket := NewLengthBasedPacketFromBinary(binary.BigEndian, bs)
		convey.So(newPacket.Meta.TotalLen, convey.ShouldEqual, p.Meta.Len()+p.Header.Len()+p.Body.Len())
		convey.So(newPacket.Meta.TotalLen, convey.ShouldEqual, p.Meta.TotalLen)
		convey.So(newPacket.Meta.HeaderLen, convey.ShouldEqual, p.Meta.HeaderLen)

		convey.So(newPacket.Meta.Len(), convey.ShouldEqual, p.Meta.Len())
		convey.So(newPacket.Header.Len(), convey.ShouldEqual, p.Header.Len())
		convey.So(newPacket.Body.Len(), convey.ShouldEqual, p.Body.Len())

		convey.So(newPacket.Header.Sequence, convey.ShouldEqual, p.Header.Sequence)
		convey.So(bytes.Equal(newPacket.Header.Extra, p.Header.Extra), convey.ShouldEqual, true)
		convey.So(bytes.Equal(newPacket.Body.Data, p.Body.Data), convey.ShouldEqual, true)
	})
}
