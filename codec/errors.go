package codec

import (
	"errors"
)

// Errors
var (
	SendToClosedError     = errors.New("Send to closed session")
	BlockingError         = errors.New("Blocking happened")
	PacketTooLargeError   = errors.New("Packet too large")
	NilBufferError        = errors.New("Buffer is nil")
	HeaderLargerThanTotal = errors.New("Header is larger than total packet")
)
