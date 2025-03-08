package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	EndPointAdress string
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
}

func (s *ServerConfig) checkEnvAddr() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.EndPointAdress = envRunAddr
	}
}
