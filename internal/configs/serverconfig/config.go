package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"net"
	"os"
	"strconv"

	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/rsareader"
)

type ServerConfig struct {
	EndPointAdress    string
	StoreInterval     uint64
	FileStoragePath   string
	RestoreData       bool
	PostgressAdress   string
	HashKey           string
	RSAPrivateKeyPath string
	RSAKey            *rsa.PrivateKey
	configPath        string
	trustedSubnet     string
	TrustedSubnet     *net.IPNet
}

type jsonConfig struct {
	Adress            string `json:"address"`
	RestoreData       bool   `json:"restore"`
	StoreInterval     uint64 `json:"store_interval"`
	StoreFilePath     string `json:"store_file"`
	PostgressAdress   string `json:"database_dsn"`
	RSAPrivateKeyPath string `json:"crypto_key"`
	TrustedSubnet     string `json:"trusted_subnet"`
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
	s.readConfigFile()
	s.parseCIDR()
}

func (s *ServerConfig) ReadRSA() error {
	key, err := rsareader.ReadPrivateRSAKey(s.RSAPrivateKeyPath)
	if err != nil {
		return err
	}

	s.RSAKey = key
	return nil
}

func (s *ServerConfig) registerFlags() {
	flag.StringVar(&s.EndPointAdress, "a", "localhost:8080", "address and port to run server")
	flag.Uint64Var(&s.StoreInterval, "i", 300, "interval to save all metrics to file")
	flag.StringVar(&s.FileStoragePath, "f", "metrics.json", "path to save metrics")
	flag.BoolVar(&s.RestoreData, "r", true, "need load saved data in start")
	flag.StringVar(&s.PostgressAdress, "d", "", "adress to connect postgres")
	flag.StringVar(&s.HashKey, "k", "", "key for sha hash")
	flag.StringVar(&s.RSAPrivateKeyPath, "crypto-key", "", "path to RSA private key")
	flag.StringVar(&s.configPath, "c", "", "path to json config")
	flag.StringVar(&s.configPath, "config", "", "path to json config")
	flag.StringVar(&s.trustedSubnet, "t", "", "subnet string")
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

	if postgresAdress := os.Getenv("DATABASE_DSN"); postgresAdress != "" {
		s.PostgressAdress = postgresAdress
	}

	if hashKey := os.Getenv("KEY"); hashKey != "" {
		s.HashKey = hashKey
	}

	if cfgPath := os.Getenv("CONFIG"); cfgPath != "" {
		s.configPath = cfgPath
	}

	if subnet := os.Getenv("TRUSTED_SUBNET"); subnet != "" {
		s.trustedSubnet = subnet
	}
}

func (s *ServerConfig) readConfigFile() {
	if s.configPath == "" {
		return
	}

	cfgFile, err := os.Open(s.configPath)
	if err != nil {
		logger.Log.Errorf("error wile open cfg file %v\n", err)
		return
	}

	defer cfgFile.Close()
	var cfg jsonConfig
	decoder := json.NewDecoder(cfgFile)
	if err := decoder.Decode(&cfg); err != nil {
		logger.Log.Errorf("error wile json decode cfg file %v\n", err)
		return
	}

	if s.EndPointAdress == "localhost:8080" {
		s.EndPointAdress = cfg.Adress
	}

	if s.RestoreData {
		s.RestoreData = cfg.RestoreData
	}

	if s.StoreInterval == 300 {
		s.StoreInterval = cfg.StoreInterval
	}

	if s.FileStoragePath == "metrics.json" {
		s.FileStoragePath = cfg.StoreFilePath
	}

	if s.PostgressAdress == "" {
		s.PostgressAdress = cfg.PostgressAdress
	}

	if s.RSAPrivateKeyPath == "" {
		s.RSAPrivateKeyPath = cfg.RSAPrivateKeyPath
	}

	if s.trustedSubnet == "" {
		s.trustedSubnet = cfg.TrustedSubnet
	}
}

func (s *ServerConfig) parseCIDR() {
	if s.trustedSubnet == "" {
		return
	}
	_, ipnet, err := net.ParseCIDR(s.trustedSubnet)
	if err != nil {
		logger.Log.Error("Invalid CIDR format in trusted_subnet: %v", err)
		return
	}

	s.TrustedSubnet = ipnet
}
