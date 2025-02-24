package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/whynullname/go-collect-metrics/internal/storage"
)

type Server struct {
}

const (
	updateHandleFuncName = "/update/"
	adress               = "localhost:8080"
)

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc(updateHandleFuncName, UpdateData)
	return http.ListenAndServe(adress, mux)
}

func UpdateData(w http.ResponseWriter, r *http.Request) {
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

	switch parts[0] {
	case storage.CounterKey:
		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storage.MemoryStorage.UpdateCounterData(parts[1], i)
		w.WriteHeader(http.StatusOK)
	case storage.GaugeKey:
		i, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storage.MemoryStorage.UpdateGaugeData(parts[1], i)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
