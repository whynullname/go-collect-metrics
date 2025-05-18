package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/middlewares"
	"github.com/whynullname/go-collect-metrics/internal/middlewares/compressmiddleware"
	"github.com/whynullname/go-collect-metrics/internal/middlewares/shamiddleware"
	"github.com/whynullname/go-collect-metrics/internal/server/handlers"
	"github.com/whynullname/go-collect-metrics/internal/usecase/metrics"
)

type Server struct {
	Config   *config.ServerConfig
	Router   chi.Router
	Handlers *handlers.Handlers
}

func NewServer(metricsUseCase *metrics.MetricsUseCase, config *config.ServerConfig, pingRepoFunc func() bool) *Server {
	serverInstance := &Server{
		Config:   config,
		Handlers: handlers.NewHandlers(metricsUseCase, pingRepoFunc),
	}
	serverInstance.Router = serverInstance.createRouter()
	return serverInstance
}

func (s *Server) createRouter() chi.Router {
	r := chi.NewRouter()
	s.registerMiddlewares(r)

	r.Route("/", func(r chi.Router) {
		r.Get("/", s.Handlers.GetAllMetrics)
		r.Get("/ping", s.Handlers.PingRepository)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", s.Handlers.GetMetricByNameFromJSON)
			r.Get("/{metricType}/{metricName}", s.Handlers.GetMetricByName)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.Handlers.UpdateMetricFromJSON)
			r.Post("/{metricType}/{merticName}/{metricValue}", s.Handlers.UpdateMetric)
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", s.Handlers.UpdateArrayJSONMetrics)
		})
	})
	return r
}

func (s *Server) registerMiddlewares(r chi.Router) {
	r.Use(middlewares.Logging)
	r.Use(compressmiddleware.GZIP)
	r.Use(shamiddleware.HashSHA256(s.Config))
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Config.EndPointAdress, s.Router)
}
