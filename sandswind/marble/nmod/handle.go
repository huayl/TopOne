package nmod

import (
	"sandswind/marble/consts"
	"sandswind/marble/errors"
	"sandswind/marble/log"
)

//test
func test(streq map[string]interface{}) map[string]interface{} {
	log.Debug("test|reqinfo:|%v", streq)
	resp := map[string]interface{}{
		consts.RES_ECODE: errors.ETSucc.GetCode(),
		consts.RES_EMSG:  errors.ETSucc.GetMessage(),
	}
	return resp
}
