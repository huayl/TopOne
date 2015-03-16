package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"
)

func TimestampToTimeString(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2008-06-01 15:04:05")
}

func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// 计算md5长度32的字符串(md5)
func StringToMd5(s string) string {
	m := md5.New()
	io.WriteString(m, s)
	return fmt.Sprintf("%x", m.Sum(nil))
}
