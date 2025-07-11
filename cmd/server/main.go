package main

import (
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/repository/postgres"
	"github.com/whynullname/go-collect-metrics/internal/server"
	"github.com/whynullname/go-collect-metrics/internal/storage/filestorage"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"

	_ "net/http/pprof"

	_ "github.com/jackc/pgx/v5/stdlib"
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

	cfg := config.NewServerConfig()
	cfg.ParseFlags()

	if err := readRSAKey(cfg); err != nil {
		return
	}

	var repo repository.Repository
	if cfg.PostgressAdress == "" {
		repo = inmemory.NewInMemoryRepository()
	} else {
		repo, err = postgres.NewPostgresRepo(cfg.PostgressAdress)
		if err != nil {
			logger.Log.Fatalln(err)
			return
		}
	}
	defer repo.CloseRepository()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	server := server.NewServer(metricsUseCase, cfg, repo.PingRepo)
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

func readRSAKey(cfg *config.ServerConfig) error {
	if cfg.RSAPrivateKeyPath != "" {
		body, err := os.ReadFile(cfg.RSAPrivateKeyPath)
		if err != nil {
			logger.Log.Errorf("Error while read RSA private key: %v\n", err)
			return err
		}

		block, _ := pem.Decode(body)
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

		if err != nil {
			logger.Log.Errorf("Error while parse RSA private key: %v\n", err)
			return err
		}
		cfg.RSAKey = key
	}

	return nil
}
