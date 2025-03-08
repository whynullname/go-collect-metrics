package inmemory

type InMemoryRepo struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewInMemoryRepository() *InMemoryRepo {
	return &InMemoryRepo{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (i *InMemoryRepo) TryGetGaugeMetricValue(metricName string) (float64, bool) {
	val, ok := i.GaugeMetrics[metricName]
	return val, ok
}

func (i *InMemoryRepo) TryGetCounterMetricValue(metricName string) (int64, bool) {
	val, ok := i.CounterMetrics[metricName]
	return val, ok
}

func (i *InMemoryRepo) UpdateGaugeMetricValue(metricName string, metricValue float64) {
	i.GaugeMetrics[metricName] = metricValue
}

func (i *InMemoryRepo) UpdateCounterMetricValue(metricName string, metricValue int64) {
	val, ok := i.TryGetCounterMetricValue(metricName)

	if !ok {
		val = metricValue
	} else {
		val += metricValue
	}

	i.CounterMetrics[metricName] = val
}

func (i *InMemoryRepo) GetAllGaugeMetrics() map[string]float64 {
	return i.GaugeMetrics
}

func (i *InMemoryRepo) GetAllCounterMetrics() map[string]int64 {
	return i.CounterMetrics
}
