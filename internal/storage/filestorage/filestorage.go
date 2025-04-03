package filestorage

import (
	"bufio"
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
	gaugeMetrics := repo.GetAllGaugeMetrics()
	counterMetrics := repo.GetAllCounterMetrics()

	outputMetrics := make([]repository.MetricsJSON, 0)
	for metricName, metricValue := range gaugeMetrics {
		outputMetrics = append(outputMetrics, repository.MetricsJSON{
			ID:    metricName,
			MType: repository.GaugeMetricKey,
			Value: &metricValue,
		})
	}

	for metricName, metricValue := range counterMetrics {
		outputMetrics = append(outputMetrics, repository.MetricsJSON{
			ID:    metricName,
			MType: repository.CounterMetricKey,
			Delta: &metricValue,
		})
	}

	s.file.Seek(0, 0)
	s.file.Truncate(0)
	defer s.file.Sync()
	return s.encoder.Encode(&outputMetrics)
}

func (s *FileStorage) ReadAllMetrics(repo repository.Repository) error {
	s.mx.RLock()
	defer s.mx.RUnlock()
	savedMetrics := make([]repository.MetricsJSON, 0)
	err := s.decoder.Decode(&savedMetrics)
	if err != nil {
		return err
	}

	for _, metric := range savedMetrics {
		if metric.MType == repository.CounterMetricKey {
			repo.UpdateCounterMetricValue(metric.ID, *metric.Delta)
		}

		if metric.MType == repository.GaugeMetricKey {
			repo.UpdateGaugeMetricValue(metric.ID, *metric.Value)
		}
	}

	return nil
}
