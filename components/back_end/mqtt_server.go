package mqtt_server

import (
	"context"
	"encoding/json"
	"fmt"
	"go-connector/config"
	"go-connector/global"
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
	fmt.Print("test server start at ", s.Addr())
	//s.OnSubscribe(handleSubscribe)
	//s.OnUnSubscribe(handleUnSubscribe)
	s.OnPublish(func(conn *mqtt.Conn, _ string, _ uint16, b []byte) {
		logger.INFO.Println("server* ", string(b))
		log.Println("server* ", string(b))
		uids, pkgId, s := package_coder.DecodePush("", 0, b)
		pkgIds, _ := json.Marshal([]reply{reply{Id: pkgId}})
		conn.Reply(pkgIds)
		fmt.Println(uids)
		if len(uids) == 0 {
			global.SidFrontChanStore.IterCb(func(sid string, v interface{}) {
				if vv, ok := v.(chan package_coder.BackendMsg); ok {
					select {
					case vv <- *s:
					default:
						log.Printf("cannot write in. %v", sid)
					}
				} else {
					log.Printf("no sid chan ok, %v", sid)
				}
			})
		} else {
			for _, uid := range uids { // todo 后端传过来的全部是 uid， 需要根据 uid 传值
				if uid < 1 {
					continue
				}
				sid, ok := global.GetSidByUid(uid)
				if !ok {
					fmt.Println("no uid/sid found")
					continue
				}
				if v, ok := global.SidFrontChanStore.Get(strconv.FormatInt(int64(sid), 10)); ok {
					if vv, ok := v.(chan package_coder.BackendMsg); ok {
						select {
						case vv <- *s:
						default:
							log.Printf("cannot write in. %v", uid)
						}
					}
				}
			}
		}
		// todo 消息找到user 进行分发s
		log.Println("2222", s.Route, uids)
	})
	<-ctx.Done()
}
