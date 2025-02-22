package main

import (
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

var storage MemStorage

func main() {
	storage = MemStorage{
		gauge:   make(map[string]float64, 0),
		counter: make(map[string]int64, 0),
	}

	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateData)
	return http.ListenAndServe(`:8080`, mux)
}

func updateData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	rest := strings.TrimPrefix(path, "/update/")
	parts := strings.Split(rest, "/")

	switch parts[0] {
	case "counter":
		i, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.counter[parts[1]] = i
		w.WriteHeader(http.StatusOK)
	case "gauge":
		i, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.gauge[parts[1]] = i
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}
