package inmemory

import (
	"sync"

	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type InMemoryRepo struct {
	mx      sync.RWMutex
	metrics []repository.Metric
}

func NewInMemoryRepository() *InMemoryRepo {
	return &InMemoryRepo{
		metrics: make([]repository.Metric, 0),
	}
}

func (i *InMemoryRepo) UpdateMetric(metric *repository.Metric) *repository.Metric {
	i.mx.RLock()
	defer i.mx.RUnlock()
	for j, savedMetric := range i.metrics {
		if savedMetric.ID == metric.ID {
			switch metric.MType {
			case repository.GaugeMetricKey:
				savedMetric.Value = metric.Value
				i.metrics[j] = savedMetric
				break
			case repository.CounterMetricKey:
				sum := (*savedMetric.Delta) + (*metric.Delta)
				savedMetric.Delta = &sum
				i.metrics[j] = savedMetric
				break
			}
			return &savedMetric
		}
	}

	i.metrics = append(i.metrics, *metric)
	return metric
}

func (i *InMemoryRepo) UpdateMetrics(metrics []repository.Metric) ([]repository.Metric, error) {
	output := make([]repository.Metric, 0)
	for _, metric := range metrics {
		output = append(output, *i.UpdateMetric(&metric))
	}
	return output, nil
}

func (i *InMemoryRepo) GetMetric(metricName string, metricType string) (*repository.Metric, bool) {
	i.mx.Lock()
	defer i.mx.Unlock()

	for _, savedMetric := range i.metrics {
		if savedMetric.ID == metricName &&
			savedMetric.MType == metricType {
			return &savedMetric, true
		}
	}

	return nil, false
}

func (i *InMemoryRepo) GetAllMetricsByType(metricType string) []repository.Metric {
	i.mx.Lock()
	defer i.mx.Unlock()
	output := make([]repository.Metric, 0)

	for _, savedMetric := range i.metrics {
		if savedMetric.MType == metricType {
			output = append(output, savedMetric)
		}
	}

	return output
}

func (i *InMemoryRepo) CloseRepository() {

}

func (i *InMemoryRepo) PingRepo() bool {
	return true
}
