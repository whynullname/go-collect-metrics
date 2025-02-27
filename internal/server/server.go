package server

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

type Server struct {
	storage        *storage.MemoryStorage
	endPointAdress string
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

func NewServer(storage *storage.MemoryStorage, endPointAdress string) *Server {
	serverInstance := &Server{
		storage:        storage,
		endPointAdress: endPointAdress,
	}
	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.GetAllData)
		r.Post("/update/{key}/{merticName}/{metricValue}", s.UpdateData)
		r.Get("/value/{metricType}/{metricName}", s.GetData)
	})
	return r
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.endPointAdress, s.Router)
}

func (s *Server) GetAllData(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("webpage").Parse(allDataHTMLTemplate)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, s.storage)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) UpdateData(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if contentType != "" && contentType != "text/plain" {
		log.Println("Content type not text/plain, return!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyName := chi.URLParam(r, "key")

	if keyName != storage.CounterKey && keyName != storage.GaugeKey {
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

	log.Printf("Data received and updated! Key %s, metricaName %s, metricValue %s \n", keyName, metricName, metricValue)
	s.storage.UpdateData(keyName, metricName, i)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetData(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	log.Printf("Try get metric type %s \n", metricType)
	if metricType != storage.CounterKey && metricType != storage.GaugeKey {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricName := chi.URLParam(r, "metricName")

	val, ok := s.storage.GetData(metricType, metricName)

	if !ok {
		log.Printf("Can't get mertic value for %s \n", metricName)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	io.WriteString(w, strconv.FormatFloat(val, 'f', -1, 64))
}
