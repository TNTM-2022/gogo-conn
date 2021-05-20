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
	"go.uber.org/zap"
	"os"
	"sync"
)

func StartMonitServer(ctx context.Context, cancelFn context.CancelFunc, wg *sync.WaitGroup) {
	//defer wg.Done()

	client := mqtt.CreateMQTTClient(&mqtt.MQTT{
		Host:            cfg.MasterHost,
		Port:            cfg.MasterPort,
		ClientID:        fmt.Sprintf("monitor-%v", *cfg.ServerID),
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
	go client.Start()

	//<-ctx.Done()
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
		logger.DEBUG.Println("monitor,doRegister", "first reg server to master")
	case true:
		//regInfo.Token = ""
		//regStr, _ := json.Marshal(regInfo)
		reconnectCb(mqttClient, regStr)
		logger.DEBUG.Println("monitor,doRegister", "redo reg server to master")
	}
}

func firstConnectCb(mqttClient *mqtt.MQTT, regStr []byte) {
	// 注册server， 携带 token
	mqttClient.Publish("register", regStr, 0, false) // 直接发送 lib/monitor/monitorAgent line 151
	logger.DEBUG.Println("monitor,doRegister", "reg server to master")

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
			logger.ERROR.Println("failed to send request to master", zap.String("error", err))
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
	})

	logger.DEBUG.Println("monitor,doRegister", "monitor started", zap.Strings("reg info", []string{string(regStr), string(subStr)}))
}
func reconnectCb(mqttClient *mqtt.MQTT, regStr []byte) {
	mqttClient.Publish("reconnect", regStr, 0, false) // 直接发送 lib/monitor/monitorAgent line 151
	logger.DEBUG.Println("monitor,doRegister", "monitor reg done after reconnection")

}

func onPublishCb(mqttClient *mqtt.MQTT, m paho.Message) {
	logger.DEBUG.Println("monitor", "handle publish cb", zap.String("topic", m.Topic()), zap.String("payload", string(m.Payload())))
	switch m.Topic() {
	case "register":
		handleRegisterTopic(m)
	case "monitor":
		handleMonitorTopic(mqttClient, m)
	case "connect":
		{
			logger.DEBUG.Println("monitor", "connect")
		}
	case "reconnect_ok":
		{
			logger.DEBUG.Println("monitor", "reconnected")
		}
	default:
		{
			logger.ERROR.Println("unhandled Topic", zap.String("topic", m.Topic()), zap.String("payload", string(m.Payload())))
		}
	}

}

// todo 多 master 机制
func handleRegisterTopic(m paho.Message) {
	logger.DEBUG.Println("monitor", "monitor server reg to master")
	var res RegisterResp
	e := json.Unmarshal(m.Payload(), &res)
	if e != nil {
		logger.ERROR.Println("parse reg response failed", zap.Error(e))
	}
	if res.Msg != "ok" {
		logger.DEBUG.Println("register >> quit >>", res.Msg)
		//QuitFn()
		// todo 这边不应该这么实现
	}
}

func handleMonitorTopic(mqttClient *mqtt.MQTT, m paho.Message) {
	monit := DecodeMonitor(m.Payload())
	ignoreModuleLog := map[string]bool{
		"onlineUser":  false,
		"systemInfo":  false,
		"__console__": false,
		"nodeInfo":    false,
	}
	if ignoreModuleLog[monit.ModuleID] {
		logger.DEBUG.Println("monitor", fmt.Sprintf("monit.Signal:%v; monit.Action:%v; monit.Command:%v;", monit.Body.Signal, monit.Body.Action, monit.Command), zap.String("payload", string(m.Payload())))
	}
	if monit.Command != "" {
		logger.ERROR.Println("not support command", zap.String("command", monit.Command))
		return
	}
	if monit.RespId > 0 {
		logger.ERROR.Println("not support respId>0", zap.Int64("respId", monit.RespId))
		return
	}
	var req, respErr, respBody, notify []byte
	switch monit.ModuleID {
	case "__console__":
		{
			req, respBody, respErr, notify = console.MonitorHandler(monit.Body.Signal, global.QuitFn, monit.Body.BlackList)
		}

	case "__monitorwatcher__":
		{
			req, respBody, respErr, notify = monitor_watcher.MonitorHandler(monit.Body.Action, monit.Body)
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
			logger.DEBUG.Println("monitor", "profiling coming soon")
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
			logger.ERROR.Println("receive unknow moduleId", zap.String("moduleId", monit.ModuleID), zap.String("payload", string(m.Payload())))
		}
	}

	if req != nil { // 应该不存在
		Request(mqttClient, "monitor", monit.ModuleID, req, func(err string, data []byte) {
			logger.INFO.Println("get a response after request to master", zap.String("error", err), zap.String("data", string(data)))
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
		logger.ERROR.Println("decode server monitor info array failed", zap.Error(e))
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
		logger.ERROR.Println("decode server monitor info failed", zap.Error(e))
	}
	return mm
}
