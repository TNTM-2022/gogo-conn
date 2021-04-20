package mointor_watcher

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	config "gogo-connector/components/config"
	"gogo-connector/components/global"
	"gogo-connector/components/monitor/types"
	"gogo-connector/libs/mqtt"
	"gogo-connector/libs/package_coder"
	"gogo-connector/libs/proto_coder"
	"log"
	"sync"
	"time"
)

type UserReq = global.UserReq
type PkgBelong = global.PkgBelong
type RemoteConnect = global.RemoteConnect

var serverOptLocker sync.Mutex

func MonitorHandler(action string, ss *types.MonitorBody) (req, respBody, respErr, notify []byte) {
	switch action {
	case "addServer":
		{
			fmt.Println("add server .............")
			AddServers([]types.RegisterInfo{ss.Server})
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
func AddServers(ss []types.RegisterInfo) {

	for _, s := range ss {
		if *config.ServerType == s.ServerType {
			fmt.Println("skip init same type server", s.ServerID)
			continue
		}

		fmt.Println(s)
		client := &RemoteConnect{
			Host:       s.Host,
			Port:       int32(s.Port),
			ServerID:   s.ServerID,
			ServerType: s.ServerType,
		}
		if func() bool {
			serverOptLocker.Lock()
			defer serverOptLocker.Unlock()

			if _, ok := global.RemoteStore.Get(s.ServerID); ok {
				return true
			}
			return false
		}() {
			continue
		}
		if err := client.Start(nil, handlePublish); err != nil {
			fmt.Println(err)
		}
		log.Println("链接服务器", s.ServerID, s.ServerType, s.ServerID, s.Host, s.Port, client.MqttClient.IsConnected())

		func() {
			serverOptLocker.Lock()
			defer serverOptLocker.Unlock()

			global.RemoteStore.Set(s.ServerID, client)
			if !global.RemoteTypeStore.Has(s.ServerType) {
				global.RemoteTypeStore.Set(s.ServerType, &global.RemoteTypeStoreType{
					Servers:   []*RemoteConnect{},
					ForwardCh: make(chan UserReq, 100000),
				})
			}
			remoteType, _ := global.RemoteTypeStore.Get(s.ServerType)
			rr, _ := remoteType.(*global.RemoteTypeStoreType)
			rr.Servers = append(rr.Servers, client)
			fmt.Println("添加", s.ServerType, s.ServerID, len(rr.Servers))
		}()

		go func(s types.RegisterInfo, client *RemoteConnect) {
			ss, ok := global.RemoteTypeStore.Get(s.ServerType)
			if !ok {
				fmt.Println("no found server in store.")
			}
			serverTypeInfo, _ := ss.(*global.RemoteTypeStoreType)

			for msg := range serverTypeInfo.ForwardCh {
				fmt.Println(">>>forward rpc to backend <<<", s.Host, s.Port, s.ServerID, msg.ServerType)
				pkg, err := proto_coder.PbToJson(msg.Route, msg.Payload)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(">><>><", string(pkg))
				msg.Payload = pkg
				p := package_coder.Encode(&msg, s.ServerID) // 后端 wrap
				if p == nil {
					continue
				}
				client.MqttClient.Publish("rpc", p, 0, true)
				fmt.Println("rpc send ok", client.MqttClient.ClientID)
			}
		}(s, client)
	}

	time.Sleep(time.Second * 100)
	//return nil
}

func handlePublish(mqtt *mqtt.MQTT, msg paho.Message) {
	package_coder.Decode(msg)
}
