package package_coder

import (
	"encoding/json"
	"fmt"
	"go-connector/logger"
)

var pkgId pkgIdType

func Encode(pkgId int64, u *BackendMsg) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()

	//pkgId := pkgId.genPkgId()
	//pkgMap.Set(strconv.FormatInt(pkgId, 10), &PkgBelong{
	//	SID:         u.Sid,
	//	StartAt:     time.Now(),
	//	ClientPkgID: u.PkgID,
	//	Route:       u.Route,
	//})

	m := &Payload{
		Id: pkgId, // 通过 此处的 id 进行路由对应的user

		Msg: &PayloadMsg{
			Namespace:  "sys",
			ServerType: u.ServerType, // todo 确认 是 本server 还是目标server
			Service:    "msgRemote",
			Method:     "forwardMessage",
			Args: []*PayloadMsgArgs{
				&PayloadMsgArgs{
					Id:    u.PkgID, // 客户端包
					Route: u.Route,
					Body:  u.Payload,
				},
				&PayloadMsgArgs{
					Id:         int64(u.Sid), // sid
					FrontendId: u.ServerId,
					//Settings:,
					// todo session 设置
					//Uid:        userReq.UID,
					//Settings: &Settings{
					//	GameId: 0,
					//	RoomId: 1,
					//	UID:    userReq.UID,
					//	DeskId: 10000000010041,
					//},
				},
			},
		},
	}

	if j, e := json.Marshal(m); e == nil {
		fmt.Println("......", string(j))
		return j
	}

	return nil
}

// {"id":1,"msg":{"namespace":"sys","service":"channelRemote","method":"broadcast","args":["broadcast.test",{"isPush":true},{"type":"broadcast","userOptions":{},"isBroadcast":true}]}}
// {"id":0,"msg":{"namespace":"sys","service":"channelRemote","method":"pushMessage","args":["push.push",{"type":"push","is_broad":true},[1],{"type":"push","userOptions":{},"isPush":true}]}}
type RawRecv struct {
	Id  int64
	Msg struct {
		Namespace string
		Service   string
		Method    string
		Args      []json.RawMessage
	}
	Resp []json.RawMessage
}

func DecodeResp(topic string, messageID uint16, payload []byte) (pkgId int64, u *BackendMsg) {
	var (
		rec RawRecv

		buf   []byte
		route string
		t     string
	)
	if e := json.Unmarshal(payload, &rec); e != nil {
		fmt.Println(e)
	}

	if rec.Resp != nil {
		pkgId = rec.Id
		decodeResp(&rec, u)
		t = "response"
	}

	logger.DEBUG.Println(route, t, string(buf))
	return
}

func DecodePush(topic string, messageID uint16, payload []byte) (sids []uint32, u *BackendMsg) {
	var (
		rec RawRecv

		buf   []byte
		route string
		t     string
	)
	if e := json.Unmarshal(payload, &rec); e != nil {
		fmt.Println(e)
	}

	if rec.Msg.Args != nil {
		sids = decodePush(&rec, u)
		t = "push"
	}

	logger.DEBUG.Println(route, t, string(buf))
	return
}

type MsgOptions struct {
	Type        string
	UserOptions json.RawMessage
	IsPush      bool
	IsBroadcast bool
}

func handlePushOrBroad(b []json.RawMessage) (route string, sids []uint32, cc json.RawMessage, userOptions json.RawMessage) {
	if len(b) < 3 {
		return
	}

	var handleType MsgOptions
	if e := json.Unmarshal(b[len(b)-1], &handleType); e != nil {
		fmt.Println(e, b[len(b)-1])
	}
	if handleType.IsPush {
		if err := json.Unmarshal(b[2], &sids); err != nil {
			logger.ERROR.Println(err)
		}
	}
	route = string(b[0])
	cc = b[1]
	userOptions = handleType.UserOptions
	return
}

func decodePush(rec *RawRecv, um *BackendMsg) (sids []uint32, ) {
	route, sids, payload, opts := handlePushOrBroad(rec.Msg.Args)

	if route == "" {
		fmt.Println("skip", sids)
		return
	}

	um.Route = route
	um.Payload = payload
	um.Opts = opts

	logger.DEBUG.Printf("push router:>> %v; jsonStr>> %v", route, string(payload))
	return
}

func decodeResp(rec *RawRecv, um *BackendMsg) {
	//_pp, ok := pkgMap.Get(strconv.FormatInt(rec.Id, 10))
	//if !ok {
	//	logger.ERROR.Println("no package info found")
	//	return
	//}
	//defer pkgMap.Remove(strconv.FormatInt(rec.Id, 10))
	//pp := _pp.(*PkgBelong)
	//
	//route, sids := pp.Route, []uint32{pp.SID}

	//var payload []byte
	if len(rec.Resp) >= 2 && rec.Resp[1] != nil {
		um.Payload = rec.Resp[1]
		logger.DEBUG.Printf("resp router:>> %v; jsonStr>> %v","", string(um.Payload))
	} else {
		//logger.DEBUG.Printf("router: %v; skip", route)
	}

	//um.Route = route
	//um.Sid = pp.SID
	//um.PkgID = pp.ClientPkgID
	//payload

	return

	//todo opts 填充值
}

