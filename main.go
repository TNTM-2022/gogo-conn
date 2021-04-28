package main

import (
	"context"
	"go-connector/components/monitor"
	"sync"
)

func main() {
	cancelCtx, cancelFn := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	monitor.StartMonitServer(cancelCtx, cancelFn, &wg)

	wg.Wait()
}
