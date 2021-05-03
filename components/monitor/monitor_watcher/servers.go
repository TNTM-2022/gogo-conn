package monitor_watcher

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"go-connector/components/monitor/types"
	config "go-connector/config"
	"go-connector/global"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
	"go-connector/logger"
	"log"
	"sync"
	"time"
)

type BackendMsg = package_coder.BackendMsg
type PkgBelong = package_coder.PkgBelong

//var serverOptLocker sync.Mutex

var servOptLocker sync.Mutex

func MonitorHandler(action string, ss *types.MonitorBody) (req, respBody, respErr, notify []byte) {
	switch action {
	case "addServer":
		{
			fmt.Println("add server .............")
			ConnectToServer(ss.Server)
			respErr = json.RawMessage(`1`)
		}
	case "removeServer":
		{
			fmt.Println(*ss)
			if ss.ServerID == "" {
				respErr = json.RawMessage(`0`)
				respBody = json.RawMessage(`0`)
				return
			}
			fmt.Println("remove server .............")
			removeServers(ss.ServerID)
			respErr = json.RawMessage(`1`)
			respBody = json.RawMessage(`1`)
		}
	case "replaceServer": // master 启动后， 重新同步信息； 防止出现 同名 但是不同端口 之类的事情发生
		fmt.Println("replace server .............")
		respErr = json.RawMessage(`1`)
		//respBody = json.RawMessage(`1`)
	case "startOver":
		fmt.Println("start over .............")
		respErr = json.RawMessage(`1`) // 全部启动了， 再发这个

	}
	return
}
func removeServers(serverId string) {
	//global.RemoteBackendTypeForwardChan.SetIfAbsent(serv.ServerType, make(chan package_coder.BackendMsg, 10000))
	//global.RemoteBackendClients.SetIfAbsent(serv.ServerType, concurrentMap.New())
	//_cmp, _ := global.RemoteBackendClients.Get(serv.ServerType) // 一定存在
	//cmp, _ := _cmp.(concurrentMap.ConcurrentMap)
	//cmp.Upsert(serv.ServerID, client, func(exists bool, oldV, newV interface{}) interface{} {
	//	if v, ok := oldV.(*mqtt_client.MQTT); exists && ok {
	//		v.Stop()
	//		fmt.Println("关闭？？？？")
	//		// todo 停止消息转发， 然后再停止server client
	//	}
	//	return newV
	//})
	var _serv interface{}
	if !global.RemoteBackendClients.RemoveCb(serverId, func(key string, v interface{}, exists bool) bool {
		if !exists {
			return false
		}
		_serv = v
		return true
	}) {
		return
	}
	serv := _serv.(*mqtt_client.MQTT)
	serv.Closing = true
	serv.SetReconnectCb(func(m *mqtt_client.MQTT) {
		m.Stop()
		servOptLocker.Lock() // todo 最好移到 client 关闭回调里做， 那样 可以很好控制消息转发； 目前 只是停止转发后端， 至于断开，是由后端自行决定
		defer servOptLocker.Unlock()
		serverTypeAllClosed := true
		global.RemoteBackendClients.IterCb(func(k string, v interface{}) {
			s := v.(*mqtt_client.MQTT)
			if s.ServerType == serv.ServerType {
				serverTypeAllClosed = false
			}
		})
		if serverTypeAllClosed {
			global.RemoteBackendTypeForwardChan.RemoveCb(serv.ServerType, func(key string, _ch interface{}, exists bool) bool {
				if !exists {
					return false
				}
				if ch, ok := _ch.(chan package_coder.BackendMsg); ok {
					close(ch)
					go func() {
						for range ch {
							// todo 增加全局拦截 返回路径不存在
						}
					}()
				}
				return true
			})
		}
		fmt.Println("服务器 不重连", serverTypeAllClosed)
	})
	fmt.Println("remove server ", serverId)
}
func ConnectToServer(serv types.RegisterInfo) {
	if *config.ServerType == serv.ServerType {
		log.Println("skip init same type server", serv.ServerID)
		return
	}

	fmt.Println("server =>", serv)

	client := mqtt_client.CreateMQTTClient(&mqtt_client.MQTT{
		Host:       serv.Host,
		Port:       fmt.Sprintf("%v", serv.Port),
		ClientID:   serv.ServerID,
		ServerType: serv.ServerType,
	})

	client.SetCallbacks(nil, func(c paho.Client, msg paho.Message) {
		OnPublishHandler(client, c, msg)
	})
	client.Start()
	//defer client.Stop()
	log.Println("链接服务器", serv.ServerID, serv.ServerType, serv.ServerID, serv.Host, serv.Port, client.IsConnected())

	// 初始化 serverType：chan  serverType：serverId：serverInfo
	// todo 资源回收  如果撤掉了 * servertype 通道需要关闭
	servOptLocker.Lock()
	global.RemoteBackendTypeForwardChan.SetIfAbsent(serv.ServerType, make(chan package_coder.BackendMsg, 10000)) // 不用回收
	//global.RemoteBackendClients.SetIfAbsent(serv.ServerType, concurrentMap.New())
	//_cmp, _ := global.RemoteBackendClients.Get(serv.ServerType) // 一定存在
	//cmp, _ := _cmp.(concurrentMap.ConcurrentMap)
	//cmp.Upsert(serv.ServerID, client, func(exists bool, oldV, newV interface{}) interface{} {
	global.RemoteBackendClients.Upsert(serv.ServerID, client, func(exists bool, oldV, newV interface{}) interface{} {
		if v, ok := oldV.(*mqtt_client.MQTT); exists && ok {
			go v.Stop()
			fmt.Println("关闭？？？？")
			// todo 停止消息转发， 然后再停止server client
		}
		return newV
	})
	servOptLocker.Unlock()

	go func(s types.RegisterInfo, client *mqtt_client.MQTT) {
		_forwardChan, ok := global.RemoteBackendTypeForwardChan.Get(s.ServerType)
		if !ok {
			log.Println("no found server in store.")
		}
		forwardChan, _ := _forwardChan.(chan package_coder.BackendMsg)

		for msg := range forwardChan {
			if client.Closing {
				forwardChan <- msg
				fmt.Println("remote closing")
				return
			}
			logger.DEBUG.Println(">>forward rpc to backend == ", s.Host, s.Port, s.ServerID, msg.ServerType)
			pkgId := client.GetReqId()
			p := package_coder.Encode(pkgId, &msg) // 后端 wrap 组装 session
			if p == nil {
				log.Println("encoding skip...")
				continue
			}

			if !client.IsConnected() { // 如果server 关闭了 消息要重新推回去
				fmt.Printf("client.IsConnected() = %v", false)
				forwardChan <- msg
				return
			}

			if !Request(client, "rpc", "", pkgId, p, &PkgBelong{
				SID:         msg.Sid,
				StartAt:     time.Now(),
				ClientPkgID: msg.PkgID,
				Route:       msg.Route,
			}) {
				forwardChan <- msg
				fmt.Println("remote closed.")
				return
			}

			logger.DEBUG.Println("rpc send ok", client.ClientID)
		}
	}(serv, client)

	//time.Sleep(time.Second * 100)
	//return nil
	//todo 卡住 不要结束
}
