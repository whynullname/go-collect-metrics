package metrics

import (
	"strconv"

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

func (m *MetricsUseCase) TryUpdateMetricValue(metricType string, metricName string, value any) error {
	if metricType == repository.CounterMetricKey {
		metricValue, err := toInt64(value)
		if err != nil {
			return err
		}

		m.repository.UpdateCounterMetricValue(metricName, metricValue)
		return nil
	} else if metricType == repository.GaugeMetricKey {
		metricValue, err := toFloat64(value)
		if err != nil {
			return err
		}

		m.repository.UpdateGaugeMetricValue(metricName, metricValue)
		return nil
	}

	return types.ErrUnsupportedMetricType
}

func (m *MetricsUseCase) TryUpdateMetricValueFromJSON(json *repository.MetricsJSON) error {
	switch json.MType {
	case repository.CounterMetricKey:
		if json.Delta == nil {
			return types.ErrMetricNilValue
		}

		newValue := m.repository.UpdateCounterMetricValue(json.ID, *json.Delta)
		json.Delta = &newValue
		return nil
	case repository.GaugeMetricKey:
		if json.Value == nil {
			return types.ErrMetricNilValue
		}

		newValue := m.repository.UpdateGaugeMetricValue(json.ID, *json.Value)
		json.Value = &newValue
		return nil
	}

	return types.ErrUnsupportedMetricType
}

func (m *MetricsUseCase) TryGetMetricValue(metricType string, metricName string) (any, error) {
	switch metricType {
	case repository.CounterMetricKey:
		val, ok := m.repository.GetCounterMetricValue(metricName)
		if !ok {
			return nil, types.ErrCantFindMetric
		}

		return val, nil
	case repository.GaugeMetricKey:
		val, ok := m.repository.GetGaugeMetricValue(metricName)
		if !ok {
			return nil, types.ErrCantFindMetric
		}

		return val, nil
	default:
		return nil, types.ErrUnsupportedMetricType
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
		return nil, types.ErrUnsupportedMetricType
	}
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, types.ErrUnsupportedMetricValueType
	}
}

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, types.ErrUnsupportedMetricValueType
	}
}
