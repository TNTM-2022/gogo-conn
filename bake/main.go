package main

import (
	"context"
	"fmt"
	"gogo-connector/components/monitor"
	"gogo-connector/components/mqtt/mqtt_server"
	"gogo-connector/components/ws_front"
	"log"
	"os"
	"os/signal"
	"sync"
)

func main() {
	ctx, cancelFn := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(3)
	fmt.Println("111")
	go ws_front.StartWsServer(ctx, &wg)
	go monitor.MonitServer(ctx, cancelFn, &wg)
	go mqtt_server.StartMqttServer(ctx, cancelFn, &wg)
	gracefulshutdown(ctx, cancelFn, &wg, monitor.QuitCtx)
}

func gracefulshutdown(ctx context.Context, cancelFn context.CancelFunc, wg *sync.WaitGroup, quitCtx context.Context) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	select {
	case <-quit:
	case <-quitCtx.Done():
		close(quit)
	}
	cancelFn()
	wg.Wait()
	log.Println("graceful shutdowning")
}
