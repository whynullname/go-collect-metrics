package repository

type Repository interface {
	TryGetGaugeMetricValue(metricName string) (float64, bool)
	TryGetCounterMetricValue(metricName string) (int64, bool)
	UpdateGaugeMetricValue(metricName string, metricValue float64)
	UpdateCounterMetricValue(metricName string, metricValue int64)
	GetAllGaugeMetrics() map[string]float64
	GetAllCounterMetrics() map[string]int64
}

const (
	GaugeMetricKey   = "gauge"
	CounterMetricKey = "counter"
)
