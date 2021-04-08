package package_coder

import (
	"encoding/json"
	"fmt"
	concurrentMap "github.com/orcaman/concurrent-map"
	"gogo-connector/components/global"
	sendProto "gogo-connector/libs/proto_coder/protos/forward_proto"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type pkgIdType struct {
	pkgid      int64
	countMutex sync.Mutex
}

type UserReq = global.UserReq
type PkgBelong = global.PkgBelong

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

var pkgId pkgIdType
var pkgMap = concurrentMap.New() // 记录发往后端的 packageId
func Encode(userReq *UserReq, serverId string) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()
	pkgId := pkgId.genPkgId()
	pkgMap.Set(strconv.FormatInt(pkgId, 10), &PkgBelong{
		UID:         userReq.UID,
		StartAt:     time.Now(),
		ClientPkgID: userReq.PkgID,
		Route:       userReq.Route,
	})
	fmt.Println(userReq.Sid, userReq.UID)
	m := &sendProto.Payload{
		Id: pkgId, // 通过 此处的 id 进行路由对应的user

		Msg: &sendProto.PayloadMsg{
			Namespace:  "sys",
			ServerType: userReq.ServerType,
			Service:    "msgRemote",
			Method:     "forwardMessage",
			Args: []*sendProto.PayloadMsgArgs{
				&sendProto.PayloadMsgArgs{
					Id:    userReq.PkgID, // 客户端包
					Route: userReq.Route,
					Body:  userReq.Payload,
				},
				&sendProto.PayloadMsgArgs{
					Id:         int64(userReq.Sid), // sid
					FrontendId: serverId,
					Uid:        userReq.UID,
					Settings: &sendProto.Settings{
						GameId: 0,
						RoomId: 1,
						UID:    userReq.UID,
						DeskId: 10000000010041,
					},
				},
			},
		},
	}

	//_, _ = pb.Marshal(m)
	// if j, e := pb.Marshal(m); e == nil {
	// 	return j
	// }
	if j, e := json.Marshal(m); e == nil {
		fmt.Println("......", string(j))
		return j
	}

	return nil
}
