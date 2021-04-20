package mqtt_server

import (
	"context"
	"fmt"
	"gogo-connector/components/config"
	"gogo-connector/libs/mqtt"
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
	//config.WsServerPort = int(port)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Print("test server start at ", s.Addr())
	//s.OnSubscribe(handleSubscribe)
	//s.OnUnSubscribe(handleUnSubscribe)
	s.OnPublish(func(b []byte) {
		fmt.Println(string(b))
	})
	<-ctx.Done()
}
