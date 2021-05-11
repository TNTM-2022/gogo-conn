package channel

import (
	"encoding/json"
	"fmt"
	"go-connector/global"
	"go-connector/libs/package_coder"
	"go-connector/libs/pomelo_coder"
	"go-connector/logger"
	"log"
	"strconv"
)

func PushMessage(rec *package_coder.RawRecv) (pkgId uint64, error string) {
	userIds, pkgId, s := decodePushMsg(rec)
	if len(userIds) == 0 {
		global.SidFrontChanStore.IterCb(func(sid string, v interface{}) {
			if vv, ok := v.(chan package_coder.BackendMsg); ok {
				select {
				case vv <- *s:
				default:
					log.Printf("cannot write in. %v", sid)
				}
			} else {
				log.Printf("no sid chan ok, %v", sid)
			}
		})
	} else {
		failedId := make([]uint32, 0, len(userIds))
		notFoundId := make([]uint32, 0, len(userIds))
		for _, uid := range userIds { // 后端传过来的全部是 uid， 需要根据 uid 传值
			if uid < 1 {
				continue
			}
			sid, ok := global.GetSidByUid(uid)
			if !ok {
				fmt.Printf("no uid/sid found; uid:%v\n", uid)
				notFoundId = append(notFoundId, uid)
				continue
			}
			if v, ok := global.SidFrontChanStore.Get(strconv.FormatInt(int64(sid), 10)); ok {
				if vv, ok := v.(chan package_coder.BackendMsg); ok {
					select {
					case vv <- *s:
					default:
						log.Printf("cannot write in. %v", uid)
						failedId = append(failedId, uid)
					}
				}
			} else {
				notFoundId = append(notFoundId, uid)
			}
		}
		error = fmt.Sprintf(`"not found session: %v; failed write in: %v"`, failedId, notFoundId)
	}
	return
}

func decodePushMsg(rec *package_coder.RawRecv) (userIds []uint32, pkgId uint64, um *package_coder.BackendMsg) {
	um = &package_coder.BackendMsg{}
	pkgId = rec.Id
	if rec.Msg.Args != nil {
		um.Route, userIds, um.Payload, um.Opts = handlePushOrBroad(rec.Msg.Args)
		um.MType = pomelo_coder.Message["TYPE_PUSH"]
		if um.Route == "" {
			fmt.Println("no route; skip", userIds)
			return
		}
	}
	return
}

type MsgOptions struct {
	Type        string
	UserOptions json.RawMessage
	IsPush      bool
	IsBroadcast bool
}

func handlePushOrBroad(b []json.RawMessage) (route string, userIds []uint32, cc json.RawMessage, userOptions json.RawMessage) {
	if len(b) < 3 {
		return
	}

	var handleType MsgOptions
	if e := json.Unmarshal(b[len(b)-1], &handleType); e != nil {
		fmt.Println("error", e, b[len(b)-1])
	}
	if handleType.IsPush {
		if err := json.Unmarshal(b[2], &userIds); err != nil {
			logger.ERROR.Println(err)
		}
	}
	if err := json.Unmarshal(b[0], &route); err != nil {
		fmt.Println("error:", err)
		return
	}
	cc = b[1]
	userOptions = handleType.UserOptions
	return
}
