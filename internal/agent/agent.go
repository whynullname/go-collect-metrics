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
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Agent struct {
	memStats       *runtime.MemStats
	Config         *config.AgentConfig
	Client         *http.Client
	metricsUseCase *metrics.MetricsUseCase
}

func NewAgent(memStats *runtime.MemStats, metricUseCase *metrics.MetricsUseCase, config *config.AgentConfig) *Agent {
	return &Agent{
		memStats:       memStats,
		metricsUseCase: metricUseCase,
		Config:         config,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (a *Agent) UpdateMetrics() {
	memStats := a.memStats
	runtime.ReadMemStats(memStats)
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "Alloc", float64(memStats.Alloc))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "Frees", float64(memStats.Frees))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "BuckHashSys", float64(memStats.BuckHashSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "GCCPUFraction", float64(memStats.GCCPUFraction))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "GCSys", float64(memStats.GCSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapAlloc", float64(memStats.HeapAlloc))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapIdle", float64(memStats.HeapIdle))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapInuse", float64(memStats.HeapInuse))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapObjects", float64(memStats.HeapObjects))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapReleased", float64(memStats.HeapReleased))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "HeapSys", float64(memStats.HeapSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "LastGC", float64(memStats.LastGC))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "Lookups", float64(memStats.Lookups))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "MCacheSys", float64(memStats.MCacheSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "Mallocs", float64(memStats.Mallocs))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "NextGC", float64(memStats.NextGC))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "NumForcedGC", float64(memStats.NumForcedGC))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "NumGC", float64(memStats.NumGC))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "OtherSys", float64(memStats.OtherSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "PauseTotalNs", float64(memStats.PauseTotalNs))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "StackInuse", float64(memStats.StackInuse))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "StackSys", float64(memStats.StackSys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "Sys", float64(memStats.Sys))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "TotalAlloc", float64(memStats.TotalAlloc))
	a.metricsUseCase.TryUpdateMetricValue(repository.GaugeMetricKey, "RandomValue", rand.Float64())

	a.metricsUseCase.TryUpdateMetricValue(repository.CounterMetricKey, "PollCount", int64(1))
}

func (a *Agent) SendMetrics() {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(repository.GaugeMetricKey)

	if err != nil {
		log.Fatal(err)
	}

	a.sendPostResponseWithMetrics(repository.GaugeMetricKey, gaugeMetrics)

	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(repository.CounterMetricKey)

	if err != nil {
		log.Fatal(err)
	}

	a.sendPostResponseWithMetrics(repository.CounterMetricKey, counterMetrics)
}

func (a *Agent) sendPostResponseWithMetrics(metricKey string, metrics map[string]any) {
	for k, v := range metrics {
		metricValue := ""

		switch value := v.(type) {
		case int64:
			metricValue = fmt.Sprintf("%d", value)
		case float64:
			metricValue = fmt.Sprintf("%.2f", value)
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.Config.EndPointAdress, metricKey, k, metricValue)
		resp, err := a.Client.Post(url, "text/plain", nil)
		resp.Body.Close()
		if err != nil {
			log.Printf("Can't send post method in %s ! Err %s \n", url, err)
			return
		}
	}
}
