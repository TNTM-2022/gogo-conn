package system_info

import (
	"encoding/json"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"gogo-connector/components/config"
	"log"
	"os"
	"runtime"
)

type sysInfoRespBody struct {
	Hostname          string            `json:"hostname"`
	OsType            string            `json:"type"` // os type
	Platform          string            `json:"platform"`
	Arch              string            `json:"arch"`
	Release           string            `json:"release"`
	Uptime            float64           `json:"uptime"` // 进程
	Loadavg           [3]float64        `json:"loadavg"`
	Totalmem          uint64            `json:"totalmem"`
	Freemem           uint64            `json:"freemem"`
	Cpus              int               `json:"cpus"`
	NetworkInterfaces string            `json:"networkInterfaces"`
	Versions          map[string]string `json:"versions"`
}
type sysInfoResp struct {
	ServerId string          `json:"serverId"`
	Body     sysInfoRespBody `json:"body"`
}

var platform, _, version, _ = host.PlatformInformation()
var hostname, _ = os.Hostname()

func MointorHandler() (req, respBody, respErr, notify []byte) {
	l, _ := load.Avg()
	m, _ := mem.VirtualMemory()
	notify, e := json.Marshal(sysInfoResp{
		ServerId: *config.ServerID,
		Body: sysInfoRespBody{
			Hostname:          hostname,
			OsType:            runtime.GOOS,
			Platform:          platform,
			Arch:              runtime.GOARCH,
			Release:           version,
			Uptime:            config.Uptime(),
			Loadavg:           [3]float64{l.Load1, l.Load5, l.Load15},
			Totalmem:          m.Total,
			Freemem:           m.Free,
			Cpus:              runtime.NumCPU(),
			NetworkInterfaces: "",
		},
	})
	if e != nil {
		log.Fatal(e)
	}
	return
}
