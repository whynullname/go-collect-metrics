package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	scanner *bufio.Scanner
	mx      sync.RWMutex
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return &FileStorage{
		file:    file,
		encoder: encoder,
		decoder: json.NewDecoder(file),
		scanner: bufio.NewScanner(file),
	}, nil
}

func (s *FileStorage) RecordMetric(interval uint64, repo repository.Repository) {
	defer s.file.Close()
	duration := time.Duration(interval) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for range ticker.C {
		s.WriteMetrics(repo)
	}
}

func (s *FileStorage) WriteMetrics(repo repository.Repository) error {
	s.mx.RLock()
	defer s.mx.RUnlock()
	gaugeMetrics := repo.GetAllMetricsByType(context.TODO(), repository.GaugeMetricKey)
	counterMetrics := repo.GetAllMetricsByType(context.TODO(), repository.CounterMetricKey)
	outputMetrics := append(gaugeMetrics, counterMetrics...)
	s.file.Seek(0, 0)
	s.file.Truncate(0)
	defer s.file.Sync()
	return s.encoder.Encode(&outputMetrics)
}

func (s *FileStorage) ReadAllMetrics(repo repository.Repository) error {
	s.mx.RLock()
	defer s.mx.RUnlock()
	savedMetrics := make([]repository.Metric, 0)
	err := s.decoder.Decode(&savedMetrics)
	if err != nil {
		return err
	}

	for _, metric := range savedMetrics {
		repo.UpdateMetric(context.TODO(), &metric)
	}

	return nil
}
