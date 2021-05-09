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
	"reflect"
	"sync"
	"time"
)

type BackendMsg = package_coder.BackendMsg
type PkgBelong = package_coder.PkgBelong

//var serverOptLocker sync.Mutex

var servOptLocker sync.Mutex

var LP TaskLoop

func init() {
	LP = CreateTaskLoop()
	go LP.Run(func(v interface{}) interface{} {
		f, ok := v.(func())
		if !ok {
			fmt.Println("wrong func param type", reflect.TypeOf(v))
		}
		if f != nil {
			f()
		}
		return nil
	})
}

func init() {
	go func() {
		t := time.Tick(time.Second * 5)
		for range t {
			fmt.Println(global.RemoteBackendClients.Keys())
			fmt.Println(global.RemoteBackendTypeForwardChan.Keys())
		}
	}()
}

// todo 服务器相关操作 是不是需要单独开个 goroutine 进行操作呢？， 防止请求间隔太短以及锁没有顺序带来的潜在问题呢？ 排除因为报文前后顺序

func MonitorHandler(action string, ss types.MonitorBody) (req, respBody, respErr, notify []byte) {
	switch action {
	case "addServer":
		{
			fmt.Println("add server >>>>")
			ConnectToServer(ss.Server)
			fmt.Println("add server  <<<<<")
			respErr = json.RawMessage(`1`)
		}
	case "removeServer":
		{
			fmt.Println(ss)
			if ss.ServerID == "" {
				respErr = json.RawMessage(`0`)
				respBody = json.RawMessage(`0`)
				return
			}
			fmt.Println("remove server .............")
			removeServer(ss.ServerID)
			respErr = json.RawMessage(`1`)
			respBody = json.RawMessage(`1`)
		}
	case "replaceServer": // master 启动后， 重新同步信息； 防止出现 同名 但是不同端口 之类的事情发生
		fmt.Println("replace server .............")
		replaceServer(ss.Servers)
		respErr = json.RawMessage(`1`)
		//respBody = json.RawMessage(`1`)
	case "startOver":
		fmt.Println("start over .............")
		respErr = json.RawMessage(`1`) // 全部启动了， 再发这个

	}
	return
}

func removeServer(serverId string) {
	//<-LP.Push(func() {
	//	fmt.Println("test", serverId)
	//})
	<-LP.Push(func() {
		//fmt.Println("del >>>>>", serverId)
		del(serverId)
		//fmt.Println("del <<<<<", serverId)
	})
}

func del(serverId string) {
	var _serv interface{}
	if !global.RemoteBackendClients.RemoveCb(serverId, func(key string, v interface{}, exists bool) bool {
		if !exists {
			return false
		}
		_serv = v
		return true
	}) {
		fmt.Println("no exists server info")
		return
	}
	serv, _ := _serv.(*mqtt_client.MQTT)
	//serv.Closing = true
	serv.SetReconnectCb(func(m *mqtt_client.MQTT) {
		fmt.Println("do reconnect cb", m.ClientID)

		m.Stop(1)            // 都重连了， 没必要等待未处理完的报文
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
		fmt.Println("服务器 不重连--", serverTypeAllClosed, serverId)
	})
	fmt.Println("remove server-- ", serverId)
}

func ConnectToServer(serv types.RegisterInfo) {
	<-LP.Push(func() {
		//fmt.Println("add >>>>>")
		add(serv)
		//fmt.Println("add <<<<<")
	})
}

func add(serv types.RegisterInfo) {
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
	//defer client.Stop()

	// 初始化 serverType：chan  serverType：serverId：serverInfo
	// todo 资源回收  如果撤掉了 * servertype 通道需要关闭
	servExists := false
	servOptLocker.Lock()
	global.RemoteBackendTypeForwardChan.SetIfAbsent(serv.ServerType, make(chan package_coder.BackendMsg, 10000)) // 不用回收
	global.RemoteBackendClients.Upsert(serv.ServerID, client, func(exists bool, oldV, newV interface{}) interface{} {
		if !exists {
			return newV
		}
		oldServ, _ := oldV.(*mqtt_client.MQTT)
		newServ, _ := newV.(*mqtt_client.MQTT)
		if oldServ.Host == newServ.Host && oldServ.Port == newServ.Port && oldServ.ClientID == newServ.ClientID && oldServ.ServerType == newServ.ServerType {
			fmt.Println("保持", oldServ.ClientID)
			servExists = true
			return oldV
		}

		v, _ := oldV.(*mqtt_client.MQTT)
		go v.Stop(5)
		fmt.Println("关闭？？？？", v.ClientID)
		// todo 停止消息转发， 然后再停止server client
		return newV
	})
	servOptLocker.Unlock()
	if servExists {
		fmt.Println("client lian le yijing ")
		return
	}

	go func(s types.RegisterInfo, client *mqtt_client.MQTT) {
		client.Start()
		log.Println("链接服务器", serv.ServerID, serv.ServerType, serv.ServerID, serv.Host, serv.Port, client.IsConnectionOpen())
		_forwardChan, ok := global.RemoteBackendTypeForwardChan.Get(s.ServerType)
		if !ok {
			log.Println("no found server in store.")
		}
		forwardChan, _ := _forwardChan.(chan package_coder.BackendMsg)

		for {
			fmt.Println("++++++++")
			select {
			case <-client.Quit:
				{
					fmt.Println("close loop")
					return
				}
			case msg := <-forwardChan:
				{
					if msg.Sid == 0 {
						continue
					}
					logger.DEBUG.Println(">>forward rpc to backend == ", s.Host, s.Port, s.ServerID, msg.ServerType)
					pkgId := client.GetReqId()
					p := package_coder.Encode(pkgId, &msg) // 后端 wrap 组装 session
					if p == nil {
						log.Println("encoding skip...")
						continue
					}

					if !client.IsConnectionOpen() { // 如果server 关闭了 消息要重新推回去
						fmt.Printf("client.IsConnectionOpen() = %v", false)
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
			}
		}
	}(serv, client)
}
func replaceServer(serv map[string]types.RegisterInfo) {
	for i := range global.RemoteBackendClients.IterBuffered() {
		oldServ := i.Val.(*mqtt_client.MQTT)
		if newServ, ok := serv[i.Key]; ok {
			ConnectToServer(newServ)
		} else {
			fmt.Println("连上了？1", oldServ.ClientID)
			if oldServ.IsConnectionOpen() {
				fmt.Println("连上了？2", oldServ.ClientID)
				break
			}
			removeServer(oldServ.ClientID)
		}
		fmt.Println(i)
	}
	fmt.Println("replace server done")
}
