package web

import "strings"

func SubStringLast(str, substr string) string {
	index := strings.Index(str, substr)
	if index != -1 {
		return str
	} else {
		return str[index+len(substr):]
	}
}
