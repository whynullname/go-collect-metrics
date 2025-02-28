package main

import (
	"log"
	"runtime"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/agent"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func main() {
	parseFlags()
	log.Printf("Start agent, try work with server in %s \n", serverEndPointAdress)
	memStats := runtime.MemStats{}
	storage := storage.NewStorage()
	instance := agent.NewAgent(&memStats, storage, serverEndPointAdress)
	updateAndSendMetrics(instance)
}

func updateAndSendMetrics(instance *agent.Agent) {
	secondPassed := time.Duration(0)

	for {
		log.Println("Update metrics")
		instance.UpdateMetrics()
		sleepDuration := time.Duration(pollInterval) * time.Second
		time.Sleep(sleepDuration)
		secondPassed += sleepDuration

		if secondPassed >= time.Duration(reportInterval)*time.Second {
			secondPassed = time.Duration(0)
			log.Println("Send metrics")
			instance.SendMetrics()
		}
	}
}
