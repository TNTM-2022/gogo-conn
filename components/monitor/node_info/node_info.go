package node_info

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"go-connector/config"
	"go-connector/logger"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"time"
)

type nodeInfoRespBody struct {
	Time       string `json:"time"`
	ServerId   string `json:"serverId"`
	ServerType string `json:"serverType"`
	Pid        string `json:"pid"`
	CpuAvg     string `json:"cpuAvg"`
	MemAvg     string `json:"memAvg"`
	Vsz        uint64 `json:"vsz"`
	Rss        uint64 `json:"rss"`
	Usr        int    `json:"usr"`
	Sys        int    `json:"sys"`
	Gue        int    `json:"gue"`
}
type nodeInfoResp struct {
	ServerId string           `json:"serverId"`
	Body     nodeInfoRespBody `json:"body"`
}

func formatTime() string {
	now := time.Now()
	h, m, s := now.Clock()
	ap := "AM"
	if h > 12 {
		ap = "PM"
		h = h - 12
	}
	return fmt.Sprintf("%v-%v-%v %v:%v:%v %v", now.Year(), now.Month(), now.Day(), h, m, s, ap)
}

func MonitorHandler() (req, respBody, respErr, notify []byte) { // 进程的 ps 数据
	p, _ := process.NewProcess(int32(config.Pid))
	cpuAvg, _ := p.CPUPercent()
	memAvg, _ := p.MemoryPercent()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	n := nodeInfoResp{
		ServerId: *config.ServerID,
		Body: nodeInfoRespBody{
			Time:       formatTime(),
			ServerId:   *config.ServerID,
			ServerType: *config.ServerType,
			Pid:        strconv.FormatInt(int64(config.Pid), 10),
			CpuAvg:     fmt.Sprintf("%.2f", cpuAvg),
			MemAvg:     fmt.Sprintf("%.2f", memAvg),
			Rss:        memStats.HeapSys + memStats.StackSys, // 实际内存的大小
			Vsz:        0,
			Usr:        0,
			Sys:        0,
			Gue:        0,
		},
	}

	notify, e := json.Marshal(n)
	if e != nil {
		logger.ERROR.Println("json.marshal error", zap.Error(e))
	}
	return

}

/**
  let mytime = date.toLocaleTimeString();
  let mytimes = n + '-' + y + '-' + r + ' ' + mytime;
*/