//
//type RawPush struct {
//	Id  int64
//	Msg struct {
//		Namespace string
//		Service   string
//		Method    string
//		Args      []json.RawMessage
//	}
//	Resp []json.RawMessage
//}
//type RawResp struct {
//	Id   int64
//	Resp []json.RawMessage
//}
//
//func DecodePush(topic string, messageID uint16, payload []byte) {
//	log.Println("topic", topic, messageID, string(payload))
//
//	var (
//		rec   RawPush
//		b     []byte
//		route string
//	)
//	if e := json.Unmarshal(payload, &rec); e != nil {
//		fmt.Println(e)
//	}
//
//	route, uids, sids, b := handlePushOrBroad(rec.Msg.Args)
//
//	if route == "" {
//		fmt.Println("skip", uids, sids)
//		return
//	}
//
//	fmt.Println("jsonStr>>>>", string(b))
//	fmt.Println("route>>>>", route)
//	//if _b, e := proto_coder.JsonToPb(route, b, true); _b != nil && e == nil {
//	//	log.Println(" json2bt转换成功-", route)
//	//	b = _b
//	//} else {
//	//	log.Println(" json2pb转换失败", route)
//	//}
//	//
//	//mm := coder.MessageEncode(0, coder.Message["TYPE_PUSH"], 0, route, b, false)
//	////mm := coder.MessageEncode(uint64(pp.ClientPkgID), coder.Message["TYPE_RESPONSE"], 0, route, b, false)
//	//fmt.Println(mm, "--->>>> content: ", string(b))
//	//mm = coder.PackageEncode(coder.Package["TYPE_DATA"], mm)
//	//global.Users.IterCb(func(k string, v interface{}) {
//	//	select {
//	//	case v.(*coder.UserConn).MsgResp <- mm:
//	//	default:
//	//	}
//	//})
//}
//
//func DecodeResp(topic string, messageID uint16, payload []byte) {
//	//log.Println("topic", m.Topic(), m.MessageID(), string(m.Payload()))
//	log.Println("topic", topic, messageID, string(payload))
//
//	var (
//		rec RawResp
//		b   []byte
//	)
//	if e := json.Unmarshal(payload, &rec); e != nil {
//		fmt.Println(e)
//	}
//
//	_pp, ok := pkgMap.Get(strconv.FormatInt(rec.Id, 10))
//	if !ok {
//		fmt.Println("no package info found")
//		return
//	}
//	defer pkgMap.Remove(strconv.FormatInt(rec.Id, 10))
//	pp := _pp.(*PkgBelong)
//	//defer delete(pkgMap, rec.Id)
//	route := pp.Route
//	fmt.Println(route)
//	if len(rec.Resp) == 2 && rec.Resp[1] != nil {
//		b = rec.Resp[1]
//	} else if len(rec.Resp) == 3 && rec.Resp[1] != nil {
//		b = rec.Resp[1]
//	} else {
//		fmt.Println("skip")
//	}
//	_ = b
//	//if len(rec.Resp) != 2 {
//	//	fmt.Println("response array len not right", rec.Resp, len(rec.Resp));
//	//	fmt.Println(string(rec.Resp[0]))
//	//	fmt.Println(string(rec.Resp[1]))
//	//	fmt.Println(string(rec.Resp[2]))
//	//	return;
//	//}
//
//	// do compose msg payload
//
//	//fmt.Println("jsonStr>>>>", string(b))
//	//fmt.Println("route>>>>", route)
//	//if _b, e := proto_coder.JsonToPb(route, b, false); _b != nil && e == nil {
//	//	log.Println(" json2bt转换成功-", route)
//	//	b = _b
//	//} else {
//	//	log.Println(" json2pb转换失败", route)
//	//}
//	//
//	//mm := coder.MessageEncode(uint64(pp.ClientPkgID), coder.Message["TYPE_RESPONSE"], 0, pp.Route, b, false)
//	//fmt.Println(mm, "--->>>> content: ", string(b))
//	//mm = coder.PackageEncode(coder.Package["TYPE_DATA"], mm)
//	//if t, ok := global.Users.Get(strconv.FormatInt(int64(pp.UID), 10)); ok {
//	//	select {
//	//	case t.(*coder.UserConn).MsgResp <- mm:
//	//	default:
//	//	}
//	//}
//}
