package mqtt

import (
	"encoding/json"
	"fmt"

	proto "github.com/huin/mqtt"
)

type RegisterInfoRemoterPaths struct {
	Namespace  string `json:"namespace"`
	ServerType string `json:"serverType"`
	Path       string `json:"path"`
}
type RegisterInfo struct {
	Main       string `json:"main"`
	Env        string `json:"env"`
	ServerID   string `json:"id"`
	Host       string `json:"host"`
	Port       int32  `json:"port"`
	ClientPort int32  `json:"clientPort"`
	Frontend   string `json:"frontend"`
	ServerType string `json:"serverType"`
	Token      string `json:"token"`
	PID        int32  `json:"pid"`

	RemotePaths  []RegisterInfoRemoterPaths `json:"remotePaths"`
	HandlerPaths []string                   `json:"handlerPaths"`
}
type MqttCmd struct {
	ClientId string `json:"clientId"`
	Cmd      string `json:"cmd"`
}

type Pub struct {
	Header    proto.Header
	TopicName string
	MessageId uint16
	Payload   []byte
}

type Register struct {
	ServerID   string       `json:"id"`
	Type       string       `json:"type"`
	ServerType string       `json:"serverType"`
	PID        int32        `json:"pid"`
	Info       RegisterInfo `json:"info"`
	Token      string       `json:"token"`
}
type RegisterResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type SubscribeBody struct {
	Action   string `json:"action"`
	ServerID string `json:"id"`
}
type Subscribe struct {
	ReqID    int64         `json:"reqId"`
	ModuleID string        `json:"moduleId"`
	Body     SubscribeBody `json:"body"`
}

type MonitorBody struct {
	Signal   string       `json:"signal"`
	Action   string       `json:"action"`
	Server   RegisterInfo `json:"server"`
	ServerID string       `json:"id"`
}
type Monitor struct {
	RespId   int64       `json:"respId"`
	ReqID    int64       `json:"reqId"`
	ModuleID string      `json:"moduleId"`
	Body     MonitorBody `json:"body"`
	Command  string      `json:"command"`
}

type MonitorServers map[string]RegisterInfo
type MonitorAllServer struct {
	RespId   int64          `json:"respId"`
	ReqID    int64          `json:"reqId"`
	ModuleID string         `json:"moduleId"`
	Body     MonitorServers `json:"body"`
	Command  string         `json:"command"`
}

func DecodeMonitor(d []byte) Monitor {
	var mm Monitor
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		fmt.Println(mm, e)
	}
	return mm
}
func DecodeMonitorAllServer(d []byte) MonitorAllServer {
	var mm MonitorAllServer
	var ss string

	if e := json.Unmarshal(d, &ss); e == nil {
		d = []byte(ss)
	}
	if e := json.Unmarshal(d, &mm); e != nil {
		fmt.Println(mm, e)
	}
	return mm
}

type MonitListInfoBody struct {
	ServerID   string  `json:"serverId"`
	ServerType string  `json:"serverType"`
	Pid        int     `json:"pid"`
	RSS        uint64  `json:"rss"`
	HeapTotal  uint64  `json:"heapTotal"`
	HeapUsed   uint64  `json:"heapUsed"`
	Uptime     float64 `json:"uptime"`
}
type MonitListInfo struct {
	ServerID string            `json:"serverId"`
	Body     MonitListInfoBody `json:"body"`
}

type MonitListInfoRes struct {
	RespID int64         `json:"respId"`
	Error  MonitListInfo `json:"error"`
}

type MonitRespOk struct {
	RespID int64           `json:"respId"`
	Body   MonitorBody `json:"body"`
}

type ClientActionRes struct {
	RespId int64 `json:"respId"`
	Error int32 `json:"error"`

}