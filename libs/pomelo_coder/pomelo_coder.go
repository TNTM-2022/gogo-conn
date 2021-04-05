package coder

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

// user conn 状态
const (
	StateInited  = 0
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

func PackageTypeHandler(t byte, b []byte, user *UserConn) {
	switch int(t) {
	case Package["TYPE_HANDSHAKE"]:
		handleHandshake(user)

	case Package["TYPE_HANDSHAKE_ACK"]:
		handleHandshakeAck(user)

	case Package["TYPE_HEARTBEAT"]:
		fmt.Println("TYPE_HEARTBEAT")

	case Package["TYPE_DATA"]:
		fmt.Println("TYPE_DATA")
		handleData(user, b)

	case Package["TYPE_KICK"]:
		fmt.Println("TYPE_KICK")

	default:
	}
	return
}

// 处理客户端handshake
// 检查整体状态 ST_INITED
// todo checkClient
// var opts = { heartbeat : setupHeartbeat(this) };
//  opts.useProto = true;
// 返回 TYPE_HANDSHAKE 报文， 携带上述的对象 packageEncode
func handleHandshake(user *UserConn) {
	if user.State != StateInited {
		user.Cancel()
		return
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
	p := PackageEncode(Package["TYPE_HANDSHAKE"], []byte(string(j)))
	user.MsgPush <- p
	user.State = StateWaitAck
	// fmt.Println("handshacke handler = 0")
}

//
func handleHandshakeAck(user *UserConn) {
	if user.State != StateWaitAck {
		user.Cancel()
		return
	}
	user.State = StateWorking
}

func genDictVersion() string {
	m := md5.Sum([]byte("{}"))
	return base64.StdEncoding.EncodeToString(m[:])
}

func handleData(user *UserConn, b []byte) {
	if user.State != StateWorking {
		user.Cancel()
		return
	}

	c := MessageDecode(b)
	fmt.Println(c.Route, string(c.Body))
	//server := strings.SplitN(c.Route, ".", 2)[0]
	//fmt.Println("servertype", server, "server.len", len(servers.ServerTypeMap[server]))
	//if server != "" {
	//	servers.ServerOptLocker.RLock()
	//	ch := servers.ServerTypeChMap[server]
	//	servers.ServerOptLocker.RUnlock()
	//	if ch != nil {
	//		fmt.Println("将要写入消息", server)
	//		//ch <- userReqType.UserReq{
	//		//	UID:        user.UID,
	//		//	Route:      c.Route,
	//		//	ServerType: server,
	//		//	Payload:    c.Body,
	//		//	PkgID:      c.ID,
	//		//	Sid:        user.Sid,
	//		//}
	//		m := interfaces.UserReq{
	//			UID:        user.UID,
	//			Route:      c.Route,
	//			ServerType: server,
	//			Payload:    c.Body,
	//			PkgID:      c.ID,
	//			Sid:        user.Sid,
	//		}
	//		select {
	//		case ch <- m:
	//		default:
	//			{
	//				fmt.Println("写入失败，队列堵塞", server)
	//			}
	//		}
	//		fmt.Println("将要写入消息 done")
	//
	//	}
	//}
}

type KickMsg struct {
	Reason string `json:"reason"`
}

var defaultKick = []byte(`{"reson":"kick"}`)

func handleKick(code int32, msg string, conn *websocket.Conn) {
	r, err := json.Marshal(KickMsg{Reason: msg})
	if err != nil {
		r = defaultKick
	}
	b := PackageEncode(Package["TYPE_KICK"], r)
	if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		fmt.Println(err)
		return
	}
}
