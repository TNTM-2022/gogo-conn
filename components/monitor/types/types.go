package types

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
	Port       int    `json:"port"`
	ClientPort int    `json:"clientPort"`
	Frontend   string `json:"frontend"`
	ServerType string `json:"serverType"`
	Token      string `json:"token"`
	PID        int32  `json:"pid"`

	RemotePaths  []RegisterInfoRemoterPaths `json:"remotePaths"`
	HandlerPaths []string                   `json:"handlerPaths"`
}

type MonitorBody struct {
	Signal    string                  `json:"signal"`
	Action    string                  `json:"action"`
	Server    RegisterInfo            `json:"server"`
	ServerID  string                  `json:"id"`
	BlackList []string                `json:"blacklist"`
	Servers   map[string]RegisterInfo `json:"servers"`
}
type Monitor struct {
	RespId   int64       `json:"respId"`
	ReqID    int64       `json:"reqId"`
	ModuleID string      `json:"moduleId"`
	Body     MonitorBody `json:"body"`
	Command  string      `json:"command"`
}
