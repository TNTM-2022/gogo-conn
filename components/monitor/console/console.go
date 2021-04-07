package console

import (
	"context"
	"encoding/json"
	"fmt"
	cfg "gogo-connector/components/config"
	"gogo-connector/components/global"
	"log"
	"runtime"
)

type MonitListInfoBody struct {
	ServerID   string  `json:"serverId"`
	ServerType string  `json:"serverType"`
	Pid        int     `json:"pid"`
	RSS        uint64  `json:"rss"`
	HeapTotal  uint64  `json:"heapTotal"`
	HeapUsed   uint64  `json:"heapUsed"`
	Uptime     float64 `json:"uptime"`
}
type MonitListInfo struct {
	ServerID string            `json:"serverId"`
	Body     MonitListInfoBody `json:"body"`
}

func MonitorHandler(signal string, quitFn context.CancelFunc, blackList []string) (req, respBody, respErr, notify []byte) {
	switch signal {
	case "stop", "kill": // todo kill 有返回值
		{
			stop(quitFn)
		}
	case "list":
		{
			respErr = list()
		}
	case "blacklist":
		{
			blacklist(blackList)
		}
	case "addCron", "removeCron", "restart": // todo restart 有返回值，实现 restart
	default:
		log.Printf("receive error signal\n")
	}
	return
}
func stop(quitFn context.CancelFunc) {
	quitFn()
}

func list() []byte {
	monitInf := MonitListInfo{
		ServerID: *cfg.ServerID,
		Body: MonitListInfoBody{
			ServerID:   *cfg.ServerID,
			ServerType: *cfg.ServerType,
			Pid:        cfg.Pid,
			Uptime:     cfg.Uptime(),
		},
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	monitInf.Body.HeapTotal = memStats.HeapSys / (1024 * 1024)
	monitInf.Body.HeapUsed = memStats.HeapInuse / (1024 * 1024)
	monitInf.Body.RSS = memStats.StackSys / (1024 * 1024)

	result, e := json.Marshal(monitInf)
	if e != nil {
		log.Println(e)
	}
	return result
}

func blacklist(blackList []string) {
	for _, v := range blackList { // todo 不做ip 校验
		global.BlackList.Set(v, true)
	}
}

func restart() {
	fmt.Println("not implement yet. do it manually")
}