package buffer
import (
	"bytes"
	"runtime"
	"testing"
	convey "github.com/smartystreets/goconvey/convey"
)


func Test_Buffer(t *testing.T) {
	convey.Convey("All func should be passed", t, func() {
		var buffer = newOutBuffer()
		PrepareBuffer(buffer)
		VerifyBuffer(&InBuffer{Data: buffer.Data})
	})
}

func PrepareBuffer(buffer *OutBuffer) {
	buffer.WriteUint8(123)
	buffer.WriteUint16LE(0xFFEE)
	buffer.WriteUint16BE(0xFFEE)
	buffer.WriteUint32LE(0xFFEEDDCC)
	buffer.WriteUint32BE(0xFFEEDDCC)
	buffer.WriteUint64LE(0xFFEEDDCCBBAA9988)
	buffer.WriteUint64BE(0xFFEEDDCCBBAA9988)
	buffer.WriteFloat32LE(88.01)
	buffer.WriteFloat64LE(99.02)
	buffer.WriteFloat32BE(88.01)
	buffer.WriteFloat64BE(99.02)
	buffer.WriteRune('好')
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))
	buffer.WriteBytes([]byte("Hello3"))
	buffer.WriteVarint(0x7FEEDDCCBBAA9988)
	buffer.WriteUvarint(0xFFEEDDCCBBAA9988)
}

func VerifyBuffer(buffer *InBuffer) {
	convey.ShouldEqual(buffer.ReadUint8(), 123)
	convey.ShouldEqual(buffer.ReadUint16LE(), 0xFFEE)
	convey.ShouldEqual(buffer.ReadUint16BE(), 0xFFEE)
	convey.ShouldEqual(buffer.ReadUint32LE(), 0xFFEEDDCC)
	convey.ShouldEqual(buffer.ReadUint32BE(), 0xFFEEDDCC)
	convey.ShouldEqual(buffer.ReadUint64LE()==0xFFEEDDCCBBAA9988, true)
	convey.ShouldEqual(buffer.ReadUint64BE()==0xFFEEDDCCBBAA9988, true)
	convey.ShouldEqual(buffer.ReadFloat32LE(), 88.01)
	convey.ShouldEqual(buffer.ReadFloat64LE(), 99.02)
	convey.ShouldEqual(buffer.ReadFloat32BE(), 88.01)
	convey.ShouldEqual(buffer.ReadFloat64BE(), 99.02)
	convey.ShouldEqual(buffer.ReadRune(), '好')
	convey.ShouldEqual(buffer.ReadString(6), "Hello1")
	convey.ShouldEqual(bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")), true)
	convey.ShouldEqual(bytes.Equal(buffer.Slice(6), []byte("Hello3")), true)
	convey.ShouldEqual(buffer.ReadVarint()== 0x7FEEDDCCBBAA9988, true)
	convey.ShouldEqual(buffer.ReadUvarint()== 0xFFEEDDCCBBAA9988, true)
}

func Benchmark_NewBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := newInBuffer()
		x.free()
	}
	b.StopTimer()
	state := BufferPoolState()
	b.Logf("Hit Rate: %2.2f%%", state.InHitRate*100.0)
	b.StartTimer()
}

func Benchmark_SetFinalizer1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, func(x *InBuffer) {
		})
	}
}

func Benchmark_SetFinalizer2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, nil)
	}
}

func Benchmark_MakeBytes_512(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 512)
	}
}

func Benchmark_MakeBytes_1024(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

func Benchmark_MakeBytes_4096(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 4096)
	}
}

func Benchmark_MakeBytes_8192(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 8192)
	}
}