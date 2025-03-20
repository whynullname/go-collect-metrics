package repository

type Repository interface {
	TryGetGaugeMetricValue(metricName string) (float64, bool)
	TryGetCounterMetricValue(metricName string) (int64, bool)
	UpdateGaugeMetricValue(metricName string, metricValue float64)
	UpdateCounterMetricValue(metricName string, metricValue int64)
	GetAllGaugeMetrics() map[string]float64
	GetAllCounterMetrics() map[string]int64
}

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)

type MetricsJson struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
