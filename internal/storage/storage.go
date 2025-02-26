package storage

type MemoryStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

const (
	GaugeKey   = "gauge"
	CounterKey = "counter"
)

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		gauge:   make(map[string]float64, 0),
		counter: make(map[string]int64, 0),
	}
}

func (s *MemoryStorage) UpdateData(dataType string, key string, value float64) {
	switch dataType {
	case GaugeKey:
		s.UpdateGaugeData(key, value)
	case CounterKey:
		s.UpdateCounterData(key, int64(value))
	}
}

func (s *MemoryStorage) GetData(dataType string, key string) (float64, bool) {
	switch dataType {
	case GaugeKey:
		return s.GetGaugeData(key)
	case CounterKey:
		value, ok := s.GetCounterData(key)
		return float64(value), ok
	}

	return 0, false
}

func (s *MemoryStorage) UpdateGaugeData(key string, value float64) {
	s.gauge[key] = value
}

func (s *MemoryStorage) UpdateCounterData(key string, value int64) {
	s.counter[key] = value
}

func (s *MemoryStorage) GetGaugeData(key string) (float64, bool) {
	val, ok := s.gauge[key]

	return val, ok
}

func (s *MemoryStorage) GetCounterData(key string) (int64, bool) {
	val, ok := s.counter[key]

	return val, ok
}

func (s *MemoryStorage) GetAllGaugeData() map[string]float64 {
	return s.gauge
}

func (s *MemoryStorage) GetAllCounterData() map[string]int64 {
	return s.counter
}
