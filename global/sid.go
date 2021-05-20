package global

import (
	"math"
	"sync"
)

//func init() {
//	for i := 1; i <= UserCap; i++ {
//		//fmt.Println(i, cap(sidPool.pool))
//		//sidPool.pool = append(sidPool.pool, uint32(i))
//		sidPool.pool[UserCap-i] = uint32(i)
//	}
//	fmt.Println(" user cap >", len(sidPool.pool), cap(sidPool.pool), sidPool.pool)
//}
//
//// session/sid 缓冲池，预防sid一直增大直到溢出.
//var sidPool = &struct {
//	pool   []uint32
//	locker sync.RWMutex
//}{
//	make([]uint32, UserCap, UserCap+1),
//	sync.RWMutex{},
//}
//
//func GetSid() (sid uint32, ok bool) {
//	sidPool.locker.Lock()
//	defer sidPool.locker.Unlock()
//	if len(sidPool.pool) == 0 {
//		return 0, false
//	}
//	sid, sidPool.pool = sidPool.pool[len(sidPool.pool)-1], sidPool.pool[:len(sidPool.pool)-1]
//	return sid, true
//}
//func BackSid(sid uint32) { // 内存可能会出现问题， 最好用 栈 方式
//	sidPool.locker.Lock()
//	defer sidPool.locker.Unlock()
//	sidPool.pool = append(sidPool.pool, sid)
//	//sidPool.pool = append(sidPool.pool, sid)
//	return
//}

var sidP uint32
var sidLocker sync.Mutex

func GetSid() (sid uint32) {
	sidLocker.Lock()
	defer sidLocker.Unlock()
	sidP++
	if sidP > math.MaxUint32-1 {
		sidP = 1
	}
	sid = sidP
	return
}
