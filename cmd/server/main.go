package main

import (
	"log"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/server"
	"github.com/whynullname/go-collect-metrics/internal/storage/filestorage"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func main() {
	cfg := config.NewServerConfig()
	cfg.ParseFlags()
	repo := inmemory.NewInMemoryRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	server := server.NewServer(metricsUseCase, cfg)
	fileStorage, err := filestorage.NewFileStorage(cfg.FileStoragePath)

	if err != nil {
		logger.Log.Errorf("Fail initialize file storage! Error: %s", err.Error())
		return
	}

	if cfg.RestoreData {
		fileStorage.ReadAllMetrics(repo)
	}

	go fileStorage.RecordMetric(cfg.StoreInterval, repo)

	logger.Log.Infof("Start server in %s \n", cfg.EndPointAdress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
