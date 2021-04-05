package rpc

import "encoding/json"

type Settings struct {
	RoomId int32   `json:"roomId,omitempty"`
	GameId int32   `json:"gameId,omitempty"`
	Pos    int32   `json:"pos,omitempty"`
	DeskId float64 `json:"deskId,omitempty"`
	UID    int32   `json:"uid,omitempty"`

	Role        int32  `json:"role,omitempty"`
	ChannelCode string `json:"channelCode,omitempty"`
	BossId      int32  `json:"bossId,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	CreatedTime int64  `json:"createdTime,omitempty"`
	Avatar      int32  `json:"avatar,omitempty"`
	EntryAt     int64  `json:"entryAt,omitempty"`
	ProxyIp     string `json:"proxyIp,omitempty"`
	IP          string `json:"IP,omitempty"`
}

type PayloadMsgArgs struct {
	Id         int64           `json:"id,omitempty"`
	Route      string          `json:"route,omitempty"`
	Body       json.RawMessage `json:"body,omitempty"`
	FrontendId string          `json:"frontendId,omitempty"`
	Uid        int32           `json:"uid,omitempty"`
	Settings   *Settings       `json:"settings,omitempty"`
	IsBf       bool            `json:"isBf,omitempty"`
}

type PayloadMsg struct {
	Namespace  string            `json:"namespace,omitempty"`
	ServerType string            `json:"serverType,omitempty"`
	Service    string            `json:"service,omitempty"`
	Method     string            `json:"method,omitempty"`
	Args       []*PayloadMsgArgs `json:"args,omitempty"`
}

type Payload struct {
	Id  int64       `json:"id,omitempty"`
	Msg *PayloadMsg `json:"msg,omitempty"`
}
