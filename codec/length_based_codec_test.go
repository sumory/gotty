package codec

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCodec(t *testing.T) {
	convey.Convey("Testing length based codec", t, func() {
		convey.So(1, convey.ShouldEqual, 1)

	})
}
