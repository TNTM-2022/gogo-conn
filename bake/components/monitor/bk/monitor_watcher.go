package bk

import (
	"encoding/json"
	"fmt"
	cmap "github.com/orcaman/concurrent-map"
	"gogo-connector/components/config"
	cfg "gogo-connector/components/config"
	"gogo-connector/components/global"
	"gogo-connector/components/monitor/types"
	mqtt "gogo-connector/libs/mqtt"
	"gogo-connector/libs/proto_coder"
	sendProto "gogo-connector/libs/proto_coder/protos/forward_proto"
	"log"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

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

var ServerIDMap = make(map[string]*types.RegisterInfo)              // serverId -> serverInfo
var ServerTypeMap = make(map[string]map[string]*types.RegisterInfo) // serverType -> {serverId: serverInfo}

var ServerConnectMap = make(map[string]*RemoteConnect) // 保存连接远程的mqtt client
var ServerTypeChMap = make(map[string]chan UserReq)    // server类型 -> 转发通道

var ServerOptLocker sync.RWMutex

func ReplaceServer() { // todo 待实现

}

func AddServer(s *types.RegisterInfo) error {
	if *cfg.ServerType == s.ServerType {
		fmt.Println("skip init same type server")
		return nil
	}

	ServerOptLocker.Lock()
	defer ServerOptLocker.Unlock()

	ServerIDMap[s.ServerID] = s
	if ServerTypeMap[s.ServerType] == nil {
		ServerTypeMap[s.ServerType] = make(map[string]*types.RegisterInfo, 10)
		ServerTypeChMap[s.ServerType] = make(chan UserReq, 10000)
	}
	log.Println(">> add server", s.ServerID, s.ServerType)
	ServerTypeMap[s.ServerType][s.ServerID] = s

	gate := RemoteConnect{
		GlobalForwardMsgChan: ServerTypeChMap[s.ServerType],
	}
	ServerConnectMap[s.ServerID] = &gate

	if err := gate.connect(s.Host, fmt.Sprintf("%v", s.Port), s.ServerID); err != nil {
		fmt.Println(err)
	}
	fmt.Println("链接服务器", s.ServerID, s.Host, s.ClientPort)
	return nil
}

func AddServers(s map[string]types.RegisterInfo) []byte {
	var wg sync.WaitGroup
	wg.Add(len(s))
	for _, ss := range s {
		sss := ss
		fmt.Println(ss.ServerType, ss.ServerID)
		go func() {
			defer wg.Done()
			e := AddServer(&sss)
			if e != nil {
				fmt.Println(e)
			}
		}()
	}
	wg.Wait()
	return []byte("1")
}
func RemoveServer(id string) {
	ServerOptLocker.Lock()
	defer ServerOptLocker.Unlock()

	if s := ServerIDMap[id]; s != nil {
		log.Println("<< remove server", s.ServerID)
		delete(ServerIDMap, id)
		if ServerTypeMap[s.ServerType] != nil {
			delete(ServerTypeMap[s.ServerType], id)
			if len(ServerTypeMap[s.ServerType]) == 0 {
				delete(ServerTypeMap, s.ServerType)
			}
		}
	}
}
func GetServerByServerId(id string) types.RegisterInfo {
	var r types.RegisterInfo
	ServerOptLocker.RLock()
	defer ServerOptLocker.RUnlock()
	if s := ServerIDMap[id]; s != nil {
		r = *s
	}
	return r
}
func GetServersByServerType(t string) []types.RegisterInfo {
	var r []types.RegisterInfo
	ServerOptLocker.RLock()
	defer ServerOptLocker.RUnlock()
	if ServerTypeMap[t] != nil {
		for _, v := range ServerTypeMap[t] {
			r = append(r, *v)
		}
	}
	return r
}

/**
pkg id 标识 从那个conn 发送出来的

*/

type PkgBelong struct {
	UID         int32
	StartAt     time.Time
	ClientPkgID int64
	Route       string
}

var pkgMap = cmap.New()

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
					Uid:        global.UserID(userReq.UID),
					Settings: &sendProto.Settings{
						GameId: 123,
						RoomId: 41,
						UID:    global.UserID(userReq.UID),
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

type RawResp struct {
	Id   int64
	Resp []json.RawMessage
}

//func handleDecode(_ mqtt.Client, m mqtt.Message) {
//	log.Println("topic", m.Topic(), m.MessageID(), string(m.Payload()))
//
//	var rec RawResp
//	if e := json.Unmarshal(m.Payload(), &rec); e != nil {
//		fmt.Println(e)
//		return
//	}
//
//	//if len(rec.Resp) != 2 {
//	//	fmt.Println("response array len not right", rec.Resp, len(rec.Resp));
//	//	fmt.Println(string(rec.Resp[0]))
//	//	fmt.Println(string(rec.Resp[1]))
//	//	fmt.Println(string(rec.Resp[2]))
//	//	return;
//	//}
//	_pp, ok := pkgMap.Get(strconv.FormatInt(rec.Id, 10))
//	if !ok {
//		fmt.Println("no package info found")
//		return
//	}
//	pp := _pp.(*PkgBelong)
//	//defer delete(pkgMap, rec.Id)
//	defer pkgMap.Remove(strconv.FormatInt(rec.Id, 10))
//	route := pp.Route
//
//	// do compose msg payload
//	var b []byte
//	if len(rec.Resp) == 2 && rec.Resp[1] != nil {
//		b = rec.Resp[1]
//	} else if len(rec.Resp) == 3 && rec.Resp[1] != nil {
//		b = rec.Resp[1]
//	} else {
//		fmt.Println("skip")
//		return
//	}
//	fmt.Println("jsonStr>>>>", string(b))
//	fmt.Println("route>>>>", route)
//	if _b, e := proto_coder.JsonToPb(route, b, false); _b != nil && e == nil {
//		log.Println(" json2bt转换成功-", route)
//		b = _b
//	} else {
//		log.Println(" json2pb转换失败", route)
//	}
//
//	mm := coder.MessageEncode(uint64(pp.ClientPkgID), coder.Message["TYPE_RESPONSE"], 0, pp.Route, b, false)
//	fmt.Println(mm, "--->>>> content: ", string(b))
//	mm = coder.PackageEncode(coder.Package["TYPE_DATA"], mm)
//	if t, ok := global.Users.Get(strconv.FormatInt(int64(pp.UID), 10)); ok {
//
//		utils.SafeSend(t.(*interfaces.UserConn).MsgResp, mm)
//	}
//}

//func handleTimeout() {
//	for {
//		n := time.Now()
//		for _, pkgInf := range (pkgMap) {
//			if n.Sub(pkgInf.StartAt).Seconds() > 4 {
//				// 超时处理
//				if t, ok:= global.Users.Get(strconv.FormatInt(int64(pkgInf.UID), 10)); ok {
//					user := t.(*global.UserConn)
//					user.MsgRes <- []byte{}
//				}
//			}
//		}
//	}
//}
//func clearTimeout() {
//
//}
