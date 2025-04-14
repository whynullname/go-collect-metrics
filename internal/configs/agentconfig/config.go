package config

import (
	"flag"
	"os"
	"strconv"
)

type AgentConfig struct {
	EndPointAdress string
	ReportInterval int
	PollInterval   int
	HashKey        string
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
}

func (a *AgentConfig) registerFlags() {
	flag.StringVar(&a.EndPointAdress, "a", "localhost:8080", "address and port to server for send metrics")
	flag.IntVar(&a.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&a.PollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.StringVar(&a.HashKey, "k", "", "key for sha hash")
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
}
