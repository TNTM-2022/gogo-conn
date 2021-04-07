package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	cfg "gogo-connector/components/config"
	"gogo-connector/components/monitor/console"
	"gogo-connector/components/monitor/mointor_watcher"
	"gogo-connector/components/monitor/monitor_log"
	"gogo-connector/components/monitor/node_info"
	"gogo-connector/components/monitor/online_user"
	"gogo-connector/components/monitor/system_info"
	"gogo-connector/components/monitor/types"
	"gogo-connector/libs/mqtt"
	"log"
	"os"
	"sync"
)

type Register struct {
	ServerID   string             `json:"id"`
	Type       string             `json:"type"`
	ServerType string             `json:"serverType"`
	PID        int32              `json:"pid"`
	Info       types.RegisterInfo `json:"info"`
	Token      string             `json:"token"`
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

type ClientActionRes struct {
	RespId int64 `json:"respId"`
	Error  int32 `json:"error"`
}

type MonitorServers map[string]types.RegisterInfo

//type MonitorAllServer struct {
//	RespId   int64          `json:"respId"`
//	ReqID    int64          `json:"reqId"`
//	ModuleID string         `json:"moduleId"`
//	Body     MonitorServers `json:"body"`
//	Command  string         `json:"command"`
//}

type MonitRespOk struct {
	RespID int64 `json:"respId"`
	Body   int32 `json:"body"`
	Error  int32 `json:"error"`
}

var QuitCtx, QuitFn = context.WithCancel(context.Background()) // graceful shutdown

type Client struct {
	reqId int64
}

func MonitServer(ctx context.Context, cancelFn context.CancelFunc, wg *sync.WaitGroup) {
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

		OnConnectCb: regServerCb,
		OnPublishCb: publishCb,
	})
	mqttClient.Start()

	<-ctx.Done()
	// 客户端退出
}

func regServerCb(mqttClient *mqtt.MQTT) {
	// 注册server， 携带 token
	m, _ := os.Getwd()
	regInfo := Register{
		ServerID:   *cfg.ServerID,
		Type:       "monitor",
		ServerType: *cfg.ServerType,
		PID:        int32(cfg.Pid),
		Info: types.RegisterInfo{
			Main:         m,
			Env:          *cfg.Env,
			ServerID:     *cfg.ServerID,
			Host:         *cfg.MqttServerHost, // mqtt server host
			Port:         int32(0),            // mqtt server port
			ClientPort:   int32(0),            // ws server port
			Frontend:     "true",
			ServerType:   "connector",
			Token:        "ok",
			RemotePaths:  make([]types.RegisterInfoRemoterPaths, 1),
			HandlerPaths: make([]string, 1),
		},
		Token: "ok",
	}
	regStr, _ := json.Marshal(regInfo)
	mqttClient.Publish("register", regStr, 0, false) // 直接发送 lib/monitor/monitorAgent line 151
	log.Println("monitor registed")

	subServer := SubscribeBody{
		Action:   "subscribe",
		ServerID: *cfg.ServerID,
	}
	subStr, _ := json.Marshal(subServer)
	//mqttClient.Publish("monitor", subStr, 0, true)
	mqttClient.Request("monitor", "__masterwatcher__", subStr, func(err string, data []byte) {
		//  if err == ""
		monitAllServerMap := DecodeMonitorAllServer(data)
		serv := make([]types.RegisterInfo, 0, len(monitAllServerMap))
		for i, v := range monitAllServerMap {
			if i != *cfg.ServerID {
				serv = append(serv, v)
			}
		}
		mointor_watcher.AddServers(serv)
	})

	fmt.Println(string(regStr), string(subStr))

	log.Println("+++ monitor start monitor")
}

func publishCb(mqttClient *mqtt.MQTT, m paho.Message) {
	fmt.Println("<<< publish cb ", m.Topic(), string(m.Payload()))
	switch m.Topic() {
	case "register":
		handleRegisterTopic(m)
	case "monitor":
		handleMonitorTopic(mqttClient, m)
	case "connect":
		{
			fmt.Println("connect")
		}
	default:
		{
			fmt.Println("unhandled Topic++++", m.Topic(), string(m.Payload()))
		}
	}

}

