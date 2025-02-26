package server

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

type Server struct {
	storage *storage.MemoryStorage
	Router  chi.Router
}

const (
	updateHandleFuncName = "/update/"
	adress               = "localhost:8080"
)

func NewServer(storage *storage.MemoryStorage) *Server {
	serverInstance := &Server{
		storage: storage,
	}
	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Post("/{key}/{merticName}/{metricValue}", s.UpdateData)
	})
	r.Get("/value/{metricType}/{metricName}", s.GetData)
	return r
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(adress, s.Router)
}

func (s *Server) UpdateData(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		log.Println("Content type not text/plain, return!")
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	io.WriteString(w, strconv.FormatFloat(val, 'f', 6, 64))
}
