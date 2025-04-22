package main

import (
	"log"
	"runtime"
	"sync"

	"github.com/whynullname/go-collect-metrics/internal/agent"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func main() {
	err := logger.Initialize("info")

	if err != nil {
		log.Fatalf("Fatal initialize logger")
		return
	}

	cfg := config.NewAgentConfig()
	cfg.ParseFlags()

	memStats := runtime.MemStats{}
	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)

	instance := agent.NewAgent(&memStats, metricsUseCase, cfg)
	logger.Log.Infof("Start agent, try work with server in %s \n", cfg.EndPointAdress)
	var wg sync.WaitGroup
	wg.Add(2)
	go instance.UpdateMetrics()
	go instance.SendActualMetrics()
	wg.Wait()
}
