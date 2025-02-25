package main

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/agent"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	log.Println("Start agent")
	memStats := runtime.MemStats{}
	storage := storage.NewStorage()
	instance := agent.NewAgent(&memStats, storage)
	var wg sync.WaitGroup
	wg.Add(1)
	go sendMetrics(instance)
	go updateMetrics(instance)
	wg.Wait()
}

func updateMetrics(instance *agent.Agent) {
	for {
		log.Println("Update metrics")
		instance.UpdateMetrics()
		time.Sleep(pollInterval * time.Second)
	}
}

func sendMetrics(instance *agent.Agent) {
	for {
		log.Println("Send metrics")
		instance.SendMetrics()
		time.Sleep(reportInterval * time.Second)
	}
}
