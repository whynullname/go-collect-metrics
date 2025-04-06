package metrics

import (
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/types"
)

type MetricsUseCase struct {
	repository repository.Repository
}

func NewMetricUseCase(repository repository.Repository) *MetricsUseCase {
	return &MetricsUseCase{
		repository: repository,
	}
}

func (m *MetricsUseCase) UpdateMetric(json *repository.Metric) (*repository.Metric, error) {
	if json == nil || (json.Delta == nil && json.Value == nil) {
		return nil, types.ErrMetricNilValue
	}

	if json.MType != repository.CounterMetricKey && json.MType != repository.GaugeMetricKey {
		return nil, types.ErrUnsupportedMetricType
	}

	metric := m.repository.UpdateMetric(json)
	if metric == nil {
		return nil, types.ErrWileUpdateMetric
	}
	return metric, nil
}

func (m *MetricsUseCase) GetMetric(metricType string, metricName string) (*repository.Metric, error) {
	if metricType != repository.CounterMetricKey && metricType != repository.GaugeMetricKey {
		return nil, types.ErrUnsupportedMetricType
	}

	metric, ok := m.repository.GetMetric(metricName, metricType)
	if !ok {
		return nil, types.ErrCantFindMetric
	}
	return metric, nil
}

func (m *MetricsUseCase) GetAllMetricsByType(metricType string) []repository.Metric {
	return m.repository.GetAllMetricsByType(metricType)
}
