package mqtt_server

import (
	"context"
	"encoding/json"
	"fmt"
	"go-connector/components/back_end/channel"
	"go-connector/components/back_end/session"
	"go-connector/config"
	mqtt "go-connector/libs/mqtt_server"
	"go-connector/libs/package_coder"
	"go-connector/logger"
	"log"
	"net"
	"strconv"
	"sync"
)

var s mqtt.Server

func init() {
	p := ""
	if config.MqttServerPort > 0 {
		p = fmt.Sprintf("%v", config.MqttServerPort)
	}
	fmt.Println(*config.MqttServerHost, p, fmt.Sprintf("%v:%v", *config.MqttServerHost, p))
	err := s.New(fmt.Sprintf("%v:%v", *config.MqttServerHost, p))
	if err != nil {
		log.Panicln(err)
	}
}

type reply struct {
	Id uint64 `json:"id"`
	//Resp json.RawMessage `json:"resp"`
}

func replyResponse(conn *mqtt.Conn, pkgId uint64, err string) {
	r := reply{
		Id: pkgId,
	}
	pkgIds, _err := json.Marshal(r)
	if _err != nil {
		err = _err.Error()
		fmt.Println(err)
	}
	if e := conn.Reply(pkgIds); e != nil {
		log.Println("reply error =", e)
	}
}
func StartMqttServer(ctx context.Context, f context.CancelFunc, wg *sync.WaitGroup) {
	defer f()
	defer wg.Done()

	h, p, err := net.SplitHostPort(s.Addr().String())
	port, _ := strconv.ParseInt(p, 10, 32)
	logger.DEBUG.Println("mqtt server =>>>", h, p, err, s.Addr(), port)
	config.MqttServerPort = int(port)
	if err != nil {
		log.Panicln(err)
	}
	//s.OnSubscribe(handleSubscribe)
	//s.OnUnSubscribe(handleUnSubscribe)
	s.OnPublish(func(conn *mqtt.Conn, _ string, _ uint16, b []byte) {
		logger.DEBUG.Println("server* ", string(b))

		var rec package_coder.RawRecv
		if e := json.Unmarshal(b, &rec); e != nil {
			fmt.Println("err: ", e)
			return
		}

		switch rec.Msg.Service {
		case "channelRemote":
			switch rec.Msg.Method {
			case "pushMessage":
				{
					pkgId, err := channel.PushMessage(&rec)
					replyResponse(conn, pkgId, err)
					return
				}
			}

		case "sessionRemote":
			switch rec.Msg.Method {
			case "pushAll":
				{
					pkgId, userId, settings := session.DecodePushAll(&rec)
					err := session.DoSave(userId, settings)
					replyResponse(conn, pkgId, err)
					fmt.Println("push all... ")
					return
				}
			case "push":
				{
					pkgId, userId, settings := session.DecodePush(&rec)
					err := session.DoSave(userId, settings)
					replyResponse(conn, pkgId, err)
					fmt.Println("push... ")
					return
				}
				// todo 删除键值
			}
		default:
			{

			}
		}
		fmt.Println("module not implemented")
	})
	<-ctx.Done()
}