func DecodeMonitor(d []byte) types.Monitor {
	var mm types.Monitor
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
// todo 多 master 机制
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

func handleMonitorTopic(mqttClient *mqtt.MQTT, m paho.Message) {
	monit := DecodeMonitor(m.Payload())
	if monit.ModuleID != "onlineUser" {
		log.Println(monit.Body, " --- monit.Signal", monit.Body.Signal, " --- monit.Action", monit.Body.Action, "--- monit.Command", monit.Command, string(m.Payload()))
	}

	if monit.Command != "" {
		fmt.Println("not support command", monit.Command)
		return
	}
	if monit.RespId > 0 {

	}
	var req, respErr, respBody, notify []byte
	switch monit.ModuleID {
	case "__console__":
		{
			req, respBody, respErr, notify = console.MonitorHandler(monit.Body.Signal, QuitFn, monit.Body.BlackList)
		}
	case "__monitorwatcher__":
		{
			req, respBody, respErr, notify = mointor_watcher.MonitorHandler(monit.Body.Action, &monit.Body)
		}
	case "onlineUser":
		{
			req, respBody, respErr, notify = online_user.MointorHandler(monit.Body.ServerID)
		}
	case "RestartNotifyModule":
		{

		}
	case "watchServer":
		{

		}
	case "monitorLog":
		{
			req, respBody, respErr, notify = monitor_log.MonitorHandler()
		}
	case "profiler":
		{
			fmt.Println("profiling coming")
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
			fmt.Println(" *************    receive unknow moduleId: %v, %v", monit.ModuleID, string(m.Payload()))
		}

	}

	if req != nil { // todo 应该不存在
		mqttClient.Request("monitor", monit.ModuleID, req, func(err string, data []byte) {
			log.Println("request get a response", string(err), string(data))
		})
	} else if notify != nil {
		mqttClient.Notify("monitor", monit.ModuleID, notify)
	} else if respBody != nil || respErr != nil {
		mqttClient.Response("monitor", monit.ReqID, respErr, respBody)
	}

	/*

		if monit.Body.Action != "" {
			switch monit.Body.Action {
			case "addServer":
				{
					mointor_watcher.AddServer(&monit.Body.Server)
					fmt.Println("------- ........ ADD")
				}
			case "removeServer":
				{
					fmt.Println("------- ........ Remove")
					mointor_watcher.RemoveServer(monit.Body.ServerID)
				}
			default:
				{
					if monit.Body.Action != "" {
						fmt.Println("unhandled Monitor Action: ", m.Topic(), string(m.Payload()))
						return
					}
				}
			}

			//if monit.ReqID == 0 {
			//	fmt.Println(monit.ReqID, "  ReqID")
			//	return
			//}

			//rr := &ClientActionRes{
			//	RespId: monit.ReqID,
			//	Error:  1,
			//}

			rr := &MonitRespOk{
				RespID: monit.ReqID,
				Body:   1,
				Error:  1, //  乱七八糟的用 注意返回值包裹在 body 还是 error 里
			}
			var e error
			if resp, e = json.Marshal(rr); e != nil {
				fmt.Println(e)
			}
		} else if monit.ModuleID == "" && monit.RespId > 0 {
			monitAllServer := DecodeMonitorAllServer(m.Payload())
			fmt.Println(string(m.Payload()))
			mointor_watcher.AddServers(monitAllServer.Body)
			fmt.Println("add servers ** ", monitAllServer.Body)
			rr := &MonitRespOk{
				RespID: monit.ReqID,
				Body:   1,
			}
			var e error
			if resp, e = json.Marshal(rr); e != nil {
				fmt.Println(e)
			}
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
	*/
}

/**
  if (self.state !== ST_REGISTERED) {
               return;
           }

           msg = protocol.parse(msg);

           if (msg.command) {
               // a command from master
               self.consoleService.command(msg.command, msg.moduleId, msg.body, function (err, res) {
                   // notify should not have a callback
               });
           } else {
               let respId = msg.respId;
               if (respId) {
                   // a response from monitor
                   let respCb = self.callbacks[respId];
                   if (!respCb) {
                       logger.warn('unknown resp id:' + respId);
                       return;
                   }
                   delete self.callbacks[respId];
                   respCb(msg.error, msg.body);
                   return;
               }

               // request from master
               self.consoleService.execute(msg.moduleId, 'monitorHandler', msg.body, function (err, res) {
                   if (protocol.isRequest(msg)) {
                       let resp = protocol.composeResponse(msg, err, res);
                       if (resp) {
                           self.doSend('monitor', resp);
                       }
                   } else {
                       // notify should not have a callback
                       logger.error('notify should not have a callback.');
                   }
               });
           }
*/

func DecodeMonitorAllServer(d []byte) MonitorServers {
	var mm MonitorServers
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		fmt.Println(mm, e)
	}
	return mm
}
