package online_user

import (
	"encoding/json"
	"fmt"
	"go-connector/global"
)

type OnlineUserResp struct {
	ServerId       string `json:"serverId"`
	TotalConnCount int    `json:"totalConnCount"`
	LoginedCount   int    `json:"loginedCount"`
	//LoginedList    []mointor_watcher.UserReq `json:"loginedList"`
}

func MointorHandler(serverId string) (req, respBody, respErr, notify []byte) {
	n := global.SessionsCount()
	res := OnlineUserResp{
		LoginedCount:   n,
		TotalConnCount: n,
		ServerId:       serverId,
	}
	notify, err := json.Marshal(res)
	if err != nil {
		fmt.Println("err: ", err)
	}
	return
}
