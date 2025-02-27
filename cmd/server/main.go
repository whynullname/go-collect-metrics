package main

import (
	"log"

	"github.com/whynullname/go-collect-metrics/internal/server"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func main() {
	parseFlags()
	storage := storage.NewStorage()
	server := server.NewServer(storage, endPointAdress)

	log.Printf("Start server in %s \n", endPointAdress)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
