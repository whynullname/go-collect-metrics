package repository

import "context"

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric *Metric) (*Metric, error)
	UpdateMetrics(ctx context.Context, metrics []Metric) ([]Metric, error)
	GetMetric(ctx context.Context, metricName string, metricType string) (*Metric, error)
	GetAllMetricsByType(ctx context.Context, metricType string) ([]Metric, error)
	PingRepo() bool
	CloseRepository()
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
