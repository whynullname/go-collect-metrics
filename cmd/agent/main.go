package main

import (
	"log"
	"runtime"
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
	updateAndSendMetrics(instance)
}

func updateAndSendMetrics(instance *agent.Agent) {
	secondPassed := time.Duration(0)

	for {
		log.Println("Update metrics")
		instance.UpdateMetrics()
		time.Sleep(pollInterval * time.Second)
		secondPassed += pollInterval * time.Second

		if secondPassed >= time.Duration(reportInterval*time.Second) {
			secondPassed = time.Duration(0)
			log.Println("Send metrics")
			instance.SendMetrics()
		}
	}
}
