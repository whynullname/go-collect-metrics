package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/rsareader"
)

type AgentConfig struct {
	EndPointAdress   string
	ReportInterval   int
	PollInterval     int
	HashKey          string
	RateLimit        int
	RSAPublicKeyPath string
	RSAKey           *rsa.PublicKey
	configPath       string
}

type jsonConfig struct {
	Adress           string `json:"address"`
	ReportInterval   int    `json:"report_interval"`
	PollInterval     int    `json:"poll_interval"`
	RSAPublicKeyPath string `json:"crypto_key"`
}

func NewAgentConfig() *AgentConfig {
	cfg := AgentConfig{}
	cfg.setDefaultValues()
	return &cfg
}

func (a *AgentConfig) setDefaultValues() {
	a.EndPointAdress = "localhost:8080"
	a.ReportInterval = 10
	a.PollInterval = 2
}

func (a *AgentConfig) ParseFlags() {
	a.registerFlags()
	flag.Parse()
	a.checkEnv()
	a.readConfigFile()
}

func (a *AgentConfig) ReadRSA() error {
	key, err := rsareader.ReadPublicRSAKey(a.RSAPublicKeyPath)
	if err != nil {
		return err
	}

	a.RSAKey = key
	return nil
}

func (a *AgentConfig) registerFlags() {
	flag.StringVar(&a.EndPointAdress, "a", "localhost:8080", "address and port to server for send metrics")
	flag.IntVar(&a.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&a.PollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.StringVar(&a.HashKey, "k", "", "key for sha hash")
	flag.IntVar(&a.RateLimit, "l", 1, "rate limit goroutines to send metrics")
	flag.StringVar(&a.RSAPublicKeyPath, "crypto-key", "", "path to RSA public key")
	flag.StringVar(&a.configPath, "c", "", "path to json config")
	flag.StringVar(&a.configPath, "config", "", "path to json config")
}

func (a *AgentConfig) checkEnv() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		a.EndPointAdress = envRunAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		i, err := strconv.ParseInt(envReportInterval, 10, 32)

		if err != nil {
			a.ReportInterval = int(i)
		}
	}

	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		i, err := strconv.ParseInt(envPoolInterval, 10, 32)

		if err != nil {
			a.PollInterval = int(i)
		}
	}

	if hashKey := os.Getenv("KEY"); hashKey != "" {
		a.HashKey = hashKey
	}

	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		i, err := strconv.ParseInt(rateLimit, 10, 32)

		if err != nil {
			a.PollInterval = int(i)
		}
	}

	if keyPath := os.Getenv("CRYPTO_KEY"); keyPath != "" {
		a.RSAPublicKeyPath = keyPath
	}
}

func (a *AgentConfig) readConfigFile() {
	if a.configPath == "" {
		return
	}

	cfgFile, err := os.Open(a.configPath)
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

	if a.EndPointAdress == "localhost:8080" {
		a.EndPointAdress = cfg.Adress
	}

	if a.ReportInterval == 10 {
		a.ReportInterval = cfg.ReportInterval
	}

	if a.PollInterval == 2 {
		a.PollInterval = cfg.PollInterval
	}

	if a.RSAPublicKeyPath == "" {
		a.RSAPublicKeyPath = cfg.RSAPublicKeyPath
	}
}
