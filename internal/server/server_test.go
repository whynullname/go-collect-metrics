package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whynullname/go-collect-metrics/internal/storage"
)

func TestUpdateData(t *testing.T) {
	tests := []struct {
		name        string
		targetUrl   string
		contentType string
		methodType  string
		wantCode    int
	}{
		{
			name:        "positive test #1 - counter metrics",
			targetUrl:   "/update/counter/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusOK,
		},
		{
			name:        "positive test #2 - gauge metrics",
			targetUrl:   "/update/gauge/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusOK,
		},
		{
			name:        "test json content type",
			targetUrl:   "/update/gauge/someMetrics/500",
			contentType: "application/json",
			methodType:  http.MethodPost,
			wantCode:    http.StatusMethodNotAllowed,
		},
		{
			name:        "test get method",
			targetUrl:   "/update/gauge/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodGet,
			wantCode:    http.StatusMethodNotAllowed,
		},
		{
			name:        "test bad url #1",
			targetUrl:   "/update/",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusNotFound,
		},
		{
			name:        "test bad url #2",
			targetUrl:   "/update/someData/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusBadRequest,
		},
	}

	storage := storage.NewStorage()
	serv := NewServer(*storage)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.methodType, test.targetUrl, nil)
			request.Header.Add("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			serv.UpdateData(w, request)

			res := w.Result()

			assert.Equal(t, test.wantCode, res.StatusCode)
		})
	}
}
