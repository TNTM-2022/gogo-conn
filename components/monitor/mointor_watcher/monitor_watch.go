package mointor_watcher

import (
	"encoding/json"
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	config "gogo-connector/components/config"
	"gogo-connector/components/monitor/types"
	"gogo-connector/libs/mqtt"
	"gogo-connector/libs/proto_coder"
	sendProto "gogo-connector/libs/proto_coder/protos/forward_proto"
	"log"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PkgBelong struct {
	UID         int32
	StartAt     time.Time
	ClientPkgID int64
	Route       string
}

type UserReq struct {
	UID        int32
	Route      string
	ServerType string
	Payload    []byte
	PkgID      int64
	Sid        uint64
}

// RequestRemote
type RemoteConnect struct {
	pkgid                int64
	GlobalForwardMsgChan chan UserReq
	countMutex           sync.Mutex
}

func (conn *RemoteConnect) genPkgId() int64 {
	conn.countMutex.Lock()
	defer conn.countMutex.Unlock()
	r := atomic.AddInt64(&conn.pkgid, 1)
	if r > math.MaxInt64-1 {
		atomic.StoreInt64(&conn.pkgid, 1)
		r = 1
	}
	return r
}

func (conn *RemoteConnect) connect(host string, port string, serverId string) error {
	log.Println("mqtt client connect to ", host, port, serverId)

	// todo pomelo bug  如果不修改connect事件返回， 这里将会一直堵着 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 中的这个  socket.on('connect', function (pkg) { 代码块内部
	mqttClient := mqtt.CreateMQTTClient(&mqtt.MQTT{
		Host:            host,
		Port:            port,
		ClientID:        "clientId-1",
		SubscriptionQos: 1,
		Persistent:      true,
		Order:           false,
		KeepAliveSec:    5,
		PingTimeoutSec:  10,

		//OnConnectCb: regServer,
		//OnPublishCb: publishCb,
	})

	mqttClient.Start()

	go func() {
		for msg := range conn.GlobalForwardMsgChan {
			fmt.Println(">>>forward rpc to backend <<<", host, port, serverId, msg.ServerType)
			pkg, err := proto_coder.PbToJson(msg.Route, msg.Payload)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(">><>><", string(pkg))
			msg.Payload = pkg
			p := conn.encode(&msg) // 后端 wrap
			if p == nil {
				continue
			}
			//token := mqttClient.Publish("rpc", 0, false, p)
			//mqttClient.Publish("rpc", p, 0)
			mqttClient.Publish("rpc", p, 0, true)
			//mqttClient.Publish("rpc", 0, false, p)
			//token.Wait() // todo 这里没啥必要去 wait，
			fmt.Println("rpc send ok")
		}
	}()

	return nil
}

func (conn *RemoteConnect) encode(userReq *UserReq) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	pkgId := conn.genPkgId()
	pkgMap.Set(strconv.FormatInt(pkgId, 10), &PkgBelong{
		UID:         userReq.UID,
		StartAt:     time.Now(),
		ClientPkgID: userReq.PkgID,
		Route:       userReq.Route,
	})
	fmt.Println(userReq.Sid, userReq.UID)
	m := &sendProto.Payload{
		Id: pkgId, // 通过 此处的 id 进行路由对应的user

		Msg: &sendProto.PayloadMsg{
			Namespace:  "sys",
			ServerType: userReq.ServerType,
			Service:    "msgRemote",
			Method:     "forwardMessage",
			Args: []*sendProto.PayloadMsgArgs{
				&sendProto.PayloadMsgArgs{
					Id:    userReq.PkgID, // 客户端包
					Route: userReq.Route,
					Body:  userReq.Payload,
				},
				&sendProto.PayloadMsgArgs{
					Id:         int64(userReq.Sid), // sid
					FrontendId: *config.ServerID,
					Uid:        userReq.UID,
					Settings: &sendProto.Settings{
						GameId: 123,
						RoomId: 41,
						UID:    userReq.UID,
						DeskId: 10000000010041,
					},
				},
			},
		},
	}

	//_, _ = pb.Marshal(m)
	// if j, e := pb.Marshal(m); e == nil {
	// 	return j
	// }
	if j, e := json.Marshal(m); e == nil {
		fmt.Println("......", string(j))
		return j
	}

	return nil
}

// -------------------------------------

var ServerIDMap = make(map[string]types.RegisterInfo)              // serverId -> serverInfo
var ServerTypeMap = make(map[string]map[string]types.RegisterInfo) // serverType -> {serverId: serverInfo}

var ServerConnectMap = make(map[string]*RemoteConnect) // 保存连接远程的mqtt client
var ServerTypeChMap = make(map[string]chan UserReq)    // server类型 -> 转发通道

var ServerOptLocker sync.RWMutex
var pkgMap = cmap.New()

func MonitorHandler(action string, ss *types.MonitorBody) (req, respBody, respErr, notify []byte) {
	switch action {
	case "addServer":
		{
			AddServers([]types.RegisterInfo{ss.Server})
			respErr = json.RawMessage(`1`)
		}
	case "removeServer":
		{
			removeServers(ss.ServerID)
			respErr = json.RawMessage(`1`)
			respBody = json.RawMessage(`1`)
		}
	case "replaceServer":
	case "startOver":
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
		var gate RemoteConnect
		func() {
			ServerOptLocker.Lock() // 使用 concurrency map 替代
			defer ServerOptLocker.Unlock()

			ServerIDMap[s.ServerID] = s
			if ServerTypeMap[s.ServerType] == nil {
				ServerTypeMap[s.ServerType] = make(map[string]types.RegisterInfo, 10)
				ServerTypeChMap[s.ServerType] = make(chan UserReq, 10000)
			}
			ServerTypeMap[s.ServerType][s.ServerID] = s
			gate = RemoteConnect{
				GlobalForwardMsgChan: ServerTypeChMap[s.ServerType],
			}
		}()
		if err := gate.connect(s.Host, fmt.Sprintf("%v", s.Port), s.ServerID); err != nil {
			fmt.Println(err)
		}
		log.Println("链接服务器", s.ServerID, s.ServerType, s.ServerID, s.Host, s.ClientPort)

	}
	//return nil
}
