package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	storage "github.com/whynullname/go-collect-metrics/internal/storage"
)

type Agent struct {
	memStats *runtime.MemStats
	storage  *storage.MemoryStorage
	Config   *config.AgentConfig
	Client   *http.Client
}

func NewAgent(memStats *runtime.MemStats, storage *storage.MemoryStorage, config *config.AgentConfig) *Agent {
	return &Agent{
		memStats: memStats,
		storage:  storage,
		Config:   config,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (a *Agent) UpdateMetrics() {
	memStats := a.memStats
	runtime.ReadMemStats(memStats)
	a.storage.UpdateGaugeMetric("Alloc", float64(memStats.Alloc))
	a.storage.UpdateGaugeMetric("Frees", float64(memStats.Frees))
	a.storage.UpdateGaugeMetric("BuckHashSys", float64(memStats.BuckHashSys))
	a.storage.UpdateGaugeMetric("GCCPUFraction", float64(memStats.GCCPUFraction))
	a.storage.UpdateGaugeMetric("GCSys", float64(memStats.GCSys))
	a.storage.UpdateGaugeMetric("HeapAlloc", float64(memStats.HeapAlloc))
	a.storage.UpdateGaugeMetric("HeapIdle", float64(memStats.HeapIdle))
	a.storage.UpdateGaugeMetric("HeapInuse", float64(memStats.HeapInuse))
	a.storage.UpdateGaugeMetric("HeapObjects", float64(memStats.HeapObjects))
	a.storage.UpdateGaugeMetric("HeapReleased", float64(memStats.HeapReleased))
	a.storage.UpdateGaugeMetric("HeapSys", float64(memStats.HeapSys))
	a.storage.UpdateGaugeMetric("LastGC", float64(memStats.LastGC))
	a.storage.UpdateGaugeMetric("Lookups", float64(memStats.Lookups))
	a.storage.UpdateGaugeMetric("MCacheSys", float64(memStats.MCacheSys))
	a.storage.UpdateGaugeMetric("Mallocs", float64(memStats.Mallocs))
	a.storage.UpdateGaugeMetric("NextGC", float64(memStats.NextGC))
	a.storage.UpdateGaugeMetric("NumForcedGC", float64(memStats.NumForcedGC))
	a.storage.UpdateGaugeMetric("NumGC", float64(memStats.NumGC))
	a.storage.UpdateGaugeMetric("OtherSys", float64(memStats.OtherSys))
	a.storage.UpdateGaugeMetric("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.storage.UpdateGaugeMetric("StackInuse", float64(memStats.StackInuse))
	a.storage.UpdateGaugeMetric("StackSys", float64(memStats.StackSys))
	a.storage.UpdateGaugeMetric("Sys", float64(memStats.Sys))
	a.storage.UpdateGaugeMetric("TotalAlloc", float64(memStats.TotalAlloc))
	a.storage.UpdateGaugeMetric("RandomValue", rand.Float64())

	a.storage.UpdateCounterMetric("PollCount")
}

func (a *Agent) SendMetrics() {
	for k, v := range a.storage.GetAllGaugeMetrics() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%.2f", a.Config.EndPointAdress, storage.GaugeKey, k, v)
		resp, err := a.Client.Post(url, "text/plain", nil)
		resp.Body.Close()
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}

	for k, v := range a.storage.GetAllCounterMetrics() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%d", a.Config.EndPointAdress, storage.CounterKey, k, v)
		resp, err := a.Client.Post(url, "text/plain", nil)
		resp.Body.Close()
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}
}
