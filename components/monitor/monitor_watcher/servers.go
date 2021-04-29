package monitor_watcher

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	concurrentMap "github.com/orcaman/concurrent-map"
	"go-connector/components/monitor/types"
	config "go-connector/config"
	"go-connector/global"
	"go-connector/libs/mqtt_client"
	"go-connector/libs/package_coder"
	"go-connector/logger"
	"log"
	"time"
)

type BackendMsg = package_coder.BackendMsg
type PkgBelong = package_coder.PkgBelong

//var serverOptLocker sync.Mutex

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
			fmt.Println("remove server .............")
			removeServers(ss.ServerID)
			respErr = json.RawMessage(`1`)
			respBody = json.RawMessage(`1`)
		}
	case "replaceServer":
		fmt.Println("replace server .............")
		respErr = json.RawMessage(`1`)
		//respBody = json.RawMessage(`1`)
	case "startOver":
		fmt.Println("start over .............")
		respErr = json.RawMessage(`1`)

	}
	return
}
func removeServers(serverId string) {
	fmt.Println("remove server ", serverId)
}
func ConnectToServer(serv types.RegisterInfo) {
	if *config.ServerType == serv.ServerType {
		log.Println("skip init same type server", serv.ServerID)
		return
	}

	fmt.Println(serv)

	client := mqtt_client.CreateMQTTClient(&mqtt_client.MQTT{
		Host:     serv.Host,
		Port:     fmt.Sprintf("%v", serv.Port),
		ClientID: serv.ServerID,
		//ServerType: serv.ServerType,
	})

	client.SetCallbacks(nil, func(c paho.Client, msg paho.Message) {
		OnPublishHandler(client, c, msg)
	})
	client.Start()
	//defer client.Stop()
	log.Println("链接服务器", serv.ServerID, serv.ServerType, serv.ServerID, serv.Host, serv.Port, client.IsConnected())

	// 初始化 serverType：chan  serverType：serverId：serverInfo
	global.RemoteBackendTypeForwardChan.SetIfAbsent(serv.ServerType, make(chan package_coder.BackendMsg, 10000))
	global.RemoteBackendClients.SetIfAbsent(serv.ServerType, concurrentMap.New())
	_cmp, _ := global.RemoteBackendClients.Get(serv.ServerType) // 一定存在
	cmp, _ := _cmp.(concurrentMap.ConcurrentMap)
	cmp.Upsert(serv.ServerID, client, func(exists bool, oldV, newV interface{}) interface{} {
		if v, ok := oldV.(*mqtt_client.MQTT); exists && ok {
			v.Stop()
			fmt.Println("关闭？？？？")
			// todo 停止消息转发， 然后再停止server client
		}
		return newV
	})

	go func(s types.RegisterInfo, client *mqtt_client.MQTT) {
		_forwardChan, ok := global.RemoteBackendTypeForwardChan.Get(s.ServerType)
		if !ok {
			log.Println("no found server in store.")
		}
		forwardChan, _ := _forwardChan.(chan package_coder.BackendMsg)

		for msg := range forwardChan {

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
			log.Println("---------------")
			Request(client, "rpc", "", pkgId, p, &PkgBelong{
				SID:         msg.Sid,
				StartAt:     time.Now(),
				ClientPkgID: msg.PkgID,
				Route:       msg.Route,
			})
			log.Println("===============")
			log.Println("rpc send ok", client.ClientID)
		}
	}(serv, client)

	//time.Sleep(time.Second * 100)
	//return nil
	//todo 卡住 不要结束
}
