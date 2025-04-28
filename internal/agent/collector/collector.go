package collector

import (
	"context"
	"math/rand/v2"
	"runtime"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type AgentCollector struct {
	MemStats       *runtime.MemStats
	metricsUseCase *metrics.MetricsUseCase
}

func NewAgentCollector(memStats *runtime.MemStats, metricsUseCase *metrics.MetricsUseCase) *AgentCollector {
	return &AgentCollector{
		MemStats:       memStats,
		metricsUseCase: metricsUseCase,
	}
}

func (a *AgentCollector) GetAllMetrics() ([]repository.Metric, error) {
	gaugeMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	if err != nil {
		return nil, err
	}
	counterMetrics, err := a.metricsUseCase.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	if err != nil {
		return nil, err
	}

	return append(gaugeMetrics, counterMetrics...), nil
}

func (a *AgentCollector) CollectMetrics() {
	memStats := a.MemStats
	runtime.ReadMemStats(a.MemStats)
	a.UpdateGaugeMetricValue("Alloc", float64(memStats.Alloc))
	a.UpdateGaugeMetricValue("Frees", float64(memStats.Frees))
	a.UpdateGaugeMetricValue("BuckHashSys", float64(memStats.BuckHashSys))
	a.UpdateGaugeMetricValue("GCCPUFraction", float64(memStats.GCCPUFraction))
	a.UpdateGaugeMetricValue("GCSys", float64(memStats.GCSys))
	a.UpdateGaugeMetricValue("HeapAlloc", float64(memStats.HeapAlloc))
	a.UpdateGaugeMetricValue("HeapIdle", float64(memStats.HeapIdle))
	a.UpdateGaugeMetricValue("HeapInuse", float64(memStats.HeapInuse))
	a.UpdateGaugeMetricValue("HeapObjects", float64(memStats.HeapObjects))
	a.UpdateGaugeMetricValue("HeapReleased", float64(memStats.HeapReleased))
	a.UpdateGaugeMetricValue("HeapSys", float64(memStats.HeapSys))
	a.UpdateGaugeMetricValue("LastGC", float64(memStats.LastGC))
	a.UpdateGaugeMetricValue("Lookups", float64(memStats.Lookups))
	a.UpdateGaugeMetricValue("MCacheSys", float64(memStats.MCacheSys))
	a.UpdateGaugeMetricValue("Mallocs", float64(memStats.Mallocs))
	a.UpdateGaugeMetricValue("NextGC", float64(memStats.NextGC))
	a.UpdateGaugeMetricValue("NumForcedGC", float64(memStats.NumForcedGC))
	a.UpdateGaugeMetricValue("NumGC", float64(memStats.NumGC))
	a.UpdateGaugeMetricValue("OtherSys", float64(memStats.OtherSys))
	a.UpdateGaugeMetricValue("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.UpdateGaugeMetricValue("StackInuse", float64(memStats.StackInuse))
	a.UpdateGaugeMetricValue("StackSys", float64(memStats.StackSys))
	a.UpdateGaugeMetricValue("Sys", float64(memStats.Sys))
	a.UpdateGaugeMetricValue("TotalAlloc", float64(memStats.TotalAlloc))
	a.UpdateGaugeMetricValue("MCacheInuse", float64(memStats.MCacheInuse))
	a.UpdateGaugeMetricValue("MSpanInuse", float64(memStats.MSpanInuse))
	a.UpdateGaugeMetricValue("MSpanSys", float64(memStats.MSpanSys))
	a.UpdateGaugeMetricValue("RandomValue", rand.Float64())
	a.UpdateCounterMetricValue("PollCount", int64(1))

	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error(err)
		return
	}

	a.UpdateGaugeMetricValue("TotalMemory", float64(v.Total))
	a.UpdateGaugeMetricValue("FreeMemory", float64(v.Free))
	a.UpdateGaugeMetricValue("CPUutilization1", v.UsedPercent)
}

func (a *AgentCollector) UpdateGaugeMetricValue(metricID string, value float64) {
	metric := repository.Metric{
		MType: repository.GaugeMetricKey,
		Value: &value,
		ID:    metricID,
	}
	a.metricsUseCase.UpdateMetric(context.TODO(), &metric)
}

func (a *AgentCollector) UpdateCounterMetricValue(metricID string, value int64) {
	metric := repository.Metric{
		MType: repository.CounterMetricKey,
		Delta: &value,
		ID:    metricID,
	}
	a.metricsUseCase.UpdateMetric(context.TODO(), &metric)
}
