package stats

import (
	"fmt"
	"sandswind/marble/log"
	"bytes"
)

// 统计状态
const (
	COUNTCONN  = "conn"     // 连接client数目
	COUNTRECV  = "recv"     // 接受消息失败
	COUNTSEND  = "send"     // 发送消息失败
	COUNTBUSIN = "business" // 业务处理消息失败
)

var (
	countList = NewCounters("CList")
)

func AddConn() {
	countList.Add(COUNTCONN, 1)
}

func DelConn() {
	countList.Add(COUNTCONN, -1)
}

func AddRecv() {
	countList.Add(COUNTRECV, 1)
}


func AddSend() {
	countList.Add(COUNTSEND, 1)
}


func AddBusin() {
	countList.Add(COUNTBUSIN, 1)
}

// startTime -- endTime区间的报告
func GetStats() {
	
	var buffer bytes.Buffer
	cList := countList.Counts()

	buffer.WriteString(fmt.Sprintf("\n统计信息如下: "))

	if uvalue, ok := cList[COUNTCONN]; ok {
		buffer.WriteString(fmt.Sprintf("\n当前连接client数目: %v人", uvalue))
	}

	if uvalue, ok := cList[COUNTRECV]; ok {
		buffer.WriteString(fmt.Sprintf("\n接受消息失败: %v次", uvalue))
	}

		if uvalue, ok := cList[COUNTSEND]; ok {
		buffer.WriteString(fmt.Sprintf("\n接受消息失败: %v次", uvalue))
	}

		if uvalue, ok := cList[COUNTBUSIN]; ok {
		buffer.WriteString(fmt.Sprintf("\n业务处理消息失败: %v次\n", uvalue))
	}
	// 写日志
	log.Info(buffer.String())
}
