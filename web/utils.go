package web

import (
	"strings"
	"unicode"
	"unsafe"
)

func SubStringLast(str, substr string) string {
	index := strings.Index(str, substr)
	if index != -1 {
		return str
	} else {
		return str[index+len(substr):]
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
