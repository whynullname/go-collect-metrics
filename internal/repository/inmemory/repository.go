package inmemory

import (
	"context"
	"sync"

	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/types"
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

func (i *InMemoryRepo) UpdateMetric(ctx context.Context, metric *repository.Metric) (*repository.Metric, error) {
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

	return metric, nil
}

func (i *InMemoryRepo) UpdateMetrics(ctx context.Context, metrics []repository.Metric) ([]repository.Metric, error) {
	output := make([]repository.Metric, 0)
	for _, metric := range metrics {
		updatedMetric, err := i.UpdateMetric(ctx, &metric)
		if err != nil {
			return nil, err
		}
		output = append(output, *updatedMetric)
	}
	return output, nil
}

func (i *InMemoryRepo) GetMetric(ctx context.Context, metricName string, metricType string) (*repository.Metric, error) {
	i.mx.RLock()
	defer i.mx.RUnlock()

	outputMetric := repository.Metric{
		MType: metricType,
		ID:    metricName,
	}
	var err error

	switch metricType {
	case repository.GaugeMetricKey:
		metricValue, ok := i.gaugeMetrics[metricName]
		if ok {
			outputMetric.Value = &metricValue
		} else {
			err = types.ErrCantFindMetric
		}
	case repository.CounterMetricKey:
		metricValue, ok := i.counterMetrics[metricName]
		if ok {
			outputMetric.Delta = &metricValue
		} else {
			err = types.ErrCantFindMetric
		}
	}

	return &outputMetric, err
}

func (i *InMemoryRepo) GetAllMetricsByType(ctx context.Context, metricType string) ([]repository.Metric, error) {
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

	return output, nil
}

func (i *InMemoryRepo) CloseRepository() {

}

func (i *InMemoryRepo) PingRepo() bool {
	return false
}
