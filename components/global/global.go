package global

import (
	concurrentMap "github.com/orcaman/concurrent-map"
	"sync"
)

const UserCap = 10

type SessionID uint32
type UserID uint32
type ServerID string

type sessionPool struct {
	pool   []SessionID
	locker sync.RWMutex
}
type UserChannel struct {
	ServerId  ServerID  `json:"sv"`
	SessionId SessionID `json:"sn"`
}

var Users = concurrentMap.New() // make(map[uint64]int32)
var Sids = concurrentMap.New()

func init() {
	for i := 0; i < UserCap; i++ {
		//fmt.Println(i, cap(sidPool.pool))
		sidPool.pool = append(sidPool.pool, SessionID(i))
	}
}

// session/sid 缓冲池，预防sid一直增大直到溢出.
var sidPool = &sessionPool{
	make([]SessionID, UserCap, UserCap+1),
	sync.RWMutex{},
}

func GetSid() (sid SessionID, ok bool) {
	sidPool.locker.Lock()
	defer sidPool.locker.Unlock()
	if len(sidPool.pool) == 0 {
		return 0, false
	}
	sid, sidPool.pool = sidPool.pool[0], sidPool.pool[1:]
	return sid, true
}
func BackSid(sid SessionID) {
	sidPool.locker.Lock()
	defer sidPool.locker.Unlock()
	sidPool.pool = append(sidPool.pool, sid)
	return
}
func GetOnlineUserNum() int {
	sidPool.locker.RLock()
	defer sidPool.locker.RUnlock()
	return len(sidPool.pool)
}
