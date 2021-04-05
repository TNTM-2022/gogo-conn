package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	cfg "gogo-connector/components/config"
	"gogo-connector/components/global"
	"gogo-connector/libs/mqtt"
	"log"
	"os"
	"runtime"
	"sync"
)

type RegisterInfoRemoterPaths struct {
	Namespace  string `json:"namespace"`
	ServerType string `json:"serverType"`
	Path       string `json:"path"`
}
type RegisterInfo struct {
	Main       string `json:"main"`
	Env        string `json:"env"`
	ServerID   string `json:"id"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	ClientPort int32  `json:"clientPort"`
	Frontend   string `json:"frontend"`
	ServerType string `json:"serverType"`
	Token      string `json:"token"`
	PID        int32  `json:"pid"`

	RemotePaths  []RegisterInfoRemoterPaths `json:"remotePaths"`
	HandlerPaths []string                   `json:"handlerPaths"`
}

type Register struct {
	ServerID   string       `json:"id"`
	Type       string       `json:"type"`
	ServerType string       `json:"serverType"`
	PID        int32        `json:"pid"`
	Info       RegisterInfo `json:"info"`
	Token      string       `json:"token"`
}

type SubscribeBody struct {
	Action   string `json:"action"`
	ServerID string `json:"id"`
}
type Subscribe struct {
	ReqID    int64         `json:"reqId"`
	ModuleID string        `json:"moduleId"`
	Body     SubscribeBody `json:"body"`
}

type RegisterResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type MonitorBody struct {
	Signal   string       `json:"signal"`
	Action   string       `json:"action"`
	Server   RegisterInfo `json:"server"`
	ServerID string       `json:"id"`
}
type Monitor struct {
	RespId   int64       `json:"respId"`
	ReqID    int64       `json:"reqId"`
	ModuleID string      `json:"moduleId"`
	Body     MonitorBody `json:"body"`
	Command  string      `json:"command"`
}

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

type MonitListInfoRes struct {
	RespID int64         `json:"respId"`
	Error  MonitListInfo `json:"error"`
}

type ClientActionRes struct {
	RespId int64 `json:"respId"`
	Error  int32 `json:"error"`
}

type MonitorServers map[string]RegisterInfo
type MonitorAllServer struct {
	RespId   int64          `json:"respId"`
	ReqID    int64          `json:"reqId"`
	ModuleID string         `json:"moduleId"`
	Body     MonitorServers `json:"body"`
	Command  string         `json:"command"`
}

type MonitRespOk struct {
	RespID int64       `json:"respId"`
	Body   MonitorBody `json:"body"`
}

type onlineUserResp struct {
	ServerId       string `json:"serverId"`
	TotalConnCount int    `json:"totalConnCount"`
	LoginedCount   int    `json:"loginedCount"`
	loginedList    []UserReq
}

func DecodeMonitorAllServer(d []byte) MonitorAllServer {
	var mm MonitorAllServer
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		fmt.Println(mm, e)
	}
	return mm
}

var QuitCtx, QuitFn = context.WithCancel(context.Background())

func RegisterServer(ctx context.Context, cancelFn context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()

	mqttClient := mqtt.CreateMQTTClient(&mqtt.MQTT{
		Host:            "127.0.0.1",
		Port:            "3005",
		ClientID:        "clientId-1",
		SubscriptionQos: 1,
		Persistent:      true,
		Order:           false,
		KeepAliveSec:    5,
		PingTimeoutSec:  10,

		OnConnectCb: regServer,
		OnPublishCb: publishCb,
	})
	mqttClient.Start()

	<-ctx.Done()
	// 客户端退出
}

func regServer(mqttClient *mqtt.MQTT) {
	// 注册server， 携带 token
	m, _ := os.Getwd()
	regInfo := Register{
		ServerID:   *cfg.ServerID,
		Type:       "monitor",
		ServerType: *cfg.ServerType,
		PID:        int32(cfg.Pid),
		Info: RegisterInfo{
			Main:       m,
			Env:        *cfg.Env,
			ServerID:   *cfg.ServerID,
			Host:       *cfg.MqttServerHost, // mqtt server host
			Port:       int32(0),            // mqtt server port
			ClientPort: int32(0),
			Frontend:   "true",
			ServerType: "connector",
			Token:      "ok",
		},
		Token: "ok",
	}
	regStr, _ := json.Marshal(regInfo)
	mqttClient.Publish("register", regStr, 0)
	fmt.Println(string(regStr))
	log.Println("+++ monitor registed")

	subServer := Subscribe{
		ReqID:    1,
		ModuleID: "__masterwatcher__",
		Body: SubscribeBody{
			Action:   "subscribe",
			ServerID: *cfg.ServerID,
		},
	}
	subStr, _ := json.Marshal(subServer)
	mqttClient.Publish("monitor", subStr, 0)

	fmt.Println(string(regStr), string(subStr))

	log.Println("+++ monitor start monitor")

}

func publishCb(mqttClient paho.Client, m paho.Message) {
	fmt.Println("publish cb ", string(m.Payload()))
	switch m.Topic() {
	case "register":
		handleRegisterTopic(m)
	case "monitor":
		handleMonitorTopic(m, mqttClient)
	case "connect":
		{
			fmt.Println("connect")
		}
	default:
		{
			fmt.Println("unhandled Topic++++>>>>", m.Topic(), string(m.Payload()))
		}
	}

}

func DecodeMonitor(d []byte) Monitor {
	var mm Monitor
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		fmt.Println(mm, e)
	}
	return mm
}

// =================================
// todo 重连 注册问题
func handleRegisterTopic(m paho.Message) {
	log.Println("-- server registed to master")
	var res RegisterResp
	e := json.Unmarshal(m.Payload(), &res)
	if e != nil {
		fmt.Println(e)
	}
	if res.Msg != "ok" {
		fmt.Println("register >> quit >>", res.Msg)
		QuitFn()
		// todo 这边不应该这么实现
	}
}

func handleMonitorTopic(m paho.Message, mqttClient paho.Client) {
	monit := DecodeMonitor(m.Payload())
	var resp []byte = nil
	log.Println(monit.Body, " --- monit.Signal", monit.Body.Signal, " --- monit.Action", monit.Body.Action, "--- monit.Command", monit.Command)

	if monit.Command != "" {
		fmt.Println("not support command")
		return
	}

	// todo 确认
	//if monit.RespId != 0 {
	//	fmt.Println("unsupport resp", monit.RespId);
	//	//return;
	//}

	if monit.Body.Signal != "" {
		switch monit.Body.Signal {
		case "list":
			{
				resp = handleListSignal(monit)
			}
		case "stop", "kill":
			{
				QuitFn()
			}
		default:
			{
				fmt.Println("unhandled Monitor Signal: ", m.Topic(), string(m.Payload()))
			}
		}
	} else if monit.Body.Action != "" {
		switch monit.Body.Action {
		case "addServer":
			{
				AddServer(&monit.Body.Server)
				fmt.Println("------- ........ ADD")
			}
		case "removeServer":
			{
				fmt.Println("------- ........ Remove")
				RemoveServer(monit.Body.ServerID)
			}
		default:
			{
				if monit.Body.Action != "" {
					fmt.Println("unhandled Monitor Action: ", m.Topic(), string(m.Payload()))
					return
				}
			}
		}

		if monit.ReqID == 0 {
			fmt.Println(monit.ReqID, "  ReqID")
			return
		}

		//rr := &ClientActionRes{
		//	RespId: monit.ReqID,
		//	Error:  1,
		//}

		rr := &MonitRespOk{
			RespID: monit.ReqID,
			Body: MonitorBody{
				Action: monit.Body.Action,
			},
		}
		var e error
		if resp, e = json.Marshal(rr); e != nil {
			fmt.Println(e)
		}
	} else if monit.ModuleID == "" && monit.RespId > 0 {
		monitAllServer := DecodeMonitorAllServer(m.Payload())
		AddServers(monitAllServer.Body)
		fmt.Println("add servers ** ", monitAllServer.Body)

	} else if monit.ModuleID == "onlineUser" {
		//fmt.Println("----- ....... OnlineUser")
		resp = reportOnlineUser(monit.Body.ServerID)
	} else {
		fmt.Println("unhandled Monit++++>>>>", m.Topic(), string(m.Payload()))
	}

	//		if (protocol.isRequest(msg)) {
	//			let resp = protocol.composeResponse(msg, err, res);
	//			if (resp) {
	//				self.doSend('monitor', resp);
	//			}
	//		}
	//		else {
	//			// notify should not have a callback
	//			logger.error('notify should not have a callback.');
	//		}
	//	});
	//}
	//

	//if monit.ReqID == 0 && resp != nil {
	//	fmt.Println("no need resp to master")
	//	return
	//}
	if resp == nil {
		fmt.Println("empty resp ")
		return
	}

	mqttClient.Publish("monitor", 0, false, resp)

}

func handleListSignal(monit Monitor) []byte {
	monitInf := MonitListInfo{
		ServerID: *cfg.ServerID,
		Body: MonitListInfoBody{
			ServerID:   *cfg.ServerID,
			ServerType: *cfg.ServerType,
			Pid:        cfg.Pid,
			Uptime:     cfg.Uptime(),
		},
	}

	//if proc, er := top.NewProcess(int32(cfg.Pid)); er == nil {
	//	if memInf, e := proc.MemoryInfo(); e == nil {
	//		monitInf.Body.RSS = memInf.RSS / (1024 * 1024)
	//	}
	//}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	monitInf.Body.HeapTotal = memStats.TotalAlloc / (1024 * 1024)
	monitInf.Body.HeapUsed = memStats.HeapInuse / (1024 * 1024)

	res := MonitListInfoRes{
		RespID: monit.ReqID,
		Error:  monitInf,
	}
	if monit.ReqID != 0 {
		if resp, e := json.Marshal(res); e != nil {
			fmt.Println(e)
		} else {
			return resp
		}
	}

	return nil
}

func reportOnlineUser(serverId string) []byte {
	res := onlineUserResp{
		LoginedCount:   len(global.Users),
		TotalConnCount: len(global.Sids),
		ServerId:       serverId,
	}
	if resp, err := json.Marshal(res); err != nil {
		fmt.Println(err)
	} else {
		return resp
	}
	return nil
}
