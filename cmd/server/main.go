package main

import (
	"log"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/server"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func main() {
	cfg := config.NewServerConfig()
	cfg.ParseFlags()
	storage := storage.NewStorage()
	server := server.NewServer(storage, cfg)

	log.Printf("Start server in %s \n", cfg.EndPointAdress)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
