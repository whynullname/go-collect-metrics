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
	s.Gauge[key] = value
}

func (s *MemoryStorage) UpdateCounterData(key string, value int64) {
	s.Counter[key] = value
}

func (s *MemoryStorage) GetGaugeData(key string) (float64, bool) {
	val, ok := s.Gauge[key]

	return val, ok
}

func (s *MemoryStorage) GetCounterData(key string) (int64, bool) {
	val, ok := s.Counter[key]

	return val, ok
}

func (s *MemoryStorage) GetAllGaugeData() map[string]float64 {
	return s.Gauge
}

func (s *MemoryStorage) GetAllCounterData() map[string]int64 {
	return s.Counter
}
