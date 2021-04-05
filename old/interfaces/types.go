package interfaces

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Sid uint64
type UserId uint64

type UserInfo struct {
	UserBase

	Nickname    string      `json:"nickname"`
	CreatedTime time.Time   `json:"createdTime"`
	Avatar      interface{} `json:"avatar"`
	ProxyIp     string      `json:"proxyIp"`
	IP          string      `json:"IP"`
}

type UserBase struct {
	UID         int64  `json:"uid"`
	DeskId      int64  `json:"deskId"`
	Pos         int32  `json:"pos"`
	Role        int8   `json:"role"`
	ChannelCode string `json:"channelCode"`
	BossId      int32  `json:"bossId"`
	EntryAt     int64  `json:"entryAt"`
}

type tokenParsed struct {
	UID int32 `json:"d"`
	jwt.StandardClaims
}

type BindRes struct {
	Ok   bool   `json:"ok"`
	Sid  string `json:"sid"`
	Fid  string `json:"frontendId"`
	Info string `json:"info"`
}

type dict struct{}
type routeToCode struct{}
type codeToRoute struct{}
type sys struct {
	Heartbeat   int         `json:"heartbeat"`
	Dict        dict        `json:"dict"`
	RouteToCode routeToCode `json:"routeToCode"`
	CodeToRoute codeToRoute `json:"codeToRoute"`
	DictVersion string      `json:"dictVersion"`
	UseDict     bool        `json:"userDict"`
	UseProto    bool        `json:"userProto"`
}
type handshake struct {
	Code int `json:"code"`
	Sys  sys `json:"sys"`
}

type UserParams struct {
	UID        int32  `json:"uid"`
	SID        uint64 `json:"sid"`
	FrontEndId string `json:"frontendId"`
	Headers    string `json:"headers"`
	RealIp     string `json:"realIp"`
	ProxyIp    string `json:"proxyIp"`
	LastSid    string `json:"lastSid"`
	LastFront  string `json:"lastFront"`
}

// UserConn 保存用户收发消息
type UserConn struct {
	// 客户端发送的消息
	//MsgReq chan []byte // 不需要了 通过 connect 的chan 进行替代
	// 客户端需要接收的消息
	MsgResp chan []byte
	// 服务器推送的消息
	MsgPush chan []byte
	Kick    chan []byte
	//MsgSend chan []byte
	State  int
	Tick   time.Time
	UID    UserId
	Ctx    context.Context
	Cancel context.CancelFunc
	Sid    Sid
}

func CreateUserConn (uid UserId, sid Sid, ctx context.Context, cancel context.CancelFunc, state int) *UserConn {
	return &UserConn{
		//MsgReq:  make(chan []byte, 1000),
		MsgResp: make(chan []byte, 1000),
		MsgPush: make(chan []byte, 1000),
		Kick:    make(chan []byte),
		//MsgSend: make(chan []byte, 1000),

		Tick:   time.Now(),
		UID:    uid,
		Ctx:    ctx,
		Cancel: cancel,
		State:  state,
		Sid:    sid,
	}
}

type UserReq struct {
	UID        UserId
	Route      string
	ServerType string
	Payload    []byte
	PkgID      int64
	Sid        Sid
}
