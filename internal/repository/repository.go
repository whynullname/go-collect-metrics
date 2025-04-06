package repository

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)

type Repository interface {
	UpdateMetric(metric *Metric) *Metric
	UpdateMetrics(metrics []Metric) ([]Metric, error)
	GetMetric(metricName string, metricType string) (*Metric, bool)
	GetAllMetricsByType(metricType string) []Metric
	PingRepo() bool
	CloseRepository()
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
