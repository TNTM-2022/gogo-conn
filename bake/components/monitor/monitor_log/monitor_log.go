package monitor_log

import "log"

func MonitorHandler() (req, respBody, respErr, notify []byte) {
	log.Println("monitorLog no implement.")
	respBody = []byte("{\"logfile\": \"not implement\", \"dataArray\": []}")
	return
}
