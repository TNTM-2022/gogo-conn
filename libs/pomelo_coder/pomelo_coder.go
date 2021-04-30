package pomelo_coder

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

// user conn 状态
const (
	StateInitOk  = 0
	StateWaitAck = 1
	StateWorking = 2
	StateClosed  = 3
)

// 握手状态
const (
	CODE_OK         = 200
	CODE_USE_ERROR  = 500
	CODE_OLD_CLIENT = 501
)

//===================== handle pomelo protocol ===========================

func InitCoder() *Coder {
	return &Coder{
		State: StateInitOk,
	}
}

type Coder struct {
	State rune
}

// HandleHandshake 处理客户端handshake
// 检查整体状态 ST_INITED
// todo checkClient
// var opts = { heartbeat : setupHeartbeat(this) };
//  opts.useProto = true;
// 返回 TYPE_HANDSHAKE 报文， 携带上述的对象 packageEncode
func (p *Coder) HandleHandshake() []byte {
	if p.State != StateInitOk {
		//user.Cancel()
		return nil
	}
	s := handshake{
		Code: CODE_OK,
		Sys: sys{
			Heartbeat:   60,
			Dict:        dict{},
			RouteToCode: routeToCode{},
			CodeToRoute: codeToRoute{},
			DictVersion: genDictVersion(),
			UseDict:     true,
			UseProto:    true,
		},
	}
	j, _ := json.Marshal(s)

	p.State = StateWaitAck
	return PackageEncode(Package["TYPE_HANDSHAKE"], []byte(string(j)))
}

// HandleHandshakeAck 握手确认
func (p *Coder) HandleHandshakeAck() bool {
	if p.State != StateWaitAck {
		return false
	}
	p.State = StateWorking
	return true
}

// 这是一个确认值， 因为不需要服务端动态下发。
func genDictVersion() string {
	m := md5.Sum([]byte("{}"))
	return base64.StdEncoding.EncodeToString(m[:])
}

func (p *Coder) HandleData(b []byte) (c DecodedMsg) {
	if p.State != StateWorking {
		return
	}

	c = MessageDecode(b)
	return c

	//fmt.Println(c.Route, string(c.Body))
	//serverType := strings.SplitN(c.Route, ".", 2)[0]
	//if serverType == "" {
	//	return
	//}
	//sss, ok1 := global.RemoteTypeStore.Get(serverType)
	//if !ok1 {
	//	fmt.Println("no found server>>", serverType, c.Route)
	//	return
	//}
	//ssss, ok2 := sss.(*global.RemoteTypeStoreType)
	//if !ok2 {
	//	fmt.Println("parse server>>", serverType)
	//	return
	//}
	//fmt.Println("serverType:", serverType, "server.len:", len(ssss.Servers), ok1, ok2)
	//
	//userreq = global.UserReq{
	//	UID:        user.UID,
	//	Route:      c.Route,
	//	ServerType: serverType,
	//	Payload:    c.Body,
	//	PkgID:      c.ID,
	//	Sid:        user.Sid,
	//}
	//
	//if ssss.ForwardCh == nil {
	//	return
	//}
	//
	//fmt.Println("将要写入消息", serverType)
	//
	//select {
	//case ssss.ForwardCh <- userreq:
	//default:
	//	{
	//		fmt.Println("写入失败，队列堵塞", serverType)
	//	}
	//}
	//
	//return
}

type KickMsg struct {
	Reason string `json:"reason"`
}

var defaultKick = []byte(`{"reson":"kick"}`)

func (p *Coder) HandleKick(code int32, msg string, conn *websocket.Conn) {
	p.State = StateClosed
	r, err := json.Marshal(KickMsg{Reason: msg})
	if err != nil {
		r = defaultKick
	}
	b := PackageEncode(Package["TYPE_KICK"], r)
	if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		fmt.Println("err:", err, code)
		return
	}
}

func (p *Coder) Close() {
	p.State = StateClosed
}
