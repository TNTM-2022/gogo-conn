package main

import (
	"context"
	backEnd "go-connector/components/back_end"
	frontEnd "go-connector/components/front_end"
	"go-connector/components/monitor"
	"sync"
)

func main() {
	cancelCtx, cancelFn := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go monitor.StartMonitServer(cancelCtx, cancelFn, &wg)
	go backEnd.StartMqttServer(cancelCtx, cancelFn, &wg)
	go frontEnd.StartFrontServer()

	wg.Wait()
}
