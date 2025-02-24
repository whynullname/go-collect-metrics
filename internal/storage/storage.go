package storage

import "fmt"

type memoryStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

const (
	GaugeKey   = "gauge"
	CounterKey = "counter"
)

var MemoryStorage *memoryStorage

func init() {
	MemoryStorage = newStorage()
}

func newStorage() *memoryStorage {
	return &memoryStorage{
		gauge:   make(map[string]float64, 0),
		counter: make(map[string]int64, 0),
	}
}

func (s *memoryStorage) UpdateGaugeData(key string, value float64) {
	fmt.Printf("Update gauge, key = %s value = %.2f\n", key, value)
	MemoryStorage.gauge[key] = value
}

func (s *memoryStorage) UpdateCounterData(key string, value int64) {
	fmt.Printf("Update counter, key = %s value = %d\n", key, value)
	MemoryStorage.counter[key] = value
}

func (s *memoryStorage) GetGaugeData(key string) (float64, bool) {
	val, ok := MemoryStorage.gauge[key]

	return val, ok
}

func (s *memoryStorage) GetCounterData(key string) (int64, bool) {
	val, ok := MemoryStorage.counter[key]

	return val, ok
}

func (s *memoryStorage) GetAllGaugeData() map[string]float64 {
	return MemoryStorage.gauge
}

func (s *memoryStorage) GetAllCounterData() map[string]int64 {
	return MemoryStorage.counter
}
