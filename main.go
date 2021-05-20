package main

import (
	"context"
	"fmt"
	backEnd "go-connector/components/back_end"
	frontEnd "go-connector/components/front_end"
	"go-connector/components/monitor"
	_ "go-connector/libs/protobuf_coder"
	"sync"
)

func init() {
	fmt.Println("注意修改： pinus 问题， 没有 connack确认。 game-server/node_modules/pinus-rpc/dist/lib/rpc-server/acceptors/mqtt-acceptor.js 44L。 client.connack({ returnCode: 0 });")
}
func main() {
	cancelCtx, cancelFn := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	backEnd.StartMqttServer(cancelCtx, cancelFn, &wg)
	frontEnd.StartFrontServer()
	monitor.StartMonitServer(cancelCtx, cancelFn, &wg)

	//
	wg.Wait()
}
