package global

import (
	"context"
	concurrentMap "github.com/orcaman/concurrent-map"
)

var BlackList = concurrentMap.New()

var QuitCtx, QuitFn = context.WithCancel(context.Background())

var Sids = concurrentMap.New()
var Session = concurrentMap.New() // sid >> session
var UidSid = concurrentMap.New()  // uid >> sid  修改 后端推送 问题

var RemoteBackendTypeForwardChan = concurrentMap.New() // serverType ->> chan backendMsg
var RemoteBackendClients = concurrentMap.New()         // serverType ->> concurrencymap[serverId] >> serverInfo
