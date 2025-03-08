package metrics

import (
	"errors"
	"fmt"

	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type MetricsUseCase struct {
	repository repository.Repository
}

func NewMetricUseCase(repository repository.Repository) *MetricsUseCase {
	return &MetricsUseCase{
		repository: repository,
	}
}

func (m *MetricsUseCase) TryUpdateMetricValue(metricType string, metricName string, value any) error {
	switch metricType {
	case repository.CounterMetricKey:
		intValue, err := value.(int64)

		if !err {
			return fmt.Errorf("metric type %s can be only int64", metricType)
		}

		m.repository.UpdateCounterMetricValue(metricName, intValue)
	case repository.GaugeMetricKey:
		floatValue, err := value.(float64)

		if !err {
			return fmt.Errorf("metric type %s can be only float64", metricType)
		}

		m.repository.UpdateGaugeMetricValue(metricName, floatValue)
	default:
		return errors.New("unsupported metric type")
	}

	return nil
}

func (m *MetricsUseCase) TryGetMetricValue(metricType string, metricName string) (any, error) {
	switch metricType {
	case repository.CounterMetricKey:
		val, ok := m.repository.TryGetCounterMetricValue(metricName)

		if !ok {
			return nil, fmt.Errorf("can't find metric with name %s", metricName)
		}

		return val, nil
	case repository.GaugeMetricKey:
		val, ok := m.repository.TryGetGaugeMetricValue(metricName)

		if !ok {
			return nil, fmt.Errorf("can't find metric with name %s", metricName)
		}

		return val, nil
	default:
		return nil, errors.New("unsupported metric type")
	}
}

func (m *MetricsUseCase) GetAllMetricsByType(metricType string) (map[string]any, error) {
	switch metricType {
	case repository.CounterMetricKey:
		metrics := m.repository.GetAllCounterMetrics()
		result := make(map[string]any)
		for k, v := range metrics {
			result[k] = v
		}
		return result, nil
	case repository.GaugeMetricKey:
		metrics := m.repository.GetAllGaugeMetrics()
		result := make(map[string]any)
		for k, v := range metrics {
			result[k] = v
		}
		return result, nil
	default:
		return nil, errors.New("unsupported metric type")
	}
}
