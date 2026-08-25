package domain

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

func NewID(prefix string) string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return strings.TrimSpace(prefix) + "_" + hex.EncodeToString(b[:])
}

func StableID(prefix string, parts ...string) string {
	h := uint64(1469598103934665603)
	for _, part := range parts {
		for _, c := range []byte(part) {
			h ^= uint64(c)
			h *= 1099511628211
		}
		h ^= 0xff
		h *= 1099511628211
	}
	const digits = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		buf[i] = digits[h&15]
		h >>= 4
	}
	return prefix + "_" + string(buf)
}
