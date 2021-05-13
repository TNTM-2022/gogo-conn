package monitor_log

import (
	"go-connector/logger"
)

func MonitorHandler() (req, respBody, respErr, notify []byte) {
	logger.INFO.Println("monitorLog no implement.")
	respBody = []byte(`{"logfile": "not implement", "dataArray": []}`)
	return
}
