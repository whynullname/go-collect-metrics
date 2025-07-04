// Пакет metrics содержит usecase структуру, которая обеспечивает слой между хенделрами и репозиторием.
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

// UpdateMetric обновить метрику в репозитории.
func (m *MetricsUseCase) UpdateMetric(ctx context.Context, json *repository.Metric) (*repository.Metric, error) {
	if json == nil || (json.Delta == nil && json.Value == nil) {
		return nil, types.ErrMetricNilValue
	}

	if json.MType != repository.CounterMetricKey && json.MType != repository.GaugeMetricKey {
		return nil, types.ErrUnsupportedMetricType
	}

	return m.repository.UpdateMetric(ctx, json)
}

// UpdateMetrics обновить массив метрик в репозитории.
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

// GetMetric получить метрику по типу и имени.
func (m *MetricsUseCase) GetMetric(ctx context.Context, metricType string, metricName string) (*repository.Metric, error) {
	return m.repository.GetMetric(ctx, metricName, metricType)
}

// GetAllMetricsByType получить слайс метрик по типу.
func (m *MetricsUseCase) GetAllMetricsByType(ctx context.Context, metricType string) ([]repository.Metric, error) {
	return m.repository.GetAllMetricsByType(ctx, metricType)
}
