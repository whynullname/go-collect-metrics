package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"

	storage "github.com/whynullname/go-collect-metrics/internal/storage"
)

type Agent struct {
	memStats     *runtime.MemStats
	storage      *storage.MemoryStorage
	serverAdress string
}

func NewAgent(memStats *runtime.MemStats, storage *storage.MemoryStorage, serverAdress string) *Agent {
	return &Agent{
		memStats:     memStats,
		storage:      storage,
		serverAdress: serverAdress,
	}
}

func (a *Agent) UpdateMetrics() {
	memStats := a.memStats
	runtime.ReadMemStats(memStats)
	a.storage.UpdateGaugeData("Alloc", float64(memStats.Alloc))
	a.storage.UpdateGaugeData("Frees", float64(memStats.Frees))
	a.storage.UpdateGaugeData("BuckHashSys", float64(memStats.BuckHashSys))
	a.storage.UpdateGaugeData("GCCPUFraction", float64(memStats.GCCPUFraction))
	a.storage.UpdateGaugeData("GCSys", float64(memStats.GCSys))
	a.storage.UpdateGaugeData("HeapAlloc", float64(memStats.HeapAlloc))
	a.storage.UpdateGaugeData("HeapIdle", float64(memStats.HeapIdle))
	a.storage.UpdateGaugeData("HeapInuse", float64(memStats.HeapInuse))
	a.storage.UpdateGaugeData("HeapObjects", float64(memStats.HeapObjects))
	a.storage.UpdateGaugeData("HeapReleased", float64(memStats.HeapReleased))
	a.storage.UpdateGaugeData("HeapSys", float64(memStats.HeapSys))
	a.storage.UpdateGaugeData("LastGC", float64(memStats.LastGC))
	a.storage.UpdateGaugeData("Lookups", float64(memStats.Lookups))
	a.storage.UpdateGaugeData("MCacheSys", float64(memStats.MCacheSys))
	a.storage.UpdateGaugeData("Mallocs", float64(memStats.Mallocs))
	a.storage.UpdateGaugeData("NextGC", float64(memStats.NextGC))
	a.storage.UpdateGaugeData("NumForcedGC", float64(memStats.NumForcedGC))
	a.storage.UpdateGaugeData("NumGC", float64(memStats.NumGC))
	a.storage.UpdateGaugeData("OtherSys", float64(memStats.OtherSys))
	a.storage.UpdateGaugeData("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.storage.UpdateGaugeData("StackInuse", float64(memStats.StackInuse))
	a.storage.UpdateGaugeData("StackSys", float64(memStats.StackSys))
	a.storage.UpdateGaugeData("Sys", float64(memStats.Sys))
	a.storage.UpdateGaugeData("TotalAlloc", float64(memStats.TotalAlloc))
	a.storage.UpdateGaugeData("RandomValue", rand.Float64())

	val, ok := a.storage.GetCounterData("PollCount")

	if !ok {
		val = 0
	} else {
		val++
	}

	a.storage.UpdateCounterData("PollCount", val)
}

func (a *Agent) SendMetrics() {
	for k, v := range a.storage.GetAllGaugeData() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%.2f", a.serverAdress, storage.GaugeKey, k, v)
		resp, err := http.Post(url, "text/plain", nil)

		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}

		defer resp.Body.Close()
	}

	for k, v := range a.storage.GetAllCounterData() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%d", a.serverAdress, storage.CounterKey, k, v)
		resp, err := http.Post(url, "text/plain", nil)

		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}

		defer resp.Body.Close()
	}
}
