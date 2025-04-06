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
	var ouputMetric *repository.Metric
	for _, savedMetric := range i.metrics {
		if savedMetric.ID == metric.ID {
			ouputMetric = &savedMetric
			break
		}
	}

	if ouputMetric == nil {
		i.metrics = append(i.metrics, *metric)
		return metric
	}

	switch metric.MType {
	case repository.GaugeMetricKey:
		ouputMetric.Value = metric.Value
		break
	case repository.CounterMetricKey:
		sum := (*ouputMetric.Delta) + (*metric.Delta)
		ouputMetric.Delta = &sum
		break
	}

	return ouputMetric
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
