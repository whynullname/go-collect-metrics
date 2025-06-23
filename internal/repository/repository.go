// Пакет repository описывает основные методы и структуры, которые нужны для репозитория.
package repository

import "context"

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)

type Repository interface {
	UpdateMetric(ctx context.Context, metric *Metric) (*Metric, error)                    // обнолвение метрики.
	UpdateMetrics(ctx context.Context, metrics []Metric) ([]Metric, error)                // обновление массива метрик.
	GetMetric(ctx context.Context, metricName string, metricType string) (*Metric, error) // получить метрику.
	GetAllMetricsByType(ctx context.Context, metricType string) ([]Metric, error)         // получить все метрики по типу.
	PingRepo() bool                                                                       // узнать, доступен ли репозиторий и можно ли к нему обращаться.
	CloseRepository()                                                                     // закрыть репозиторий.
}

// Metric хранит всю информацию о метрике.
type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metric) GetValue() float64 {
	if m == nil || m.Value == nil {
		return 0
	}
	return *m.Value
}

func (m *Metric) GetDelta() int64 {
	if m == nil || m.Delta == nil {
		return 0
	}
	return *m.Delta
}
