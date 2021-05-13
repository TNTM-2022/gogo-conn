package package_coder

import (
	"encoding/json"
	"fmt"
	"go-connector/global"
	"go-connector/logger"
	"go.uber.org/zap"
)

func Encode(pkgId int64, u *BackendMsg) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from ", r)
		}
	}()

	pkgContentInfo, _ := json.Marshal(PkgPayloadInfo{
		PkgID: u.PkgID, // 客户端包
		Route: u.Route,
		Body:  u.Payload,
	})

	_session, ok := global.GetSessionBySid(u.Sid)
	if !ok {
		logger.ERROR.Println("no session found", zap.Uint32("userId", u.Sid))
		return nil
	}
	session, _ := json.Marshal(_session)

	m := &Payload{
		Id: pkgId,
		Msg: PayloadMsg{
			Namespace:  "sys",
			ServerType: u.ServerType, // 远程 server type // todo 确认 是 本server 还是目标server
			Service:    "msgRemote",
			Method:     "forwardMessage",
			Args: [2]json.RawMessage{
				pkgContentInfo,
				session,
			},
		},
	}

	if j, e := json.Marshal(m); e == nil {
		return j
	}

	return nil
}

//RawRecv {"id":1,"msg":{"namespace":"sys","service":"channelRemote","method":"broadcast","args":["broadcast.test",{"isPush":true},{"type":"broadcast","userOptions":{},"isBroadcast":true}]}}
//RawRecv {"id":0,"msg":{"namespace":"sys","service":"channelRemote","method":"pushMessage","args":["push.push",{"type":"push","is_broad":true},[1],{"type":"push","userOptions":{},"isPush":true}]}}
type RawRecv struct {
	Id  uint64 `json:"id"`
	Msg struct {
		Namespace string            `json:"namespace"`
		Service   string            `json:"service"`
		Method    string            `json:"method"`
		Args      []json.RawMessage `json:"args"`
	} `json:"msg"`
	Resp []json.RawMessage `json:"resp"`
}
