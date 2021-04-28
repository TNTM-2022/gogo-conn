package online_user

import (
	"encoding/json"
	"fmt"
	"gogo-connector/components/global"
	"gogo-connector/components/monitor/mointor_watcher"
)

type OnlineUserResp struct {
	ServerId       string                    `json:"serverId"`
	TotalConnCount int                       `json:"totalConnCount"`
	LoginedCount   int                       `json:"loginedCount"`
	LoginedList    []mointor_watcher.UserReq `json:"loginedList"`
}

func MointorHandler(serverId string) (req, respBody, respErr, notify []byte) {
	res := OnlineUserResp{
		LoginedCount:   global.Users.Count(),
		TotalConnCount: global.Sids.Count(),
		ServerId:       serverId,
	}
	notify, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
	}
	return
}
