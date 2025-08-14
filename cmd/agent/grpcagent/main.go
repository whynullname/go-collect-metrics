package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/whynullname/go-collect-metrics/internal/agent/grpcagent"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/rsareader"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	err := logger.Initialize("info")
	if err != nil {
		log.Fatalf("Fatal initialize logger")
		return
	}

	logger.Log.Infof("Build version: %s", buildVersion)
	logger.Log.Infof("Build date: %s", buildDate)
	logger.Log.Infof("Build commit: %s", buildCommit)

	cfg := config.NewAgentConfig()
	cfg.ParseFlags()

	if err := cfg.ReadRSA(); !errors.Is(err, rsareader.ErrEmptyKeyPath) {
		return
	}

	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	instance := grpcagent.NewGRPCAgent(metricsUseCase, cfg)
	err = instance.ConnetToServer()

	if err != nil {
		logger.Log.Fatalln(err)
	}

	defer instance.CloseConnection()
	logger.Log.Infof("Start agent, try work with server in %s \n", cfg.EndPointAdress)
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	var wg sync.WaitGroup
	wg.Add(2)
	go instance.UpdateMetrics(ctx, &wg)
	go instance.SendActualMetrics(ctx, &wg)
	<-exit
	cancel()
	wg.Wait()
	close(exit)
}
