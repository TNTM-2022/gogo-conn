package monitor

import (
	"encoding/json"
	"fmt"
	cfg "gogo-connector/components/config"
	"gogo-connector/components/global"
	"log"
	"runtime"
)

type console struct{}

func (c *console) Stop() {
	QuitFn()
}

func (c *console) List(monit Monitor) []byte {
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

	res := MonitListInfoRes{
		RespID: monit.ReqID,
		Error:  monitInf,
	}
	resp, e := json.Marshal(res)
	if e != nil {
		log.Println(e)
		return nil
	}
	return resp
	//if monit.ReqID != 0 {
	//	if resp, e := json.Marshal(res); e != nil {
	//	} else {
	//		return resp
	//	}
	//}
	//
	//return nil
}

func (c *console) Kill() {
	c.Stop()
}

func (c *console) Blacklist(monit Monitor) {
	for _, v := range monit.Body.BlackList { // todo 不做ip 校验
		global.BlackList.Set(v, true)
	}
}

func (c *console) Restart() {
	fmt.Println("not implement yet. do it manually")
}

var Console = console{}
