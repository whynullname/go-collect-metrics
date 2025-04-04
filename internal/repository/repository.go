package repository

type Repository interface {
	GetGaugeMetricValue(metricName string) (float64, bool)
	GetCounterMetricValue(metricName string) (int64, bool)
	UpdateGaugeMetricValue(metricName string, metricValue float64) float64
	UpdateCounterMetricValue(metricName string, metricValue int64) int64
	GetAllGaugeMetrics() map[string]float64
	GetAllCounterMetrics() map[string]int64
	CloseRepository()
}

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
