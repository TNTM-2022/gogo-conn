package main

import (
	"fmt"
	mqtt "go-connector/libs/mqtt_server"
)

func main() {
	var s mqtt.Server

	if err := s.New(":1883"); err != nil {
		return
	}
	fmt.Print("test server start at ", s.Addr())
	s.OnSubscribe(handleSubscribe)
	s.OnUnSubscribe(handleUnSubscribe)
	s.OnPublish(handlePublish)
	select {}

}

func handleSubscribe(m string) {
	fmt.Println("test - sub", m)
}

func handleUnSubscribe(m string) {
	fmt.Println("test - unsub", m)
}

func handlePublish(m []byte) {
	fmt.Println("test - pub", m)
}
