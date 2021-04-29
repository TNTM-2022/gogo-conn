package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/components/monitor/console"
	"go-connector/components/monitor/monitor_watcher"
	"go-connector/components/monitor/node_info"
	"go-connector/components/monitor/online_user"
	"go-connector/components/monitor/system_info"
	"go-connector/components/monitor/types"
	cfg "go-connector/config"
	"go-connector/global"
	mqtt "go-connector/libs/mqtt_client"
	"go-connector/logger"
	"log"
	"os"
	"sync"
)

func StartMonitServer(ctx context.Context, cancelFn context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	client := mqtt.CreateMQTTClient(&mqtt.MQTT{
		Host:            "127.0.0.1",
		Port:            "3005",
		ClientID:        "clientId-1",
		SubscriptionQos: 1,
		Persistent:      true,
		Order:           true,
		KeepAliveSec:    10,
		PingTimeoutSec:  30,
	})
	client.SetCallbacks(doRegisterServer, func(c paho.Client, msg paho.Message) {
		if !OnPublishHandler(client, c, msg) {
			onPublishCb(client, msg)
		}
	})
	client.Start()

	<-ctx.Done()
	// 客户端退出
}

type RegisterInfo = types.RegisterInfo
type RegisterInfoRemoterPaths = types.RegisterInfoRemoterPaths

type Register = struct {
	ServerID   string       `json:"id"`
	Type       string       `json:"type"`
	ServerType string       `json:"serverType"`
	PID        int32        `json:"pid"`
	Info       RegisterInfo `json:"info"`
	Token      string       `json:"token"`
}

type MonitorServers map[string]RegisterInfo

type RegisterResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type MonitorBody struct {
	Signal    string       `json:"signal"`
	Action    string       `json:"action"`
	Server    RegisterInfo `json:"server"`
	ServerID  string       `json:"id"`
	BlackList []string     `json:"blacklist"`
}

func doRegisterServer(mqttClient *mqtt.MQTT) {
	m, _ := os.Getwd()
	regInfo := Register{
		ServerID:   *cfg.ServerID,
		Type:       "monitor",
		ServerType: *cfg.ServerType,
		PID:        int32(cfg.Pid),
		Info: RegisterInfo{
			Main:     m,
			Env:      *cfg.Env,
			ServerID: *cfg.ServerID,
			Host:     *cfg.MqttServerHost, // mqtt server host
			Port:     cfg.MqttServerPort,  // mqtt server port
			//ClientPort:   cfg.MqttServerPort,  // ws server port
			Frontend:     "true",
			ServerType:   "connector",
			Token:        "",
			RemotePaths:  make([]RegisterInfoRemoterPaths, 1),
			HandlerPaths: make([]string, 1),
		},
		Token: "",
	}

	regStr, _ := json.Marshal(regInfo)
	switch mqttClient.IsReconnect() {
	case false:
		firstConnectCb(mqttClient, regStr)
		fmt.Println("first connect")
	case true:
		//regInfo.Token = ""
		//regStr, _ := json.Marshal(regInfo)
		reconnectCb(mqttClient, regStr)
		fmt.Println("re connect")

	}
}

func firstConnectCb(mqttClient *mqtt.MQTT, regStr []byte) {
	// 注册server， 携带 token
	mqttClient.Publish("register", regStr, 0, false) // 直接发送 lib/monitor/monitorAgent line 151
	log.Println("monitor client regist done")

	subServer := struct {
		Action   string `json:"action"`
		ServerID string `json:"id"`
	}{
		Action:   "subscribe",
		ServerID: *cfg.ServerID,
	}
	subStr, _ := json.Marshal(subServer)
	Request(mqttClient, "monitor", "__masterwatcher__", subStr, func(err string, data []byte) {
		if err != "" {
			fmt.Println("????", err)
		}
		monitAllServerMap := DecodeAllServerMonitorInfo(data)
		serv := make([]RegisterInfo, 0, len(monitAllServerMap))
		for i, v := range monitAllServerMap {
			if i != *cfg.ServerID {
				serv = append(serv, v)
			}
		}
		for _, s := range serv {
			monitor_watcher.ConnectToServer(s)
		}
		//mointor_watcher.AddServers(serv) // todo 添加server
		logger.DEBUG.Println("???", serv)
	})

	logger.INFO.Println(string(regStr), string(subStr))

	logger.DEBUG.Println("+++ monitor start monitor")
}
func reconnectCb(mqttClient *mqtt.MQTT, regStr []byte) {
	mqttClient.Publish("reconnect", regStr, 0, false) // 直接发送 lib/monitor/monitorAgent line 151
	logger.DEBUG.Println("monitor registed")

}

