package core

import "time"

func GetTimeUnixMilli() int64 {
	return time.Now().UnixMilli()
}
