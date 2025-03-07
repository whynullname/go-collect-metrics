package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	config "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func TestUpdateMetrics(t *testing.T) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	data := storage.NewStorage()
	cfg := config.NewAgentConfig()
	agInstance := NewAgent(&memStats, data, cfg)
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
			dataType:        storage.GaugeKey,
			dataName:        "Alloc",
			shouldDataExist: true,
			dataValue:       float64(memStats.Alloc),
		},
		{
			name:            "Positive test gauge data #2",
			dataType:        storage.GaugeKey,
			dataName:        "NextGC",
			shouldDataExist: true,
			dataValue:       float64(memStats.NextGC),
		},
		{
			name:            "Positiove test counter data #1",
			dataType:        storage.CounterKey,
			dataName:        "PollCount",
			shouldDataExist: true,
			dataValue:       1,
		},
		{
			name:            "Try get non-existent counter data",
			dataType:        storage.CounterKey,
			dataName:        "TestDataName",
			shouldDataExist: false,
			dataValue:       0,
		},
		{
			name:            "Try get non-existent gauge data",
			dataType:        storage.GaugeKey,
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
			val, ok := agInstance.storage.GetMetricValue(test.dataType, test.dataName)
			assert.Equal(t, test.shouldDataExist, ok)

			if !ok {
				return
			}

			assert.Equal(t, test.dataValue, val)
		})
	}
}
