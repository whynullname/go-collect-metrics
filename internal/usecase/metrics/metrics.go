package metrics

import (
	"context"

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

func (m *MetricsUseCase) UpdateMetric(ctx context.Context, json *repository.Metric) (*repository.Metric, error) {
	if json == nil || (json.Delta == nil && json.Value == nil) {
		return nil, types.ErrMetricNilValue
	}

	if json.MType != repository.CounterMetricKey && json.MType != repository.GaugeMetricKey {
		return nil, types.ErrUnsupportedMetricType
	}

	metric := m.repository.UpdateMetric(ctx, json)
	if metric == nil {
		return nil, types.ErrWileUpdateMetric
	}
	return metric, nil
}

func (m *MetricsUseCase) UpdateMetrics(ctx context.Context, metrics []repository.Metric) ([]repository.Metric, error) {
	for _, metric := range metrics {
		if metric.Delta == nil && metric.Value == nil {
			return nil, types.ErrMetricNilValue
		}

		if metric.MType != repository.CounterMetricKey && metric.MType != repository.GaugeMetricKey {
			return nil, types.ErrUnsupportedMetricType
		}
	}

	return m.repository.UpdateMetrics(ctx, metrics)
}

func (m *MetricsUseCase) GetMetric(ctx context.Context, metricType string, metricName string) (*repository.Metric, error) {
	if metricType != repository.CounterMetricKey && metricType != repository.GaugeMetricKey {
		return nil, types.ErrUnsupportedMetricType
	}

	metric, ok := m.repository.GetMetric(ctx, metricName, metricType)
	if !ok {
		return nil, types.ErrCantFindMetric
	}
	return metric, nil
}

func (m *MetricsUseCase) GetAllMetricsByType(ctx context.Context, metricType string) []repository.Metric {
	return m.repository.GetAllMetricsByType(ctx, metricType)
}
