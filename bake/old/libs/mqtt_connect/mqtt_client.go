package mqtt_connect

import (
	"encoding/json"
	"fmt"
	"go-connector/interfaces"
	"log"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cmap "github.com/orcaman/concurrent-map"
	sendProto "go-connector/common/protos/forward_proto"
	userReqType "go-connector/common/types"
	"go-connector/common/utils"
	conf "go-connector/config"
	"go-connector/global"
	coder "go-connector/libs/pomelo_coder"
	"go-connector/libs/proto_coder"
)

// RequestRemote
// monitor
type Gate struct {
	pkgid                int64
	GlobalForwardMsgChan chan userReqType.UserReq
	countMutex           sync.Mutex
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

//var pkgMap = make(map[int64]PkgBelong, 100000)
var pkgMap = cmap.New()

func (g *Gate) genPid() int64 {
	g.countMutex.Lock()
	defer g.countMutex.Unlock()
	r := atomic.AddInt64(&g.pkgid, 1)
	if r > math.MaxInt64-1 {
		atomic.StoreInt64(&g.pkgid, 1)
		r = 1
	}
	return r
}

func (gate *Gate) StartGate(addr string, serverId string) error {
	log.Println("gate connect ", addr, serverId)
	opt := mqtt.NewClientOptions()
	opt.AddBroker(fmt.Sprintf("tcp://%s", addr))
	opt.SetDefaultPublishHandler(handleDecode)
	mqttClient := mqtt.NewClient(opt)
	defer mqttClient.Disconnect(0)
	token := mqttClient.Connect() // todo pomelo bug  如果不修改connect事件返回， 这里将会一直堵着 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 中的这个  socket.on('connect', function (pkg) { 代码块内部
	//mqttClient.Connect() // todo pomelo bug  如果不修改connect事件返回， 这里将会一直堵着 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 中的这个  socket.on('connect', function (pkg) { 代码块内部
	if !token.WaitTimeout(1 * time.Second) {
		log.Println("mqtt connect timeout")
		//return nil;
	}

	for msg := range gate.GlobalForwardMsgChan {
		fmt.Println(">>>forward rpc to backend <<<", addr, serverId, msg.ServerType, addr)
		fmt.Println("sss >> ", msg)
		pkg, err := proto_coder.PbToJson(msg.Route, msg.Payload)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(">><>><", string(pkg))
		msg.Payload = pkg
		p := gate.encode(&msg)
		if p == nil {
			continue
		}
		//token := mqttClient.Publish("rpc", 0, false, p)
		mqttClient.Publish("rpc", 0, false, p)
		//mqttClient.Publish("rpc", 0, false, p)
		token.Wait() // todo 这里没啥必要去 wait，
		fmt.Println("rpc sender")
	}

	return nil
}

func (gate *Gate) encode(userReq *userReqType.UserReq) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	pkgId := gate.genPid()
	pkgMap.Set(strconv.FormatInt(pkgId, 10), &PkgBelong{
		UID:         userReq.UID,
		StartAt:     time.Now(),
		ClientPkgID: userReq.PkgID,
		Route:       userReq.Route,
	})
	fmt.Println(userReq.Sid, userReq.UID)
	m := &sendProto.Payload{
		Id: pkgId, // 通过 此处的 id 进行路由对用的user

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
					FrontendId: *conf.ServerID,
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

type RawResp struct {
	Id   int64
	Resp []json.RawMessage
}

func handleDecode(_ mqtt.Client, m mqtt.Message) {
	log.Println("topic", m.Topic(), m.MessageID(), string(m.Payload()))

	var rec RawResp
	if e := json.Unmarshal(m.Payload(), &rec); e != nil {
		fmt.Println(e)
		return
	}

	//if len(rec.Resp) != 2 {
	//	fmt.Println("response array len not right", rec.Resp, len(rec.Resp));
	//	fmt.Println(string(rec.Resp[0]))
	//	fmt.Println(string(rec.Resp[1]))
	//	fmt.Println(string(rec.Resp[2]))
	//	return;
	//}
	_pp, ok := pkgMap.Get(strconv.FormatInt(rec.Id, 10))
	if !ok {
		fmt.Println("no package info found")
		return
	}
	pp := _pp.(*PkgBelong)
	//defer delete(pkgMap, rec.Id)
	defer pkgMap.Remove(strconv.FormatInt(rec.Id, 10))
	route := pp.Route

	// do compose msg payload
	var b []byte
	if len(rec.Resp) == 2 && rec.Resp[1] != nil {
		b = rec.Resp[1]
	} else if len(rec.Resp) == 3 && rec.Resp[1] != nil {
		b = rec.Resp[1]
	} else {
		fmt.Println("skip")
		return
	}
	fmt.Println("jsonStr>>>>", string(b))
	fmt.Println("route>>>>", route)
	if _b, e := proto_coder.JsonToPb(route, b, false); _b != nil && e == nil {
		log.Println(" json2bt转换成功-", route)
		b = _b
	} else {
		log.Println(" json2pb转换失败", route)
	}

	mm := coder.MessageEncode(uint64(pp.ClientPkgID), coder.Message["TYPE_RESPONSE"], 0, pp.Route, b, false)
	fmt.Println(mm, "--->>>> content: ", string(b))
	mm = coder.PackageEncode(coder.Package["TYPE_DATA"], mm)
	if t, ok := global.Users.Get(strconv.FormatInt(int64(pp.UID), 10)); ok {

		utils.SafeSend(t.(*interfaces.UserConn).MsgResp, mm)
	}
}

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
