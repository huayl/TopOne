package main

import (
	"sandswind/marble/log"
	"time"
)

func main() {
	log.Init("")
	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		log.Error("idfijfidjdfifjigjfigfjigf")
		log.Error("dhfduf======%d.%s---%s", 23, "323", "#@#@#@#@AAAAAAAAAAAADSDS")
	}
}
