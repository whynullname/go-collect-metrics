package main

import (
	"log"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/server"
)

func main() {
	cfg := config.NewServerConfig()
	cfg.ParseFlags()
	repo := inmemory.NewInMemoryRepository()
	server := server.NewServer(repo, cfg)

	log.Printf("Start server in %s \n", cfg.EndPointAdress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
