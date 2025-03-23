package metrics

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/whynullname/go-collect-metrics/internal/logger"
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

	return errors.New("unsupported metric type")
}

func (m *MetricsUseCase) TryUpdateMetricValueFromJSON(json *repository.MetricsJSON) error {
	if json.MType == repository.CounterMetricKey {
		if json.Delta == nil {
			return errors.New("delta for update conter metric is nil")
		}
		logger.Log.Infof("Add new counter metric %s", json.ID)
		m.repository.UpdateCounterMetricValue(json.ID, *json.Delta)
		newValue, _ := m.repository.TryGetCounterMetricValue(json.ID)
		json.Delta = &newValue
		return nil
	} else if json.MType == repository.GaugeMetricKey {
		if json.Value == nil {
			return errors.New("value for update gauge metric is nil")
		}

		logger.Log.Infof("Add new gauge metric %s", json.ID)
		m.repository.UpdateGaugeMetricValue(json.ID, *json.Value)
		return nil
	}

	return errors.New("unsupported metric type")
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

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

func toFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
