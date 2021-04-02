package main

import (
	"context"
	types "go-connector/interfaces"
	"go-connector/libs/front_server"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func getTCPListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:12345")
	return l, err
}

func main() {
	ctx, cancelFn := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	mc := types.MainControl{ctx, &wg}

	l, err := getTCPListener()
	if err != nil {
		log.Fatal(err)
	}
	host, port, err := net.SplitHostPort(l.Addr().String())
	log.Println("ws listen on port:", host, port)

	wg.Add(1)
	go front_server.StartWsServer(mc, l)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	cancelFn()
	wg.Wait()
	log.Println("graceful shutdowning")
}
