package session

import (
	"encoding/json"
	"fmt"
	"go-connector/global"
	"go-connector/libs/package_coder"
)

func PushAll(rec *package_coder.RawRecv) (pkgId uint64, error string) {
	error = "user not found"
	pkgId, userId, settings := decodePushAll(rec)
	sid, ok := global.GetSidByUid(userId)
	if !ok {

		return
	}
	fmt.Println("set session", pkgId)
	session, ok := global.GetSessionBySid(sid)
	if !ok {
		return
	}
	if session.Uid != userId {
		return
	}
	session.Set(settings)
	return
}

func decodePushAll(rec *package_coder.RawRecv) (pkgId uint64, userId uint32, settings map[string]json.RawMessage) {
	pkgId = rec.Id
	if rec.Msg.Args != nil {
		if err := json.Unmarshal(rec.Msg.Args[0], &userId); err != nil {
			fmt.Println("unmarshal session uid", err)
			return
		}
		if err := json.Unmarshal(rec.Msg.Args[1], &settings); err != nil {
			fmt.Println("unmarshal session settings", err)
		}
	}
	return
}
