package online_user

import (
	"encoding/json"
	"go-connector/global"
	"go-connector/logger"
	"go.uber.org/zap"
)

type OnlineUserResp struct {
	ServerId       string `json:"serverId"`
	TotalConnCount int    `json:"totalConnCount"`
	LoginedCount   int    `json:"loginedCount"`
	//LoginedList    []mointor_watcher.UserReq `json:"loginedList"`
}

func MointorHandler(serverId string) (req, respBody, respErr, notify []byte) {
	res := OnlineUserResp{
		LoginedCount:   global.UidCount(),
		TotalConnCount: global.SessionsCount(),
		ServerId:       serverId,
	}
	notify, err := json.Marshal(res)
	if err != nil {
		logger.ERROR.Println("json.marshal error", zap.Error(err))
	}
	return
}
