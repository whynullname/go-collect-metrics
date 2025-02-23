package main

import (
	"math/rand/v2"
	"runtime"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
)

func main() {
	gaugeMetrics := make(map[string]float64)
	counterMetrics := make(map[string]int64)
	go updateMetrics(&gaugeMetrics, &counterMetrics)
	go sendMetrics(&gaugeMetrics, &counterMetrics)
}

func updateMetrics(gaugeMetrics *map[string]float64, counterMetrics *map[string]int64) {
	memStats := runtime.MemStats{}

	for {
		runtime.ReadMemStats(&memStats)
		(*gaugeMetrics)["Alloc"] = float64(memStats.Alloc)
		(*gaugeMetrics)["BuckHashSys"] = float64(memStats.BuckHashSys)
		(*gaugeMetrics)["Frees"] = float64(memStats.Frees)
		(*gaugeMetrics)["GCCPUFraction"] = float64(memStats.GCCPUFraction)
		(*gaugeMetrics)["GCSys"] = float64(memStats.GCSys)
		(*gaugeMetrics)["HeapAlloc"] = float64(memStats.HeapAlloc)
		(*gaugeMetrics)["HeapIdle"] = float64(memStats.HeapIdle)
		(*gaugeMetrics)["HeapInuse"] = float64(memStats.HeapInuse)
		(*gaugeMetrics)["HeapObjects"] = float64(memStats.HeapObjects)
		(*gaugeMetrics)["HeapReleased"] = float64(memStats.HeapReleased)
		(*gaugeMetrics)["HeapSys"] = float64(memStats.HeapSys)
		(*gaugeMetrics)["LastGC"] = float64(memStats.LastGC)
		(*gaugeMetrics)["Lookups"] = float64(memStats.Lookups)
		(*gaugeMetrics)["MCacheSys"] = float64(memStats.MCacheSys)
		(*gaugeMetrics)["Mallocs"] = float64(memStats.Mallocs)
		(*gaugeMetrics)["NextGC"] = float64(memStats.NextGC)
		(*gaugeMetrics)["NumForcedGC"] = float64(memStats.NumForcedGC)
		(*gaugeMetrics)["NumGC"] = float64(memStats.NumGC)
		(*gaugeMetrics)["OtherSys"] = float64(memStats.OtherSys)
		(*gaugeMetrics)["PauseTotalNs"] = float64(memStats.PauseTotalNs)
		(*gaugeMetrics)["StackInuse"] = float64(memStats.StackInuse)
		(*gaugeMetrics)["StackSys"] = float64(memStats.StackSys)
		(*gaugeMetrics)["Sys"] = float64(memStats.Sys)
		(*gaugeMetrics)["TotalAlloc"] = float64(memStats.TotalAlloc)
		(*gaugeMetrics)["RandomValue"] = rand.Float64()
		(*counterMetrics)["PollCount"]++
		time.Sleep(pollInterval * time.Second)
	}
}

func sendMetrics(gaugeMetrics *map[string]float64, counterMetrics *map[string]int64) {
	for {
		time.Sleep(reportInterval * time.Second)
	}
}
