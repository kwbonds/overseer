package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func waitForSignalInterrupt() {
	endWaiter := sync.WaitGroup{}
	endWaiter.Add(1)

	onSignalInterrupt(func() {
		endWaiter.Done()
	})

	endWaiter.Wait()
}

func onSignalInterrupt(fn func()) {
	// We listen for SIGTERM, SIGINT, to please k8s and keyboard users.
	onSignals(fn, syscall.SIGINT, syscall.SIGTERM)
}

func onSignals(fn func(), sig ...os.Signal) {
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, sig...)
		<-signalCh

		fn()
	}()
}

func indent(text, indent string) string {
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}
