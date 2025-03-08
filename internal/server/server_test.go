package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whynullname/go-collect-metrics/internal/agent"
	configAgent "github.com/whynullname/go-collect-metrics/internal/configs/agentconfig"
	configServer "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/repository/inmemory"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

func TestUpdateData(t *testing.T) {
	tests := []struct {
		name        string
		targetURL   string
		contentType string
		methodType  string
		wantCode    int
	}{
		{
			name:        "positive test #1 - counter metrics",
			targetURL:   "/update/counter/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusOK,
		},
		{
			name:        "positive test #2 - gauge metrics",
			targetURL:   "/update/gauge/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusOK,
		},
		{
			name:        "test json content type",
			targetURL:   "/update/gauge/someMetrics/500",
			contentType: "application/json",
			methodType:  http.MethodPost,
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "test get method",
			targetURL:   "/update/gauge/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodGet,
			wantCode:    http.StatusMethodNotAllowed,
		},
		{
			name:        "test bad url #1",
			targetURL:   "/update/",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusNotFound,
		},
		{
			name:        "test bad url #2",
			targetURL:   "/update/someData/someMetrics/500",
			contentType: "text/plain",
			methodType:  http.MethodPost,
			wantCode:    http.StatusBadRequest,
		},
		{
			name:        "test bad gauge value",
			targetURL:   "/update/gauge/someMetric/badValue",
			contentType: "text/plaint",
			methodType:  http.MethodPost,
			wantCode:    http.StatusBadRequest,
		},
	}

	repo := inmemory.NewInMemoryRepository()
	cfg := configServer.NewServerConfig()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	serv := NewServer(metricsUseCase, cfg)
	client := httptest.NewServer(serv.Router)
	defer client.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.methodType, client.URL+test.targetURL, nil)
			request.RequestURI = ""
			request.Header.Add("Content-Type", test.contentType)

			resp, err := client.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.wantCode, resp.StatusCode)
		})
	}
}

func TestGetData(t *testing.T) {
	memStats := runtime.MemStats{}
	repo := inmemory.NewInMemoryRepository()
	agentCfg := configAgent.NewAgentConfig()
	serverCfg := configServer.NewServerConfig()
	metricsUseCase := metrics.NewMetricUseCase(repo)
	agent := agent.NewAgent(&memStats, metricsUseCase, agentCfg)
	serv := NewServer(metricsUseCase, serverCfg)
	client := httptest.NewServer(serv.Router)
	defer client.Close()
	agent.UpdateMetrics()

	tests := []struct {
		name       string
		method     string
		url        string
		headerCode int
		response   string
	}{
		{
			name:       "positive test #1",
			method:     http.MethodGet,
			url:        "/value/gauge/Alloc",
			headerCode: http.StatusOK,
			response:   strconv.FormatUint(memStats.Alloc, 10),
		},
		{
			name:       "positive test #2",
			method:     http.MethodGet,
			url:        "/value/gauge/NextGC",
			headerCode: http.StatusOK,
			response:   strconv.FormatUint(memStats.NextGC, 10),
		},
		{
			name:       "bad http method",
			method:     http.MethodPost,
			url:        "/value/gauge/NextGC",
			headerCode: http.StatusMethodNotAllowed,
			response:   "",
		},
		{
			name:       "bad data type",
			method:     http.MethodGet,
			url:        "/value/badDataType/someData",
			headerCode: http.StatusNotFound,
			response:   "",
		},
		{
			name:       "bad data name",
			method:     http.MethodGet,
			url:        "/value/gauge/badDataName",
			headerCode: http.StatusNotFound,
			response:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, client.URL+test.url, nil)
			request.RequestURI = ""
			resp, err := client.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, test.headerCode, resp.StatusCode)

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.response, string(data))
		})
	}
}
