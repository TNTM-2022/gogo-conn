package config

import (
	"fmt"
	"math"
	"os"
	"time"
)

var (
	_tmp            = ""
	ProtoPath       = &_tmp
	_serverid       = fmt.Sprintf("connector-%v", time.Now().Format("200601021504")) //奇葩， 必须是这个时间点
	ServerID        = &_serverid
	_serverType     = "connector"
	ServerType      = &_serverType
	_e              = "production"
	Env             = &_e // flag.String("env", "development", "env")
	_MqttServerHost = "127.0.0.1"
	MqttServerHost  = &_MqttServerHost
	MqttServerPort  = 0
	WsServerPort    = 12345
	Pid             = os.Getpid()
	startTick       = time.Now()
)

func Uptime() float64 {
	return math.Floor(time.Now().Sub(startTick).Seconds()/60*100) / 100
}
