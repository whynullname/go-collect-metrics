package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/whynullname/go-collect-metrics/internal/storage"
)

type Server struct {
	storage storage.MemoryStorage
}

const (
	updateHandleFuncName = "/update/"
	adress               = "localhost:8080"
)

func NewServer(storage storage.MemoryStorage) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(updateHandleFuncName, s.UpdateData)
	return http.ListenAndServe(adress, mux)
}

func (s *Server) UpdateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("Method not Post, return!")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		fmt.Println("Content type not text/plain, return!")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, updateHandleFuncName)
	parts := strings.Split(rest, "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println("Data received")
	switch parts[0] {
	case storage.CounterKey:
		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s.storage.UpdateCounterData(parts[1], i)
		w.WriteHeader(http.StatusOK)
	case storage.GaugeKey:
		i, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		s.storage.UpdateGaugeData(parts[1], i)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
