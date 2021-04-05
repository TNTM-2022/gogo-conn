package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"go-connector/interfaces"
	"log"
	"os"
	"runtime"
	"time"

	cfg "go-connector/config"
	//proto "github.com/huin/mqtt"
	//"github.com/jeffallen/mqtt"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	top "github.com/shirou/gopsutil/process"
)

// RequestRemote

var quit = make(chan struct{})

func endQuit(q chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	close(q)
}

func Monit(mc interfaces.MainControl, fn context.CancelFunc) {
	defer func() {
		log.Println("stopping monitor")
		fn()
	}()
	addr := fmt.Sprintf("%s:%d", *cfg.MonitorHost, *cfg.MonitorPort)
	fmt.Println(addr)
	log.Println("monitor connect ", addr)
	opt := mqtt.NewClientOptions()
	opt.SetDefaultPublishHandler(HandlePublish)
	opt.AddBroker("tcp://" + addr)
	mqttClient := mqtt.NewClient(opt)
	defer mqttClient.Disconnect(0)
	token := mqttClient.Connect()
	if !token.WaitTimeout(2 * time.Second) {
		log.Println("mqtt monitor timeout")
		return
	}
	log.Println("+++ connected mqtt monitor")

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
			Host:       *cfg.Host,
			Port:       int32(cfg.Port),
			ClientPort: int32(*cfg.ClientPort),
			Frontend:   "true",
			ServerType: "connector",
			Token:      "ok",
		},
	}
	reg, _ := json.Marshal(regInfo)
	fmt.Println(string(reg))
	mqttClient.Publish("register", 0, false, reg)
	log.Println("+++ monitor registed")

	subServer := Subscribe{
		ReqID:    1,
		ModuleID: "__masterwatcher__",
		Body: SubscribeBody{
			Action:   "subscribe",
			ServerID: *cfg.ServerID,
		},
	}
	ss, _ := json.Marshal(subServer)
	mqttClient.Publish("monitor", 0, false, ss)
	fmt.Println(string(reg), string(ss))

	log.Println("+++ monitor start monitor")

	select {
	case <-quit:
		{
			fn()
		}
	case <-mc.Ctx.Done():
		{
			endQuit(quit)
		}
	}
	return
}

func HandlePublish(mqttClient mqtt.Client, m mqtt.Message) {
	switch m.Topic() {
	case "register":
		{
			log.Println("-- server registed to master")
			var res RegisterResp
			e := json.Unmarshal(m.Payload(), &res)
			if e != nil {
				fmt.Println(e)
			}
			if res.Msg != "ok" {
				fmt.Println("register >>>", res.Msg)
				endQuit(quit)
				// todo 这边不应该这么实现
			}
		}
	case "monitor":
		{
			monit := DecodeMonitor(m.Payload())
			var resp []byte = nil
			log.Println(" --- monit.Signal", monit.Body.Signal, " --- monit.Action", monit.Body.Action)

			if monit.Command != "" {
				fmt.Println("not support command")
				return
			}

			//if monit.RespId != 0 {
			//	fmt.Println("unsupport resp", monit.RespId);
			//	//return;
			//}

			if monit.Body.Signal != "" {
				switch monit.Body.Signal {
				case "list":
					{
						monitInf := MonitListInfo{
							ServerID: *cfg.ServerID,
							Body: MonitListInfoBody{
								ServerID:   *cfg.ServerID,
								ServerType: *cfg.ServerType,
								Pid:        cfg.Pid,
								Uptime:     cfg.Uptime(),
							},
						}

						if proc, er := top.NewProcess(int32(cfg.Pid)); er == nil {
							if memInf, e := proc.MemoryInfo(); e == nil {
								monitInf.Body.RSS = memInf.RSS / (1024 * 1024)
							}
						}
						var memStats runtime.MemStats
						runtime.ReadMemStats(&memStats)
						monitInf.Body.HeapTotal = memStats.TotalAlloc / (1024 * 1024)
						monitInf.Body.HeapUsed = memStats.HeapInuse / (1024 * 1024)

						res := MonitListInfoRes{
							RespID: monit.ReqID,
							Error:  monitInf,
						}
						if monit.ReqID != 0 {
							var e error
							if resp, e = json.Marshal(res); e != nil {
								fmt.Println(e)
							}
						}
					}
				case "stop", "kill":
					{
						endQuit(quit)
					}
				default:
					{
						fmt.Println("unhandled Monitor++++>>>>", m.Topic, string(m.Payload()))
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
							fmt.Println("unhandled Action++++>>>>", m.Topic, string(m.Payload()))
							return
						}
					}
				}

				if monit.ReqID == 0 {
					fmt.Println(monit.ReqID, "  ReqID")
					return
				}

				rr := &ClientActionRes{
					RespId: monit.ReqID,
					Error:  1,
				}

				//rr := &MonitRespOk{
				//	RespID:monit.ReqID,
				//	Body: MonitorBody{
				//		Action: monit.Body.Action,
				//	},
				//}
				var e error
				if resp, e = json.Marshal(rr); e != nil {
					fmt.Println(e)
				}
			} else if monit.ModuleID == "" && monit.RespId > 0 {
				monitAllServer := DecodeMonitorAllServer(m.Payload())
				AddServers(monitAllServer.Body)

			} else {
				fmt.Println("unhandled Monit++++>>>>", m.Topic, string(m.Payload()))
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

			if monit.ReqID == 0 && resp != nil {
				fmt.Println("no need resp to master")
				return
			}
			if resp == nil {
				fmt.Println("empty resp ")
				return
			}

			mqttClient.Publish("monitor", 0, false, resp)

		}
	case "connect":
		{
			fmt.Println("connect")
		}
	default:
		{
			fmt.Println("unhandled Topic++++>>>>", m.Topic, string(m.Payload()))
		}
	}

}
