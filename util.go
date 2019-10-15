package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func WaitForCtrlC() {
	endWaiter := sync.WaitGroup{}
	endWaiter.Add(1)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-signalChannel
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
