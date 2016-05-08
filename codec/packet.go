package codec

import (
	"encoding/binary"
)

//Packet 数据包接口
type Packet interface {
	Encode(bo binary.ByteOrder) ([]byte, error) // packet --> bytes
	Transform(m Message) error                  // packet --> message
}
