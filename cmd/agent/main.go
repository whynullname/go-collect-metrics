package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/whynullname/go-collect-metrics/internal/agent"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
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

	if err := readRSAKey(cfg); err != nil {
		return
	}

	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	instance := agent.NewAgent(metricsUseCase, cfg)

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

func readRSAKey(cfg *config.AgentConfig) error {
	if cfg.RSAPublicKeyPath != "" {
		body, err := os.ReadFile(cfg.RSAPublicKeyPath)
		if err != nil {
			logger.Log.Errorf("Error while read RSA public key: %v\n", err)
			return err
		}

		block, _ := pem.Decode(body)
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)

		if err != nil {
			logger.Log.Errorf("Error while parse RSA public key: %v\n", err)
			return err
		}
		cfg.RSAKey = key
	}

	return nil
}
