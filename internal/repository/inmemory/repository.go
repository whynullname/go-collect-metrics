package inmemory

import "sync"

type InMemoryRepo struct {
	mx             sync.RWMutex
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func NewInMemoryRepository() *InMemoryRepo {
	return &InMemoryRepo{
		mx:             sync.RWMutex{},
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
	}
}

func (i *InMemoryRepo) TryGetGaugeMetricValue(metricName string) (float64, bool) {
	i.mx.RLock()
	defer i.mx.RUnlock()
	val, ok := i.GaugeMetrics[metricName]
	return val, ok
}

func (i *InMemoryRepo) TryGetCounterMetricValue(metricName string) (int64, bool) {
	i.mx.RLock()
	defer i.mx.RUnlock()
	val, ok := i.CounterMetrics[metricName]
	return val, ok
}

func (i *InMemoryRepo) UpdateGaugeMetricValue(metricName string, metricValue float64) {
	i.mx.Lock()
	defer i.mx.Unlock()
	i.GaugeMetrics[metricName] = metricValue
}

func (i *InMemoryRepo) UpdateCounterMetricValue(metricName string, metricValue int64) {
	i.mx.Lock()
	defer i.mx.Unlock()

	val, ok := i.CounterMetrics[metricName]

	if !ok {
		val = metricValue
	} else {
		val += metricValue
	}

	i.CounterMetrics[metricName] = val
}

func (i *InMemoryRepo) GetAllGaugeMetrics() map[string]float64 {
	i.mx.Lock()
	defer i.mx.Unlock()
	return i.GaugeMetrics
}

func (i *InMemoryRepo) GetAllCounterMetrics() map[string]int64 {
	i.mx.Lock()
	defer i.mx.Unlock()
	return i.CounterMetrics
}
