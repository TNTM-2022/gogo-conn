package global

import (
	cmap "github.com/orcaman/concurrent-map"
)

var Users = cmap.New() // make(map[uint64]int32)
var Sids = cmap.New()

type UserChannel struct {
	ServerId  string `json:"sv"`
	SessionId uint64 `json:"sn"`
}

func (ch *UserChannel) CreateUserChannel () {

}


func (ch *UserChannel) DeleteUserChannel () {

}

