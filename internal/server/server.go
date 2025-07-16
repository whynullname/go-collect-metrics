package server

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-chi/chi/v5"
	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
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
	server   *http.Server
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
	r.Mount("/debug", pprofRouter())
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

func (s *Server) ListenAndServe(exit chan os.Signal, idleConn chan struct{}) error {
	s.server = &http.Server{
		Addr:    s.Config.EndPointAdress,
		Handler: s.Router,
	}
	go s.gracefullShutdown(exit, idleConn)
	return s.server.ListenAndServe()
}

func (s *Server) gracefullShutdown(exit chan os.Signal, idleConn chan struct{}) {
	<-exit

	if err := s.server.Shutdown(context.Background()); err != nil {
		logger.Log.Errorln("error while shutdown server %v\n", err)
	}

	idleConn <- struct{}{}
}

func pprofRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/pprof/", http.HandlerFunc(pprof.Index))
	r.Get("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Get("/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Get("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Post("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Get("/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.Get("/pprof/{name}", http.HandlerFunc(pprof.Index))
	return r
}
