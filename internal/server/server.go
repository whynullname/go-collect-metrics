package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
	"github.com/whynullname/go-collect-metrics/internal/middlewares"
	"github.com/whynullname/go-collect-metrics/internal/middlewares/compressmiddleware"
	"github.com/whynullname/go-collect-metrics/internal/repository"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Server struct {
	metricsUseCase *metrics.MetricsUseCase
	Config         *config.ServerConfig
	Router         chi.Router
}

const (
	updateHandleFuncName = "/update/"
	allDataHTMLTemplate  = `
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

func NewServer(metricsUseCase *metrics.MetricsUseCase, config *config.ServerConfig) *Server {
	err := logger.Initialize("info")

	if err != nil {
		log.Fatalf("Fatal initialize logger")
	}

	serverInstance := &Server{
		metricsUseCase: metricsUseCase,
		Config:         config,
	}

	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.Logging)
	r.Use(compressmiddleware.GZIP)

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.GetAllMetrics)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", s.GetMetricByNameFromJSON)
			r.Get("/{metricType}/{metricName}", s.GetMetricByName)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.UpdateMetricForJSON)
			r.Post("/{key}/{merticName}/{metricValue}", s.UpdateMetric)
		})
	})
	return r
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Config.EndPointAdress, s.Router)
}

func (s *Server) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("webpage").Parse(allDataHTMLTemplate)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gaugeMetrics, err := s.metricsUseCase.GetAllMetricsByType(repository.GaugeMetricKey)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	counterMetrics, err := s.metricsUseCase.GetAllMetricsByType(repository.CounterMetricKey)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
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

	err := s.metricsUseCase.TryUpdateMetricValue(metricType, metricName, metricValue)

	if err != nil {
		logger.Log.Errorf("Error with update metrics: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) UpdateMetricForJSON(w http.ResponseWriter, r *http.Request) {
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

	err = s.metricsUseCase.TryUpdateMetricValueFromJSON(&metricJSON)

	if err != nil {
		logger.Log.Errorf("Error with update metrics: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := json.Marshal(metricJSON)

	if err != nil {
		logger.Log.Errorf("Error with marshal output JSON: %w", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (s *Server) GetMetricByName(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	val, err := s.metricsUseCase.TryGetMetricValue(metricType, metricName)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch value := val.(type) {
	case int64:
		io.WriteString(w, fmt.Sprintf("%d", value))
	case float64:
		io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func (s *Server) GetMetricByNameFromJSON(w http.ResponseWriter, r *http.Request) {
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

	value, err := s.metricsUseCase.TryGetMetricValue(metricJSON.MType, metricJSON.ID)

	if err != nil {
		logger.Log.Infof("ERROR! %w", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch metricJSON.MType {
	case repository.CounterMetricKey:
		intValue, _ := value.(int64)
		metricJSON.Delta = &intValue
	case repository.GaugeMetricKey:
		floatValue, _ := value.(float64)
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
