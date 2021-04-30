package console

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	cfg "go-connector/config"
	"go-connector/global"
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

	if proc, err := process.NewProcess(int32(cfg.Pid)); err == nil {
		mem, err := proc.MemoryInfo()
		if err != nil {
			fmt.Println("error=> ", err)
		}
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		monitInf.Body.HeapTotal = (memStats.HeapIdle + memStats.HeapInuse) / (1024 * 1024)
		monitInf.Body.HeapUsed = memStats.HeapInuse / (1024 * 1024)
		monitInf.Body.RSS = mem.RSS / (1024 * 1024)
	}

	result, e := json.Marshal(monitInf)
	if e != nil {
		log.Println("e=> ", e)
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
