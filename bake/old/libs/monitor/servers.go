package mqtt

import (
	"fmt"
	"log"
	"sync"

	userReqType "go-connector/common/types"
	cfg "go-connector/config"
	connect "go-connector/libs/mqtt_connect"
)

var ServerIDMap = make(map[string]*RegisterInfo)
var ServerTypeMap = make(map[string]map[string]*RegisterInfo)
var ServerConnectMap = make(map[string]*connect.Gate)
var ServerTypeChMap = make(map[string]chan userReqType.UserReq)
var ServerOptLocker sync.RWMutex

func AddServer(s *RegisterInfo) {
	if *cfg.ServerType == s.ServerType {
		fmt.Println("skip init same type server")
		return
	}

	ServerOptLocker.Lock()
	defer ServerOptLocker.Unlock()

	ServerIDMap[s.ServerID] = s
	if ServerTypeMap[s.ServerType] == nil {
		ServerTypeMap[s.ServerType] = make(map[string]*RegisterInfo, 10)
		ServerTypeChMap[s.ServerType] = make(chan userReqType.UserReq, 10000)
	}
	log.Println(">> add server", s.ServerID, s.ServerType)
	ServerTypeMap[s.ServerType][s.ServerID] = s

	gate := connect.Gate{
		GlobalForwardMsgChan: ServerTypeChMap[s.ServerType],
	}
	ServerConnectMap[s.ServerID] = &gate

	go func(s RegisterInfo) {
		fmt.Println("链接服务器", s.ServerID)
		if err := gate.StartGate(fmt.Sprintf("%s:%d", s.Host, s.Port), s.ServerID); err != nil {
			fmt.Println(err)
		}
		log.Println(fmt.Sprintf("connect to server: %s, host: %s, port: %d\n", s.ServerID, s.Host, s.Port))
	}(*s)
}
func AddServers(s map[string]RegisterInfo) {
	for _, ss := range s {
		sss := ss
		fmt.Println(ss.ServerType, ss.ServerID)
		go AddServer(&sss)
	}
}
func RemoveServer(id string) {
	ServerOptLocker.Lock()
	defer ServerOptLocker.Unlock()

	if s := ServerIDMap[id]; s != nil {
		log.Println("<< remove server", s.ServerID)
		delete(ServerIDMap, id)
		if ServerTypeMap[s.ServerType] != nil {
			delete(ServerTypeMap[s.ServerType], id)
			if len(ServerTypeMap[s.ServerType]) == 0 {
				delete(ServerTypeMap, s.ServerType)
			}
		}
	}
}
func GetServerByServerId(id string) RegisterInfo {
	var r RegisterInfo
	ServerOptLocker.RLock()
	defer ServerOptLocker.RUnlock()
	if s := ServerIDMap[id]; s != nil {
		r = *s
	}
	return r
}
func GetServersByServerType(t string) []RegisterInfo {
	var r []RegisterInfo
	ServerOptLocker.RLock()
	defer ServerOptLocker.RUnlock()
	if ServerTypeMap[t] != nil {
		for _, v := range ServerTypeMap[t] {
			r = append(r, *v)
		}
	}
	return r
}
