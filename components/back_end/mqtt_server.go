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
	err := s.New("127.0.0.1:44155")
	if err != nil {
		log.Panicln(err)
	}
}

type reply struct {
	Id uint64 `json:"id"`
	//Resp json.RawMessage `json:"resp"`
}

func replyResponse(conn *mqtt.Conn, pkgId uint64, error string) {
	fmt.Println(error)
	r := reply{
		Id: pkgId,
	}
	pkgIds, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("reply", pkgId, error, string(pkgIds))
	_ = conn.Reply(pkgIds)
}
func StartMqttServer(ctx context.Context, f context.CancelFunc, wg *sync.WaitGroup) {
	defer f()
	defer wg.Done()

	h, p, err := net.SplitHostPort(s.Addr().String())
	port, _ := strconv.ParseInt(p, 10, 32)
	fmt.Println("mqtt server =>>>", h, p, err, s.Addr(), port)
	config.MqttServerPort = int(port)
	if err != nil {
		log.Panicln(err)
	}
	//s.OnSubscribe(handleSubscribe)
	//s.OnUnSubscribe(handleUnSubscribe)
	s.OnPublish(func(conn *mqtt.Conn, _ string, _ uint16, b []byte) {
		log.Println("server* ", string(b))
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
					fmt.Println("push message")
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
