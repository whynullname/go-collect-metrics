package storage

type MemoryStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

const (
	GaugeKey   = "gauge"
	CounterKey = "counter"
)

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		Gauge:   make(map[string]float64, 0),
		Counter: make(map[string]int64, 0),
	}
}

func (s *MemoryStorage) UpdateMetrics(dataType string, key string, value float64) {
	switch dataType {
	case GaugeKey:
		s.UpdateGaugeMetric(key, value)
	case CounterKey:
		s.UpdateCounterMetric(key)
	}
}

func (s *MemoryStorage) GetMetrics(dataType string, key string) (float64, bool) {
	switch dataType {
	case GaugeKey:
		return s.GetGaugeMetric(key)
	case CounterKey:
		value, ok := s.GetCounterMetric(key)
		return float64(value), ok
	}

	return 0, false
}

func (s *MemoryStorage) UpdateGaugeMetric(key string, value float64) {
	s.Gauge[key] = value
}

func (s *MemoryStorage) UpdateCounterMetric(key string) {
	val, ok := s.GetCounterMetric(key)

	if !ok {
		val = 1
	} else {
		val++
	}

	s.Counter[key] = val
}

func (s *MemoryStorage) GetGaugeMetric(key string) (float64, bool) {
	val, ok := s.Gauge[key]
	return val, ok
}

func (s *MemoryStorage) GetCounterMetric(key string) (int64, bool) {
	val, ok := s.Counter[key]
	return val, ok
}

func (s *MemoryStorage) GetAllGaugeMetrics() map[string]float64 {
	return s.Gauge
}

func (s *MemoryStorage) GetAllCounterMetrics() map[string]int64 {
	return s.Counter
}
