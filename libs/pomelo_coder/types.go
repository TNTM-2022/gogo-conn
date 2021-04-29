package pomelo_coder

// MType msg type
type MType struct {
	Type byte
	Body []byte
}

// DecodedMsg struct
type DecodedMsg struct {
	ID            uint64 `json:"id"`
	Type          byte   `json:"type"` // 请求类型  TYPE_REQUEST, TYPE_NOTIFY, ...
	CompressRoute bool   `json:"compressRoute"`
	Route         string `json:"route"`
	Body          []byte `json:"body"`
	CompressGzip  bool   `json:"compressGzip"`
}

//type UserConn struct {
//	// 客户端发送的消息
//	//MsgReq chan []byte // 不需要了 通过 connect 的chan 进行替代
//	// 客户端需要接收的消息
//	MsgResp chan []byte
//	// 服务器推送的消息
//	MsgPush chan []byte
//	Kick    chan []byte
//	//MsgSend chan []byte
//	State  int
//	Tick   time.Time
//	UID    global.UserID
//	Ctx    context.Context
//	Cancel context.CancelFunc
//	Sid    global.SessionID
//}

type handshake struct {
	Code int `json:"code"`
	Sys  sys `json:"sys"`
}

type sys struct {
	Heartbeat   int         `json:"heartbeat"`
	Dict        dict        `json:"dict"`
	RouteToCode routeToCode `json:"routeToCode"`
	CodeToRoute codeToRoute `json:"codeToRoute"`
	DictVersion string      `json:"dictVersion"`
	UseDict     bool        `json:"userDict"`
	UseProto    bool        `json:"userProto"`
}

type routeToCode struct{}
type codeToRoute struct{}
type dict struct{}
