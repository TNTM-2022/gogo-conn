package package_coder

import (
	"encoding/json"
	"fmt"
	paho "github.com/eclipse/paho.mqtt.golang"
	concurrentMap "github.com/orcaman/concurrent-map"
	"gogo-connector/components/global"
	coder "gogo-connector/libs/pomelo_coder"
	"gogo-connector/libs/proto_coder"
	sendProto "gogo-connector/libs/proto_coder/protos/forward_proto"
	"log"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type pkgIdType struct {
	pkgid      int64
	countMutex sync.Mutex
}

type UserReq = global.UserReq
type PkgBelong = global.PkgBelong

func (p *pkgIdType) genPkgId() int64 {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()
	r := atomic.AddInt64(&p.pkgid, 1)
	if r > math.MaxInt64-1 {
		atomic.StoreInt64(&p.pkgid, 1)
		r = 1
	}
	return r
}

var pkgId pkgIdType
var pkgMap = concurrentMap.New() // 记录发往后端的 packageId
func Encode(userReq *UserReq, serverId string) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	pkgId := pkgId.genPkgId()
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
					FrontendId: serverId,
					Uid:        userReq.UID,
					Settings: &sendProto.Settings{
						GameId: 0,
						RoomId: 1,
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

func Decode(m paho.Message) {
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
	defer pkgMap.Remove(strconv.FormatInt(rec.Id, 10))
	pp := _pp.(*PkgBelong)
	//defer delete(pkgMap, rec.Id)
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
		select {
		case t.(*coder.UserConn).MsgResp <- mm:
		default:
		}
	}
}
