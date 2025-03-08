package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func TestUpdateMetrics(t *testing.T) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	repo := inmemory.NewInMemoryRepository()
	cfg := config.NewAgentConfig()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	agInstance := NewAgent(&memStats, metricsUseCase, cfg)
	agInstance.UpdateMetrics()
	tests := []struct {
		name            string
		dataType        string
		dataName        string
		shouldDataExist bool
		dataValue       float64
	}{
		{
			name:            "Positive test data #1",
			dataType:        repository.GaugeMetricKey,
			dataName:        "Alloc",
			shouldDataExist: true,
			dataValue:       float64(memStats.Alloc),
		},
		{
			name:            "Positive test gauge data #2",
			dataType:        repository.GaugeMetricKey,
			dataName:        "NextGC",
			shouldDataExist: true,
			dataValue:       float64(memStats.NextGC),
		},
		{
			name:            "Positiove test counter data #1",
			dataType:        repository.CounterMetricKey,
			dataName:        "PollCount",
			shouldDataExist: true,
			dataValue:       1,
		},
		{
			name:            "Try get non-existent counter data",
			dataType:        repository.CounterMetricKey,
			dataName:        "TestDataName",
			shouldDataExist: false,
			dataValue:       0,
		},
		{
			name:            "Try get non-existent gauge data",
			dataType:        repository.GaugeMetricKey,
			dataName:        "TestDataName",
			shouldDataExist: false,
			dataValue:       0,
		},
		{
			name:            "Try get non-existent data type",
			dataType:        "testDataType",
			dataName:        "TestDataName",
			shouldDataExist: false,
			dataValue:       0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := agInstance.metricsUseCase.TryGetMetricValue(test.dataType, test.dataName)

			if test.shouldDataExist {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, test.dataValue, val)
		})
	}
}
