package buffer

import (
	"bytes"
	convey "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_Buffer(t *testing.T) {
	convey.Convey("All func should be passed", t, func() {
		buffer := NewBuffer(0, 0)
		PrepareBuffer(buffer)
		VerifyBuffer(t, buffer)
	})
}

func PrepareBuffer(buffer *Buffer) {
	buffer.WriteVarint(0x12345678AABBCCDD)
	buffer.WriteUvarint(0x12345678AABBCCDD)
	buffer.WriteUint8(0x12)
	buffer.WriteUint16LE(0x1234)
	buffer.WriteUint16BE(0x1234)
	buffer.WriteUint24LE(0x123456)
	buffer.WriteUint24BE(0x123456)
	buffer.WriteUint32LE(0x12345678)
	buffer.WriteUint32BE(0x12345678)
	buffer.WriteUint40LE(0x12345678AA)
	buffer.WriteUint40BE(0x12345678AA)
	buffer.WriteUint48LE(0x12345678AABB)
	buffer.WriteUint48BE(0x12345678AABB)
	buffer.WriteUint56LE(0x12345678AABBCC)
	buffer.WriteUint56BE(0x12345678AABBCC)
	buffer.WriteUint64LE(0x12345678AABBCCDD)
	buffer.WriteUint64BE(0x12345678AABBCCDD)
	buffer.WriteFloat32LE(88.01)
	buffer.WriteFloat64LE(99.02)
	buffer.WriteFloat32BE(88.01)
	buffer.WriteFloat64BE(99.02)
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))

	buffer.WriteRune('好')
}

func VerifyBuffer(t *testing.T, buffer *Buffer) {
	convey.So(buffer.ReadVarint() == 0x12345678AABBCCDD, convey.ShouldBeTrue)
	convey.So(buffer.ReadUvarint() == 0x12345678AABBCCDD, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint8() == 0x12, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint16LE() == 0x1234, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint16BE() == 0x1234, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint24LE() == 0x123456, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint24BE() == 0x123456, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint32LE() == 0x12345678, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint32BE() == 0x12345678, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint40LE() == 0x12345678AA, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint40BE() == 0x12345678AA, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint48LE() == 0x12345678AABB, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint48BE() == 0x12345678AABB, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint56LE() == 0x12345678AABBCC, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint56BE() == 0x12345678AABBCC, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint64LE() == 0x12345678AABBCCDD, convey.ShouldBeTrue)
	convey.So(buffer.ReadUint64BE() == 0x12345678AABBCCDD, convey.ShouldBeTrue)
	convey.So(buffer.ReadFloat32LE() == 88.01, convey.ShouldBeTrue)
	convey.So(buffer.ReadFloat64LE() == 99.02, convey.ShouldBeTrue)
	convey.So(buffer.ReadFloat32BE() == 88.01, convey.ShouldBeTrue)
	convey.So(buffer.ReadFloat64BE() == 99.02, convey.ShouldBeTrue)
	convey.So(buffer.ReadString(6) == "Hello1", convey.ShouldBeTrue)
	convey.So(bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")), convey.ShouldBeTrue)

	r, _, _ := buffer.ReadRune()
	convey.So(r == '好', convey.ShouldBeTrue)
}