func onPublishCb(mqttClient *mqtt.MQTT, m paho.Message) {
	logger.INFO.Println("<<< publish cb ", m.Topic(), string(m.Payload()))
	switch m.Topic() {
	case "register":
		handleRegisterTopic(m)
	case "monitor":
		handleMonitorTopic(mqttClient, m)
	case "connect":
		{
			logger.DEBUG.Println("connect")
		}
	case "reconnect_ok":
		{
			logger.DEBUG.Println("reconnected")
		}
	default:
		{
			logger.DEBUG.Println("unhandled Topic++++", m.Topic(), string(m.Payload()))
		}
	}

}

// todo 多 master 机制
func handleRegisterTopic(m paho.Message) {
	log.Println("monitor server registed to master")
	var res RegisterResp
	e := json.Unmarshal(m.Payload(), &res)
	if e != nil {
		logger.ERROR.Println(e)
	}
	if res.Msg != "ok" {
		logger.DEBUG.Println("register >> quit >>", res.Msg)
		//QuitFn()
		// todo 这边不应该这么实现
	}
}

func handleMonitorTopic(mqttClient *mqtt.MQTT, m paho.Message) {
	monit := DecodeMonitor(m.Payload())
	// ignoreModuleLog:= make()
	ignoreModuleLog := map[string]bool{
		"onlineUser":  false,
		"systemInfo":  false,
		"__console__": false,
		"nodeInfo":    false,
	}
	if ignoreModuleLog[monit.ModuleID] {
		log.Println(">>> monit.Signal", monit.Body.Signal, " >>> monit.Action", monit.Body.Action, ">>> monit.Command", monit.Command, string(m.Payload()))
	}
	if monit.Command != "" {
		logger.ERROR.Println("not support command", monit.Command)
		return
	}
	if monit.RespId > 0 {

	}
	var req, respErr, respBody, notify []byte
	switch monit.ModuleID {
	case "__console__":
		{
			req, respBody, respErr, notify = console.MonitorHandler(monit.Body.Signal, global.QuitFn, monit.Body.BlackList)
		}

	case "__monitorwatcher__":
		{
			req, respBody, respErr, notify = monitor_watcher.MonitorHandler(monit.Body.Action, &monit.Body)
		}
	//case "RestartNotifyModule":
	//	{
	//
	//	}
	//case "watchServer":
	//	{
	//
	//	}

	case "onlineUser":
		{
			req, respBody, respErr, notify = online_user.MointorHandler(monit.Body.ServerID)
		}
	case "monitorLog":
		{
			//req, respBody, respErr, notify = monitor_log.MonitorHandler()
		}
	case "profiler":
		{
			logger.DEBUG.Println("profiling coming")
			return
		}
	case "scripts":
		{
			//req, respBody, respErr, notify =
		}
	case "nodeInfo":
		{
			req, respBody, respErr, notify = node_info.MonitorHandler()
		}
	case "systemInfo":
		{
			req, respBody, respErr, notify = system_info.MointorHandler()
		}
	default:
		{
			log.Printf(" *************    receive unknow moduleId: %v, %v", monit.ModuleID, string(m.Payload()))
		}
	}

	if req != nil { // todo 应该不存在
		Request(mqttClient, "monitor", monit.ModuleID, req, func(err string, data []byte) {
			log.Println("request get a response", string(err), string(data))
		})
	} else if notify != nil {
		Notify(mqttClient, "monitor", monit.ModuleID, notify)
	} else if respBody != nil || respErr != nil {
		Response(mqttClient, "monitor", monit.ReqID, respErr, respBody)
	}
}

func DecodeAllServerMonitorInfo(d []byte) MonitorServers {
	var mm MonitorServers
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		logger.ERROR.Println(mm, e)
	}
	return mm
}

func DecodeMonitor(d []byte) types.Monitor {
	var mm types.Monitor
	var ss string
	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		logger.ERROR.Println(mm, e)
	}
	return mm
}
