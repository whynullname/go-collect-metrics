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

func (s *memoryStorage) UpdateGauge(key string, value float64) {
	fmt.Printf("Update gauge, key = %s value = %.2f\n", key, value)
	MemoryStorage.gauge[key] = value
}

func (s *memoryStorage) UpdateCounter(key string, value int64) {
	fmt.Printf("Update counter, key = %s value = %d\n", key, value)
	MemoryStorage.counter[key] = value
}
