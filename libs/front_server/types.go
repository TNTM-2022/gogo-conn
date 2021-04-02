package front_server

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
