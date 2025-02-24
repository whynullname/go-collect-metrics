package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	storage "github.com/whynullname/go-collect-metrics/internal/storage"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	log.Println("Start agent")
	var wg sync.WaitGroup
	wg.Add(1)
	go updateMetrics()
	go sendMetrics()
	wg.Wait()
}

func updateMetrics() {
	memStats := runtime.MemStats{}

	for {
		runtime.ReadMemStats(&memStats)
		storage.MemoryStorage.UpdateGaugeData("Alloc", float64(memStats.Alloc))
		storage.MemoryStorage.UpdateGaugeData("BuckHashSys", float64(memStats.BuckHashSys))
		storage.MemoryStorage.UpdateGaugeData("Frees", float64(memStats.Frees))
		storage.MemoryStorage.UpdateGaugeData("GCCPUFraction", float64(memStats.GCCPUFraction))
		storage.MemoryStorage.UpdateGaugeData("GCSys", float64(memStats.GCSys))
		storage.MemoryStorage.UpdateGaugeData("HeapAlloc", float64(memStats.HeapAlloc))
		storage.MemoryStorage.UpdateGaugeData("HeapIdle", float64(memStats.HeapIdle))
		storage.MemoryStorage.UpdateGaugeData("HeapInuse", float64(memStats.HeapInuse))
		storage.MemoryStorage.UpdateGaugeData("HeapObjects", float64(memStats.HeapObjects))
		storage.MemoryStorage.UpdateGaugeData("HeapReleased", float64(memStats.HeapReleased))
		storage.MemoryStorage.UpdateGaugeData("HeapSys", float64(memStats.HeapSys))
		storage.MemoryStorage.UpdateGaugeData("LastGC", float64(memStats.LastGC))
		storage.MemoryStorage.UpdateGaugeData("Lookups", float64(memStats.Lookups))
		storage.MemoryStorage.UpdateGaugeData("MCacheSys", float64(memStats.MCacheSys))
		storage.MemoryStorage.UpdateGaugeData("Mallocs", float64(memStats.Mallocs))
		storage.MemoryStorage.UpdateGaugeData("NextGC", float64(memStats.NextGC))
		storage.MemoryStorage.UpdateGaugeData("NumForcedGC", float64(memStats.NumForcedGC))
		storage.MemoryStorage.UpdateGaugeData("NumGC", float64(memStats.NumGC))
		storage.MemoryStorage.UpdateGaugeData("OtherSys", float64(memStats.OtherSys))
		storage.MemoryStorage.UpdateGaugeData("PauseTotalNs", float64(memStats.PauseTotalNs))
		storage.MemoryStorage.UpdateGaugeData("StackInuse", float64(memStats.StackInuse))
		storage.MemoryStorage.UpdateGaugeData("StackSys", float64(memStats.StackSys))
		storage.MemoryStorage.UpdateGaugeData("Sys", float64(memStats.Sys))
		storage.MemoryStorage.UpdateGaugeData("TotalAlloc", float64(memStats.TotalAlloc))
		storage.MemoryStorage.UpdateGaugeData("RandomValue", rand.Float64())

		val, ok := storage.MemoryStorage.GetCounterData("PollCount")

		if !ok {
			val = 0
		} else {
			val++
		}

		storage.MemoryStorage.UpdateCounterData("PollCount", val)

		time.Sleep(pollInterval * time.Second)
	}
}

func sendMetrics() {
	for {
		fmt.Println("Send metrics")
		for k, v := range storage.MemoryStorage.GetAllGaugeData() {
			url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%.2f", storage.GaugeKey, k, v)
			http.Post(url, "text/plain", nil)
		}

		for k, v := range storage.MemoryStorage.GetAllCounterData() {
			url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", storage.CounterKey, k, v)
			http.Post(url, "text/plain", nil)
		}

		time.Sleep(reportInterval * time.Second)
	}
}
