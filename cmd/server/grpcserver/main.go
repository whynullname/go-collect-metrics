package main

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/repository/postgres"
	"github.com/whynullname/go-collect-metrics/internal/rsareader"
	"github.com/whynullname/go-collect-metrics/internal/server/grpcserver"
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

	if err := cfg.ReadRSA(); !errors.Is(err, rsareader.ErrEmptyKeyPath) {
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
	server := grpcserver.NewGrpcServer(metricsUseCase, cfg)
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

	exit := make(chan os.Signal, 1)
	idleConnChan := make(chan struct{}, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	if err := server.ListenServer(exit, idleConnChan); err != nil {
		log.Fatal(err)
	}

	<-idleConnChan
	close(exit)
	close(idleConnChan)
}
