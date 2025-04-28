package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	instance := agent.NewAgent(metricsUseCase, cfg)

	logger.Log.Infof("Start agent, try work with server in %s \n", cfg.EndPointAdress)
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	go instance.UpdateMetrics(ctx)
	go instance.SendActualMetrics(ctx)
	<-exit
	cancel()
}
