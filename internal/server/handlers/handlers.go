package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/repository/postgres"
	"github.com/whynullname/go-collect-metrics/internal/repository/types"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

const (
	allDataHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Storage Data</title>
</head>
<body>
    <h1>Gauge data</h1>
    <ul>
        {{range $key, $value := .Gauge}}
            <li><strong>{{$key}}:</strong> {{$value}}</li>
        {{end}}
    </ul>

    <h1>Counter data</h1>
    <ul>
        {{range $key, $value := .Counter}}
            <li><strong>{{$key}}:</strong> {{$value}}</li>
        {{end}}
    </ul>
</body>
</html>`
)

type Handlers struct {
	metricsUseCase *metrics.MetricsUseCase
	postgres       *postgres.Postgres //Сделано тестово для 10 инкремента, обязательно нужно сдеалть нормальным репозиторием
}

func NewHandlers(metricsUseCase *metrics.MetricsUseCase, postgres *postgres.Postgres) *Handlers {
	return &Handlers{
		metricsUseCase: metricsUseCase,
		postgres:       postgres,
	}
}

func (h *Handlers) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("webpage").Parse(allDataHTMLTemplate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gaugeMetrics, err := h.metricsUseCase.GetAllMetricsByType(repository.GaugeMetricKey)
	if err != nil {
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	counterMetrics, err := h.metricsUseCase.GetAllMetricsByType(repository.CounterMetricKey)
	if err != nil {
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	data := map[string]any{
		"Gauge":   gaugeMetrics,
		"Counter": counterMetrics,
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "text/plain" {
		logger.Log.Info("Content type not text/plain, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := chi.URLParam(r, "key")
	metricName := chi.URLParam(r, "merticName")
	metricValue := chi.URLParam(r, "metricValue")
	logger.Log.Infof("Data received! Key %s, metricaName %s, metricValue %s \n", metricType, metricName, metricValue)

	err := h.metricsUseCase.TryUpdateMetricValue(metricType, metricName, metricValue)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %s", err)

		if errors.Is(err, types.ErrUnsupportedMetricValueType) || errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) UpdateMetricForJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && contentType != "application/json" {
		logger.Log.Info("Content type not application/json, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metricJSON repository.MetricsJSON
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &metricJSON)
	if err != nil {
		logger.Log.Infof("Error while read from body %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.metricsUseCase.TryUpdateMetricValueFromJSON(&metricJSON)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)

		if errors.Is(err, types.ErrMetricNilValue) || errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(metricJSON)
	if err != nil {
		logger.Log.Errorf("Error with marshal output JSON: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (h *Handlers) GetMetricByName(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	val, err := h.metricsUseCase.TryGetMetricValue(metricType, metricName)
	if err != nil {
		if errors.Is(err, types.ErrCantFindMetric) {
			w.WriteHeader(http.StatusNotFound)
		}
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	switch value := val.(type) {
	case int64:
		io.WriteString(w, fmt.Sprintf("%d", value))
	case float64:
		io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (h *Handlers) GetMetricByNameFromJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		logger.Log.Info("Content type not application/json, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var buff bytes.Buffer
	var metricJSON repository.MetricsJSON

	if _, err := buff.ReadFrom(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Log.Infof("ERROR! %w", err)
		return
	}

	if err := json.Unmarshal(buff.Bytes(), &metricJSON); err != nil {
		logger.Log.Infof("ERROR! %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := h.metricsUseCase.TryGetMetricValue(metricJSON.MType, metricJSON.ID)
	if err != nil {
		if errors.Is(err, types.ErrCantFindMetric) {
			w.WriteHeader(http.StatusNotFound)
		}
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	switch metricJSON.MType {
	case repository.CounterMetricKey:
		intValue, ok := value.(int64)
		if !ok {
			return
		}

		metricJSON.Delta = &intValue
	case repository.GaugeMetricKey:
		floatValue, ok := value.(float64)
		if !ok {
			return
		}

		metricJSON.Value = &floatValue
	}

	resp, err := json.Marshal(metricJSON)
	if err != nil {
		logger.Log.Infof("ERROR! %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) PingPostgres(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := h.postgres.DB.PingContext(ctx); err != nil {
		logger.Log.Info(h.postgres.Adress)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
