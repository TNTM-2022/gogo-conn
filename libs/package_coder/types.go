package package_coder

import (
	"encoding/json"
	concurrentMap "github.com/orcaman/concurrent-map"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

//type Settings struct { // map string json.rawMessage
//	RoomId int32         `json:"roomId,omitempty"`
//	GameId int32         `json:"gameId,omitempty"`
//	Pos    int32         `json:"pos,omitempty"`
//	DeskId float64       `json:"deskId,omitempty"`
//	UID    global.UserID `json:"uid,omitempty"`
//
//	Role        int32  `json:"role,omitempty"`
//	ChannelCode string `json:"channelCode,omitempty"`
//	BossId      int32  `json:"bossId,omitempty"`
//	Nickname    string `json:"nickname,omitempty"`
//	CreatedTime int64  `json:"createdTime,omitempty"`
//	Avatar      int32  `json:"avatar,omitempty"`
//	EntryAt     int64  `json:"entryAt,omitempty"`
//	ProxyIp     string `json:"proxyIp,omitempty"`
//	IP          string `json:"IP,omitempty"`
//}

type PkgPayloadInfo struct {
	PkgID uint64          `json:"id"`
	Route string          `json:"route"`
	Body  json.RawMessage `json:"body"`
	IsBf  bool            `json:"isBf,omitempty"` // 自定义添加的， 用来标识是不是透传 protobuf 给后端了
}

type PayloadMsg struct {
	Namespace  string             `json:"namespace,omitempty"`
	ServerType string             `json:"serverType,omitempty"`
	Service    string             `json:"service,omitempty"`
	Method     string             `json:"method,omitempty"`
	Args       [2]json.RawMessage `json:"args,omitempty"`
}

type Payload struct {
	Id  int64      `json:"id,omitempty"`
	Msg PayloadMsg `json:"msg,omitempty"`
}

type UserID = uint32
type pkgIdType struct {
	pkgid      int64
	countMutex sync.Mutex
}

type BackendMsg struct {
	//UID        UserID
	Route      string
	ServerType string
	Payload    []byte
	PkgID      uint64
	Sid        uint32
	//ServerId   string
	FrontServerId string
	Opts          json.RawMessage

	MType         byte `json:"type"`          // 请求类型  TYPE_REQUEST, TYPE_NOTIFY, ...
	CompressRoute bool `json:"compressRoute"` // 是否压缩陆游
	CompressGzip  bool `json:"compressGzip"`  // 是否使用 gzip
}

type PkgBelong struct {
	//UID         UserID // 这里不对， 没有鉴权 没有 uid， 只有sid
	SID         uint32
	StartAt     time.Time
	ClientPkgID uint64
	Route       string

	CompressRoute bool // 是否压缩陆游
	CompressGzip  bool // 是否使用 gzip
}

var pkgMap = concurrentMap.New() // 记录发往后端的 packageId
func (p *pkgIdType) genPkgId() int64 {
	p.countMutex.Lock()
	defer p.countMutex.Unlock()
	r := atomic.AddInt64(&p.pkgid, 1)
	if r > math.MaxInt64-1 {
		atomic.StoreInt64(&p.pkgid, 1)
		r = 1
	}
	return r
}
