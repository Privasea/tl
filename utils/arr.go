package utils

import "time"

func InArray(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}

// Millisec 转换为毫秒
func Millisec(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
