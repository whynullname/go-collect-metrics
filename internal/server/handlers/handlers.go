package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/repository"
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
	pingRepoFunc   func() bool
}

func NewHandlers(metricsUseCase *metrics.MetricsUseCase, pingRepoFunc func() bool) *Handlers {
	return &Handlers{
		metricsUseCase: metricsUseCase,
		pingRepoFunc:   pingRepoFunc,
	}
}

// TODO: ПЕРЕДЕЛАТЬ ПОД repository.Metric
func (h *Handlers) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("webpage").Parse(allDataHTMLTemplate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gaugeMetrics := h.metricsUseCase.GetAllMetricsByType(r.Context(), repository.GaugeMetricKey)
	counterMetrics := h.metricsUseCase.GetAllMetricsByType(r.Context(), repository.CounterMetricKey)
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

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "merticName")
	metricValue := chi.URLParam(r, "metricValue")

	var metricObject repository.Metric
	switch metricType {
	case repository.CounterMetricKey:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			logger.Log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metricObject = repository.Metric{
			MType: repository.CounterMetricKey,
			Delta: &value,
			ID:    metricName,
		}
	case repository.GaugeMetricKey:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			logger.Log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metricObject = repository.Metric{
			MType: repository.GaugeMetricKey,
			Value: &value,
			ID:    metricName,
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Log.Infof("Data received! Key %s, metricaName %s, metricValue %s \n", metricType, metricName, metricValue)
	h.metricsUseCase.UpdateMetric(r.Context(), &metricObject)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) UpdateMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && contentType != "application/json" {
		logger.Log.Info("Content type not application/json, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metricJSON repository.Metric
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Log.Error("Error while read from body %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metricJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedMetric, err := h.metricsUseCase.UpdateMetric(r.Context(), &metricJSON)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)
		if errors.Is(err, types.ErrMetricNilValue) || errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(updatedMetric)
	if err != nil {
		logger.Log.Errorf("Error with marshal output JSON: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (h *Handlers) UpdateArrayJSONMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && contentType != "application/json" {
		logger.Log.Info("Content type not application/json, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metrics []repository.Metric
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	outputMetrics, err := h.metricsUseCase.UpdateMetrics(r.Context(), metrics)
	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)

		if errors.Is(err, types.ErrMetricNilValue) || errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(outputMetrics)
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

	val, err := h.metricsUseCase.GetMetric(r.Context(), metricType, metricName)
	if err != nil {
		if errors.Is(err, types.ErrCantFindMetric) {
			w.WriteHeader(http.StatusNotFound)
		}
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	switch metricType {
	case repository.CounterMetricKey:
		io.WriteString(w, fmt.Sprintf("%d", *val.Delta))
	case repository.GaugeMetricKey:
		io.WriteString(w, strconv.FormatFloat(*val.Value, 'f', -1, 64))
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
	var metricJSON repository.Metric

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

	ouputMetric, err := h.metricsUseCase.GetMetric(r.Context(), metricJSON.MType, metricJSON.ID)
	if err != nil {
		if errors.Is(err, types.ErrCantFindMetric) {
			w.WriteHeader(http.StatusNotFound)
		}
		if errors.Is(err, types.ErrUnsupportedMetricType) {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	resp, err := json.Marshal(ouputMetric)
	if err != nil {
		logger.Log.Infof("ERROR! %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) PingRepository(w http.ResponseWriter, r *http.Request) {
	if h.pingRepoFunc() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
