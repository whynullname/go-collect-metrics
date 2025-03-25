package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/whynullname/go-collect-metrics/internal/logger"
)

type ServerConfig struct {
	EndPointAdress  string
	StoreInterval   uint64
	FileStoragePath string
	RestoreData     bool
}

func NewServerConfig() *ServerConfig {
	cfg := ServerConfig{}
	cfg.setDefaultsValues()
	return &cfg
}

func (s *ServerConfig) setDefaultsValues() {
	s.EndPointAdress = "localhost:8080"
}

func (s *ServerConfig) ParseFlags() {
	s.registerFlags()
	flag.Parse()
	s.checkEnvAddr()
}

func (s *ServerConfig) registerFlags() {
	flag.StringVar(&s.EndPointAdress, "a", "localhost:8080", "address and port to run server")
	flag.Uint64Var(&s.StoreInterval, "i", 300, "interval to save all metrics to file")
	flag.StringVar(&s.FileStoragePath, "f", "metrics.json", "path to save metrics")
	flag.BoolVar(&s.RestoreData, "r", true, "need load saved data in start")
}

func (s *ServerConfig) checkEnvAddr() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.EndPointAdress = envRunAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		interval, err := strconv.ParseUint(storeInterval, 10, 64)

		if err != nil {
			logger.Log.Errorf("Can't parse STORE_INTERVAL env! Error %s", err.Error())
			return
		}

		s.StoreInterval = interval
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		s.FileStoragePath = fileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restore, err := strconv.ParseBool(envRestore)

		if err != nil {
			logger.Log.Errorf("Can't parse RESTORE env! Error %s", err.Error())
			return
		}

		s.RestoreData = restore
	}
}
