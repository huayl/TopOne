package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"sandswind/marble/log"
	"sandswind/marble/utils"
)

type Service struct {
	Http         string
	Https        string
	Tcp          string
	CpuNum       int32 `toml:"cpu_num"`
	AccTimeout   int32 `toml:"accept_timeout"`
	ReadTimeout  int32 `toml:"read_timeout"`
	WriteTimeout int32 `toml:"write_timeout"`
}

var LocalConfig Service

func Init(configFile string) {
	flag := true
	jsonstr, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Warn("read_file_error|%v", err)
		flag = false
	} else {
		if _, err := toml.Decode(string(jsonstr), &LocalConfig); err != nil {
			log.Warn("toml.Decode|%v", err)
			flag = false
		}
	}

	if !flag {
		LocalConfig.Http = ":18080"
		LocalConfig.Https = ":18443"
		LocalConfig.Tcp = ":18000"
		LocalConfig.CpuNum = 2
		LocalConfig.AccTimeout = 60
		LocalConfig.ReadTimeout = 60
		LocalConfig.WriteTimeout = 60
	}

	log.Info("==== LocalConfig ====")
	utils.PrintObject(LocalConfig)
	log.Info("=====================")

}
