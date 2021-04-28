package config

import (
	"math"
	"os"
	"time"
)

var _tmp = ""
var ProtoPath = &_tmp
var _serverid = "connector-D"
var ServerID = &_serverid
var _serverType = "connector"
var ServerType = &_serverType

var Pid = os.Getpid()

var _e = "development"
var Env = &_e // flag.String("env", "development", "env")

var _MqttServerHost = "127.0.0.1"
var MqttServerHost = &_MqttServerHost

var MqttServerPort = 8080
var WsServerPort = 12345

var startTick = time.Now()

func Uptime() float64 {
	return math.Floor(time.Now().Sub(startTick).Seconds()/60*100) / 100
}
