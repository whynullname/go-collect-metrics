package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type Agent struct {
	memStats   *runtime.MemStats
	repository repository.Repository
	Config     *config.AgentConfig
	Client     *http.Client
}

func NewAgent(memStats *runtime.MemStats, repository repository.Repository, config *config.AgentConfig) *Agent {
	return &Agent{
		memStats:   memStats,
		repository: repository,
		Config:     config,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (a *Agent) UpdateMetrics() {
	memStats := a.memStats
	runtime.ReadMemStats(memStats)
	a.repository.UpdateGaugeMetricValue("Alloc", float64(memStats.Alloc))
	a.repository.UpdateGaugeMetricValue("Frees", float64(memStats.Frees))
	a.repository.UpdateGaugeMetricValue("BuckHashSys", float64(memStats.BuckHashSys))
	a.repository.UpdateGaugeMetricValue("GCCPUFraction", float64(memStats.GCCPUFraction))
	a.repository.UpdateGaugeMetricValue("GCSys", float64(memStats.GCSys))
	a.repository.UpdateGaugeMetricValue("HeapAlloc", float64(memStats.HeapAlloc))
	a.repository.UpdateGaugeMetricValue("HeapIdle", float64(memStats.HeapIdle))
	a.repository.UpdateGaugeMetricValue("HeapInuse", float64(memStats.HeapInuse))
	a.repository.UpdateGaugeMetricValue("HeapObjects", float64(memStats.HeapObjects))
	a.repository.UpdateGaugeMetricValue("HeapReleased", float64(memStats.HeapReleased))
	a.repository.UpdateGaugeMetricValue("HeapSys", float64(memStats.HeapSys))
	a.repository.UpdateGaugeMetricValue("LastGC", float64(memStats.LastGC))
	a.repository.UpdateGaugeMetricValue("Lookups", float64(memStats.Lookups))
	a.repository.UpdateGaugeMetricValue("MCacheSys", float64(memStats.MCacheSys))
	a.repository.UpdateGaugeMetricValue("Mallocs", float64(memStats.Mallocs))
	a.repository.UpdateGaugeMetricValue("NextGC", float64(memStats.NextGC))
	a.repository.UpdateGaugeMetricValue("NumForcedGC", float64(memStats.NumForcedGC))
	a.repository.UpdateGaugeMetricValue("NumGC", float64(memStats.NumGC))
	a.repository.UpdateGaugeMetricValue("OtherSys", float64(memStats.OtherSys))
	a.repository.UpdateGaugeMetricValue("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.repository.UpdateGaugeMetricValue("StackInuse", float64(memStats.StackInuse))
	a.repository.UpdateGaugeMetricValue("StackSys", float64(memStats.StackSys))
	a.repository.UpdateGaugeMetricValue("Sys", float64(memStats.Sys))
	a.repository.UpdateGaugeMetricValue("TotalAlloc", float64(memStats.TotalAlloc))
	a.repository.UpdateGaugeMetricValue("RandomValue", rand.Float64())

	a.repository.UpdateCounterMetricValue("PollCount", 1)
}

func (a *Agent) SendMetrics() {
	for k, v := range a.repository.GetAllGaugeMetrics() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%.2f", a.Config.EndPointAdress, repository.GaugeMetricKey, k, v)
		resp, err := a.Client.Post(url, "text/plain", nil)
		resp.Body.Close()
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}

	for k, v := range a.repository.GetAllCounterMetrics() {
		url := fmt.Sprintf("http://%s/update/%s/%s/%d", a.Config.EndPointAdress, repository.CounterMetricKey, k, v)
		resp, err := a.Client.Post(url, "text/plain", nil)
		resp.Body.Close()
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}
}
