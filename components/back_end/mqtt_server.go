package mqtt_server

import (
	"context"
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
	err := s.New("")
	if err != nil {
		log.Panicln(err)
	}
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
	s.OnPublish(func(b []byte) {
		logger.INFO.Println("server* ", string(b))
		sids, s := package_coder.DecodePush("", 0, b)

		if len(sids) == 0 {
			global.SidFrontChanStore.IterCb(func(k string, v interface{}) {
				if vv, ok := v.(chan package_coder.BackendMsg); ok {
					vv <- *s
				} else {
					log.Printf("no sid chan ok")
				}
			})
		} else {
			for sid := range sids {
				fmt.Println(sid)
			}
		}

		// todo 消息找到user 进行分发s
		logger.ERROR.Println("2222", s.Route)
	})
	<-ctx.Done()
}
