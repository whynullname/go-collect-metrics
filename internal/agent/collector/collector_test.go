package collector

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func TestCollectAndGetMetrics(t *testing.T) {
	repo := inmemory.NewInMemoryRepository()
	useCase := metrics.NewMetricUseCase(repo)
	memStats := &runtime.MemStats{}
	collector := NewAgentCollector(memStats, useCase)
	collector.CollectMetrics()
	savedMetrics, err := collector.GetAllMetrics()
	if err != nil {
		t.Fatal(err)
		return
	}

	alloc := float64(memStats.Alloc)
	frees := float64(memStats.Frees)
	poolCount := int64(1)
	tests := []struct {
		testName        string
		shouldDataExist bool
		metricValue     repository.Metric
	}{
		{
			testName:        "Get Alloc metric",
			shouldDataExist: true,
			metricValue: repository.Metric{
				ID:    "Alloc",
				MType: repository.GaugeMetricKey,
				Value: &alloc,
			},
		},
		{
			testName:        "Get Frees metrics",
			shouldDataExist: true,
			metricValue: repository.Metric{
				ID:    "Frees",
				MType: repository.GaugeMetricKey,
				Value: &frees,
			},
		},
		{
			testName:        "Get non-existent metric",
			shouldDataExist: false,
			metricValue: repository.Metric{
				ID:    "test",
				MType: repository.CounterMetricKey,
			},
		},
		{
			testName:        "Get Poll Count metric",
			shouldDataExist: true,
			metricValue: repository.Metric{
				ID:    "PollCount",
				MType: repository.CounterMetricKey,
				Delta: &poolCount,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var savedMetric *repository.Metric
			for _, metric := range savedMetrics {
				if test.metricValue.ID == metric.ID {
					savedMetric = &metric
				}
			}

			if test.shouldDataExist {
				assert.NotNil(t, savedMetric)
				assert.Equal(t, savedMetric.Value, test.metricValue.Value)
				assert.Equal(t, savedMetric.Delta, test.metricValue.Delta)
			} else {
				assert.Nil(t, savedMetric)
			}
		})
	}
}
