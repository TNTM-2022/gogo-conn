package session

import (
	"encoding/json"
	"fmt"
	"go-connector/global"
	"go-connector/libs/package_coder"
)

func DoSave(userId uint32, settings map[string]json.RawMessage) (error string) {
	error = "user not found"
	sid, ok := global.GetSidByUid(userId)
	if !ok {

		return
	}
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

func DecodePush(rec *package_coder.RawRecv) (pkgId uint64, userId uint32, settings map[string]json.RawMessage) {
	pkgId = rec.Id
	settings = make(map[string]json.RawMessage)
	var key string
	if rec.Msg.Args != nil {
		if err := json.Unmarshal(rec.Msg.Args[0], &userId); err != nil {
			fmt.Println("unmarshal session uid", err)
			return
		}
		if err := json.Unmarshal(rec.Msg.Args[1], &key); err != nil {
			fmt.Println("unmarshal session settings", err)
			return
		}
		settings[key] = rec.Msg.Args[2]
	}
	return
}
func DecodePushAll(rec *package_coder.RawRecv) (pkgId uint64, userId uint32, settings map[string]json.RawMessage) {
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
