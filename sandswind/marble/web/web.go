package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sandswind/marble/log"
	"strings"
)

var (
	ipheader = [...]string{
		"remote-host",
		"x-real-ip",
		"x-forwarded-for",
		"Proxy-Client-IP",
		"WL-Proxy-Client-IP",
	}
)
var errInternal = []byte(`{"result": 44, "msg": "internal_err"}`)

const (
	WEBSERVICE = "/service/"
)

//获取连接客户端地址
func getClientIp(r *http.Request) string {
	for _, header := range ipheader {
		ip := strings.Trim(r.Header.Get(header), " ")
		if ip == "" || strings.ToLower(ip) == "unknown" {
			continue
		}
		if header == "x-forwarded-for" {
			return strings.SplitN(ip, ",", 2)[0]
		} else {
			return ip
		}
	}
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}

func modruntime(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	result := make(map[string]interface{})
	if r.Method != "POST" {
		return nil
	}

	log.Info("Client ip-address is: %s", getClientIp(r))
	jreq := make(map[string]interface{})
	r.Body = http.MaxBytesReader(w, r.Body, 4*1024*1024)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("ReadAll|%s", err.Error())
		result["result"] = -101
		result["msg"] = "request_error"
		return result
	}
	log.Info("body|%s", body)
	err = json.Unmarshal(body, &jreq)
	if err != nil {
		log.Error("Unmarshal|%s", err.Error())
		result["result"] = -101
		result["msg"] = "request_error"
		return result
	}

	result["result"] = 100
	result["msg"] = "request_error"
	return result

	//log.Info("method:|%v", method)
	//if err := bizz.CheckArgments(jreq, bizz.VER_V3, method); err != nil {
	//	var resp = bizz.ErrorToV3Result(err)
	//	resp["reqdata"] = jreq["reqdata"]
	//	return resp
	//}
	//return bizz.V3DoWork(method, jreq)
}

func modhandler(w http.ResponseWriter, r *http.Request) {
	result := modruntime(w, r)
	if result != nil {
		if m, e := json.Marshal(result); e != nil {
			log.Error("Marshal|%v|%v", result, e)
			w.Write(errInternal)
		} else {
			log.Info("resp|%s", m)
			w.Write(m)
		}
	}
}

// Web主处理
func Start(httpaddr string) {
	log.Info("Start listen on:%v", httpaddr)
	log.Info("Start web uri:%v", WEBSERVICE)

	http.HandleFunc(WEBSERVICE, modhandler)

	err := http.ListenAndServe(httpaddr, nil)
	log.Fatal("web|ListenAndServe|%v", err)
}
