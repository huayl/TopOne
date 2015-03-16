package nmod

import (
	"sandswind/marble/consts"
	"sandswind/marble/errors"
	"sandswind/marble/log"
	"sandswind/marble/utils"
)

//异常处理
func ErrToResp(e interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if err, ok := e.(*errors.SysError); ok {
		result[consts.RES_ECODE] = err.GetCode()
		result[consts.RES_EMSG] = err.GetMessage()
		log.Debug("ErrToResp|ecode:%v|emsg:%v|exmsg:%v", err.GetCode(), err.GetMessage(), err.GetExMessage())

	} else {
		result[consts.RES_ECODE] = errors.ETFail.GetCode()
		result[consts.RES_EMSG] = errors.ETFail.GetMessage()
		utils.PrintPanic(e)
	}
	return result
}

// 业务功能处理模块
func DoWork(method string, req map[string]interface{}) (resp map[string]interface{}) {
	defer func() {
		if err := recover(); err != nil {
			resp = ErrToResp(err)
		}
	}()
	log.Debug("DoWork|method:%v", method)
	switch method {
	//test
	case "test":
		return test(req)
	default:
		resp := map[string]interface{}{
			consts.RES_ECODE: errors.ETNoMod.GetCode(),
			consts.RES_EMSG:  errors.ETNoMod.GetMessage(),
		}
		return resp
	}
}
