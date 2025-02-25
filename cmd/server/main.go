package main

import (
	"log"

	"github.com/whynullname/go-collect-metrics/internal/server"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func main() {
	storage := storage.NewStorage()
	server := server.NewServer(storage)

	log.Println("Start server")

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
