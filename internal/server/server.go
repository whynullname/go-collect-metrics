package server

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository"
)

type Server struct {
	repository repository.Repository
	Config     *config.ServerConfig
	Router     chi.Router
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

func NewServer(repository repository.Repository, config *config.ServerConfig) *Server {
	serverInstance := &Server{
		repository: repository,
		Config:     config,
	}
	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.GetAllMetrics)
		r.Post("/update/{key}/{merticName}/{metricValue}", s.UpdateMetric)
		r.Get("/value/{metricType}/{metricName}", s.GetMetricByName)
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

	err = tmpl.Execute(w, s.repository)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && contentType != "text/plain" {
		log.Println("Content type not text/plain, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := chi.URLParam(r, "key")

	if metricType != repository.CounterMetricKey && metricType != repository.GaugeMetricKey {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricName := chi.URLParam(r, "merticName")
	metricValue := chi.URLParam(r, "metricValue")

	i, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//TODO: Refactor this
	log.Printf("Data received and updated! Key %s, metricaName %s, metricValue %s \n", metricType, metricName, metricValue)

	switch metricType {
	case repository.CounterMetricKey:
		s.repository.UpdateCounterMetricValue(metricName, int64(i))
		break
	case repository.GaugeMetricKey:
		s.repository.UpdateGaugeMetricValue(metricName, i)
		break
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetMetricByName(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	log.Printf("Try get metric type %s \n", metricType)
	if metricType != repository.CounterMetricKey && metricType != repository.GaugeMetricKey {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricName := chi.URLParam(r, "metricName")

	val := float64(0)
	ok := false
	switch metricType {
	case repository.CounterMetricKey:
		val, ok = s.repository.TryGetCounterMetricValue(metricName)
		break
	case repository.GaugeMetricKey:
		val, ok = s.repository.TryGetGaugeMetricValue(metricName)
		break
	}

	if !ok {
		log.Printf("Can't get mertic value for %s \n", metricName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatFloat(val, 'f', -1, 64))
}
