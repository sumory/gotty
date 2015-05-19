package codec

import (
	"fmt"
	"testing"
)

func TestDelimiters(t *testing.T) {
	fmt.Println("Testing delimiters")

	fmt.Printf("%s %s %s %s\n", CRLF, CRLF_DELIMITER, LF_DELIMIER, STR_CRLF)
}
