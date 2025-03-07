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
	a.storage.SetGaugeMetricValue("Alloc", float64(memStats.Alloc))
	a.storage.SetGaugeMetricValue("Frees", float64(memStats.Frees))
	a.storage.SetGaugeMetricValue("BuckHashSys", float64(memStats.BuckHashSys))
	a.storage.SetGaugeMetricValue("GCCPUFraction", float64(memStats.GCCPUFraction))
	a.storage.SetGaugeMetricValue("GCSys", float64(memStats.GCSys))
	a.storage.SetGaugeMetricValue("HeapAlloc", float64(memStats.HeapAlloc))
	a.storage.SetGaugeMetricValue("HeapIdle", float64(memStats.HeapIdle))
	a.storage.SetGaugeMetricValue("HeapInuse", float64(memStats.HeapInuse))
	a.storage.SetGaugeMetricValue("HeapObjects", float64(memStats.HeapObjects))
	a.storage.SetGaugeMetricValue("HeapReleased", float64(memStats.HeapReleased))
	a.storage.SetGaugeMetricValue("HeapSys", float64(memStats.HeapSys))
	a.storage.SetGaugeMetricValue("LastGC", float64(memStats.LastGC))
	a.storage.SetGaugeMetricValue("Lookups", float64(memStats.Lookups))
	a.storage.SetGaugeMetricValue("MCacheSys", float64(memStats.MCacheSys))
	a.storage.SetGaugeMetricValue("Mallocs", float64(memStats.Mallocs))
	a.storage.SetGaugeMetricValue("NextGC", float64(memStats.NextGC))
	a.storage.SetGaugeMetricValue("NumForcedGC", float64(memStats.NumForcedGC))
	a.storage.SetGaugeMetricValue("NumGC", float64(memStats.NumGC))
	a.storage.SetGaugeMetricValue("OtherSys", float64(memStats.OtherSys))
	a.storage.SetGaugeMetricValue("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.storage.SetGaugeMetricValue("StackInuse", float64(memStats.StackInuse))
	a.storage.SetGaugeMetricValue("StackSys", float64(memStats.StackSys))
	a.storage.SetGaugeMetricValue("Sys", float64(memStats.Sys))
	a.storage.SetGaugeMetricValue("TotalAlloc", float64(memStats.TotalAlloc))
	a.storage.SetGaugeMetricValue("RandomValue", rand.Float64())

	a.storage.AddValueToCounterMetric("PollCount", 1)
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
