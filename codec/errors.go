package codec

import (
	"errors"
)

// Errors
var (
	SendToClosedError   = errors.New("Send to closed session")
	BlockingError       = errors.New("Blocking happened")
	PacketTooLargeError = errors.New("Packet too large")
	PacketTooSmallError = errors.New("Packet too small")
	NilBufferError      = errors.New("Buffer is nil")
	HeaderTooLargeError = errors.New("Header is larger than total packet")
	HeaderTooSmallError = errors.New("Header size should not be less then zero")
)
