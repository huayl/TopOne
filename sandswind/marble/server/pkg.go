package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"sandswind/marble/consts"
	"sandswind/marble/errors"
	"sandswind/marble/log"
	"sandswind/marble/nmod"
	"sandswind/marble/utils"
)

type Header struct {
	Version uint16
	Bodylen uint32
	Addr    uint32
	Mothed  uint16
	AppId   uint32
}

//client接入处理主函数
func ModCall(conn net.Conn, h *Header, body []byte) {
	result := handlerFunc(h, body)
	if result != nil {
		msg := BuildAnswer(h, result)
		Send(conn, msg)
	}
}

//调用功能模块函数
func handlerFunc(h *Header, body []byte) map[string]interface{} {
	result := make(map[string]interface{})
	jreq := make(map[string]interface{})
	if bytes.Equal(nil, body) != true {
		err := json.Unmarshal(body, &jreq)
		if err != nil {
			log.Error("handlerFunc|Json_Unmarshal|err|%v", err)
			result["ecode"] = "2001"
			result["emsg"] = "request_error"
			return result
		}
	}
	//获取method
	method, _ := GetMethodStr(h.Mothed)
	if method == "" {
		result["ecode"] = -101
		result["emsg"] = "method_is_empty"
		return result
	}

	return nmod.DoWork(method, jreq)
}

//业务码转换
func GetMethodStr(method uint16) (string, bool) {
	switch method {
	//test
	case consts.MOD_REQ_TEST:
		return "test", false
	default:
		log.Error("unknown_method|%v", method)
	}
	return "", false
}

//发送client应答
func Send(conn net.Conn, buff *bytes.Buffer) error {
	err := sendMsg(conn, buff.Bytes())
	utils.PutBuff(buff)
	return err

}

//发送报文数据给client
func sendMsg(conn net.Conn, buff []byte) error {
	_, err := conn.Write(buff)
	if err != nil {
		log.Error("Send|%v", err)
		conn.Close()
		return err

	} else {
		return nil

	}
}

//组建结果包
func BuildPkg(h *Header, buildBody func(*bytes.Buffer)) *bytes.Buffer {
	var headlen uint32 = 14

	sendbuf := utils.GetBuff()
	sendbuf.Grow(int(headlen))

	var lenpos int
	binary.Write(sendbuf, binary.LittleEndian, h.Version)
	lenpos = sendbuf.Len()
	binary.Write(sendbuf, binary.LittleEndian, uint32(0))
	binary.Write(sendbuf, binary.LittleEndian, h.Addr)
	binary.Write(sendbuf, binary.LittleEndian, h.Mothed+1)
	binary.Write(sendbuf, binary.LittleEndian, h.AppId)
	if buildBody != nil {
		buildBody(sendbuf)
		var bodylen uint32 = uint32(sendbuf.Len()) - headlen
		b := sendbuf.Bytes()
		binary.LittleEndian.PutUint32(b[lenpos:], bodylen)
	}
	return sendbuf
}

//组建结果报文
func BuildAnswer(h *Header, body interface{}) *bytes.Buffer {
	return BuildPkg(h, func(b *bytes.Buffer) {
		bb, err := json.Marshal(body)
		if err != nil {
			log.Error("BuildMsgJson|Marshal err|%v", err)
			panic(errors.ETFail)
		}
		b.Write(bb)
		log.Info("BuildMsgJson|%v", bb)
	})
}
