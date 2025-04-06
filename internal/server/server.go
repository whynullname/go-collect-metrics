package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/middlewares"
	"github.com/whynullname/go-collect-metrics/internal/middlewares/compressmiddleware"
	"github.com/whynullname/go-collect-metrics/internal/repository/postgres"
	"github.com/whynullname/go-collect-metrics/internal/server/handlers"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Server struct {
	Config   *config.ServerConfig
	Router   chi.Router
	Handlers *handlers.Handlers
}

func NewServer(metricsUseCase *metrics.MetricsUseCase, config *config.ServerConfig, postgres *postgres.Postgres) *Server {
	serverInstance := &Server{
		Config:   config,
		Handlers: handlers.NewHandlers(metricsUseCase, postgres),
	}
	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()
	registerMiddlewares(r)

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.Handlers.GetAllMetrics)
		r.Get("/ping", s.Handlers.PingPostgres)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", s.Handlers.GetMetricByNameFromJSON)
			r.Get("/{metricType}/{metricName}", s.Handlers.GetMetricByName)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.Handlers.UpdateMetricFromJSON)
			r.Post("/{metricType}/{merticName}/{metricValue}", s.Handlers.UpdateMetric)
		})
	})
	return r
}

func registerMiddlewares(r chi.Router) {
	r.Use(middlewares.Logging)
	r.Use(compressmiddleware.GZIP)
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Config.EndPointAdress, s.Router)
}
