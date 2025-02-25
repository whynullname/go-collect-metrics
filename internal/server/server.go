package server

import (
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

	log.Println("Data received")

	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	i, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.storage.UpdateData(keyName, metricName, i)
	w.WriteHeader(http.StatusOK)
}
