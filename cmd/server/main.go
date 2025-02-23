package main

import (
	"net/http"
	"strconv"
	"strings"

	storage "github.com/whynullname/go-collect-metrics/internal"
)

const (
	updateHandleFuncName = "/update/"
	port                 = ":8080"
)

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(updateHandleFuncName, updateData)
	return http.ListenAndServe(port, mux)
}

func updateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

		storage.MemoryStorage.UpdateCounter(parts[1], i)
		w.WriteHeader(http.StatusOK)
	case storage.GaugeKey:
		i, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		storage.MemoryStorage.UpdateGauge(parts[1], i)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
