package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func TestUpdateGaugeMetrics(t *testing.T) {
	logger.Initialize("info")
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	repo := inmemory.NewInMemoryRepository()
	cfg := config.NewAgentConfig()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	agInstance := NewAgent(&memStats, metricsUseCase, cfg)
	agInstance.UpdateMetrics()

	alloc := float64(memStats.Alloc)
	nexGC := float64(memStats.NextGC)
	tests := []struct {
		name            string
		shouldDataExist bool
		data            repository.Metric
	}{
		{
			name:            "Positive test data #1",
			shouldDataExist: true,
			data: repository.Metric{
				MType: repository.GaugeMetricKey,
				ID:    "Alloc",
				Value: &alloc,
			},
		},
		{
			name:            "Positive test gauge data #2",
			shouldDataExist: true,
			data: repository.Metric{
				MType: repository.GaugeMetricKey,
				ID:    "NextGC",
				Value: &nexGC,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := agInstance.metricsUseCase.GetMetric(test.data.MType, test.data.ID)

			if test.shouldDataExist {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, test.data.Value, val.Value)
		})
	}
}
