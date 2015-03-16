package utils

import (
	"bytes"

	"fmt"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"sandswind/marble/errors"
	"sandswind/marble/log"
	"strconv"
	"strings"
	"sync"
)

func PrintPanic(err interface{}) {
	if err != nil {
		log.Error("* panic:|%v", err)
	}
	for skip := 2; ; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fmt.Printf("* %v : %v\n", file, line)
	}
}

//转换数字字符串
func StrToInt32(instr string) (int32, error) {
	tmp, err := strconv.ParseInt(instr, 10, 32)
	if err != nil {
		log.Error("StrToInt32|Invalid_String|%v|%v", err, instr)
		return 0, errors.ETDatafmt
	}
	return int32(tmp), nil
}

//转换数字字符串
func StrToUint32(instr string) (uint32, error) {
	tmp, err := strconv.ParseUint(instr, 10, 32)
	if err != nil {
		log.Error("StrToUint32|Invalid_String|%v|%v", err, instr)
		return 0, errors.ETDatafmt
	}
	return uint32(tmp), nil
}

//转换数字字符串
func StrToUint64(instr string) (uint64, error) {
	tmp, err := strconv.ParseUint(instr, 10, 64)
	if err != nil {
		log.Error("StrToUint32|Invalid_String|%v|%v", err, instr)
		return 0, errors.ETDatafmt
	}
	return tmp, nil
}

//转换为Bool
func ToBool(arg interface{}) (bool, error) {
	if arg == nil {
		return false, errors.ETDataNil
	}
	switch v := arg.(type) {
	case bool:
		return v, nil
	default:
		return false, errors.ETDatafmt
	}
}

func MaskStringChar(str string, s int, _e int, c byte) string {
	if len(str) <= s+_e {
		return str
	}
	var result = make([]byte, len(str))
	copy(result, str)
	e := len(str) - _e
	for i := s; i < e; i++ {
		result[i] = c
	}
	return string(result)
}

func MaskString(str string, s int, e int) string {
	return MaskStringChar(str, s, e, '*')
}

func GenRandomKey() string {
	buff := GetBuff()
	buff.Grow(16)
	for i := 0; i < 16; i++ {
		buff.WriteByte(byte(rand.Intn(256)))
	}
	r := ToMd5(buff.Bytes())
	PutBuff(buff)
	return r
}

func IpToUint32(strIp string) (uint32, error) {
	parts := strings.Split(strIp, ".")
	if len(parts) != 4 {
		log.Error("ip_node_len_not_eq_4|%v", strIp)
		return 0, errors.ETParam
	}
	intip0, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, err
	}
	intip1, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, err
	}
	intip2, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return 0, err
	}
	intip3, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return 0, err
	}
	ip := uint32(intip0<<24 + intip1<<16 + intip2<<8 + intip3)
	return ip, nil
}

func Uint32ToIp(ip uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

//转换为字符串
func ToString(arg interface{}) (string, error) {
	if arg == nil {
		return "", errors.ETDataNil
	}
	switch v := arg.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

//生成随机数字
func GenRandNumStr(max int32) string {
	s := fmt.Sprintf("%0*d", max, rand.Int63())
	return s[len(s)-int(max):]
}

//生成随机字符串
func GenRandAsciiStr(max int32) string {
	buff := GetBuff()
	buff.Grow(int(max))
	for i := 0; i < int(max); i++ {
		buff.WriteByte('a' + byte(rand.Intn(26)))
	}
	r := buff.String()
	PutBuff(buff)
	return r
}

var buffPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0))
	},
}

func GetBuff() *bytes.Buffer {
	return buffPool.Get().(*bytes.Buffer)
}

func PutBuff(buf *bytes.Buffer) {
	buf.Reset()
	buffPool.Put(buf)
}

func PrintObject(o interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("ObjectToString|Panic|%v", err)
		}
	}()
	v := reflect.ValueOf(o)
	for i := 0; i < v.NumField(); i++ {
		log.Info("Object ", v.Type().Field(i).Name, v.Field(i).Interface())
	}
}

func GetLocalIp() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error("get_interface_addrs_failed:|%v", err)
		return nil, err
	}
	result := make([]string, 0, 4)
	for _, addr := range addrs {
		strip := addr.String()
		strip = strings.Split(strip, "/")[0]
		log.Debug("find ip:|%v", strip)
		if strip == "127.0.0.1" {
			log.Debug("skip|lo")
			continue
		}
		if addr.Network() != "ip+net" && addr.Network() != "ip" {
			log.Debug("skip|not_ip|%v|%v", addr.String(), addr.Network())
			continue
		}
		if strings.Contains(strip, ":") {
			log.Debug("skip|ipv6|%v", strip)
			continue
		}
		result = append(result, strip)
	}
	return result, nil
}
