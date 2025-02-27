package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func TestUpdateMetrics(t *testing.T) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	storage := storage.NewStorage()
	agInstance := NewAgent(&memStats, storage, "localhost:8080")
	agInstance.UpdateMetrics()
	tests := []struct {
		name      string
		dataName  string
		dataValue float64
	}{
		{
			name:      "positive test gauge data #1",
			dataName:  "Alloc",
			dataValue: float64(memStats.Alloc),
		},
		{
			name:      "positive test gauge data #2",
			dataName:  "NextGC",
			dataValue: float64(memStats.NextGC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, ok := agInstance.storage.GetGaugeData(test.dataName)
			assert.True(t, ok)
			assert.Equal(t, val, test.dataValue)
		})
	}
}
