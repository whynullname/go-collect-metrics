package main

import (
	"log"

	"github.com/whynullname/go-collect-metrics/internal/server"
)

func main() {
	server := server.NewServer()

	log.Println("Start server")

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
