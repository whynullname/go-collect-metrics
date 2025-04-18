package inmemory

import (
	"sync"

	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type InMemoryRepo struct {
	mx             sync.RWMutex
	counterMetrics map[string]int64
	gaugeMetrics   map[string]float64
}

func NewInMemoryRepository() *InMemoryRepo {
	return &InMemoryRepo{
		counterMetrics: make(map[string]int64, 0),
		gaugeMetrics:   make(map[string]float64, 0),
	}
}

func (i *InMemoryRepo) UpdateMetric(metric *repository.Metric) *repository.Metric {
	i.mx.Lock()
	defer i.mx.Unlock()

	switch metric.MType {
	case repository.GaugeMetricKey:
		i.gaugeMetrics[metric.ID] = *metric.Value
	case repository.CounterMetricKey:
		metricValue, ok := i.counterMetrics[metric.ID]
		if !ok {
			i.counterMetrics[metric.ID] = *metric.Delta
		} else {
			sum := metricValue + (*metric.Delta)
			i.counterMetrics[metric.ID] = sum
			metric.Delta = &sum
		}
	}

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
	i.mx.RLock()
	defer i.mx.RUnlock()

	outputMetric := repository.Metric{
		MType: metricType,
		ID:    metricName,
	}
	isContains := false

	switch metricType {
	case repository.GaugeMetricKey:
		metricValue, ok := i.gaugeMetrics[metricName]
		if ok {
			outputMetric.Value = &metricValue
		}
		isContains = ok
	case repository.CounterMetricKey:
		metricValue, ok := i.counterMetrics[metricName]
		if ok {
			outputMetric.Delta = &metricValue
		}
		isContains = ok
	}

	return &outputMetric, isContains
}

func (i *InMemoryRepo) GetAllMetricsByType(metricType string) []repository.Metric {
	i.mx.RLock()
	defer i.mx.RUnlock()
	output := make([]repository.Metric, 0)

	switch metricType {
	case repository.GaugeMetricKey:
		for name, value := range i.gaugeMetrics {
			output = append(output, repository.Metric{
				ID:    name,
				MType: repository.GaugeMetricKey,
				Value: &value,
			})
		}
	case repository.CounterMetricKey:
		for name, delta := range i.counterMetrics {
			output = append(output, repository.Metric{
				ID:    name,
				MType: repository.CounterMetricKey,
				Delta: &delta,
			})
		}
	}

	return output
}

func (i *InMemoryRepo) CloseRepository() {

}

func (i *InMemoryRepo) PingRepo() bool {
	return false
}
