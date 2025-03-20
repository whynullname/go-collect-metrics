package main

import (
	"log"
	"runtime"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/agent"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func main() {
	cfg := config.NewAgentConfig()
	cfg.ParseFlags()
	log.Printf("Start agent, try work with server in %s \n", cfg.EndPointAdress)

	memStats := runtime.MemStats{}
	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)

	instance := agent.NewAgent(&memStats, metricsUseCase, cfg)
	updateAndSendMetrics(instance)
}

func updateAndSendMetrics(instance *agent.Agent) {
	secondPassed := time.Duration(0)

	for {
		log.Println("Update metrics")
		instance.UpdateMetrics()
		sleepDuration := time.Duration(instance.Config.PollInterval) * time.Second
		time.Sleep(sleepDuration)
		secondPassed += sleepDuration

		if secondPassed >= time.Duration(instance.Config.ReportInterval)*time.Second {
			secondPassed = time.Duration(0)
			log.Println("Send metrics")
			instance.SendMetricsByJSON()
		}
	}
}
