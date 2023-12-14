package work

import (
	"crypto/rand"
	"fmt"
	"io"
)

func makeIdentifier() string {
	b := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
