package server

import (
	"flag"
	"metricapp/internal/logger"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type MetricServer struct{}

func (ms *MetricServer) Start() {
	var port string

	var cfg struct {
		Address string `env:"ADDRESS"`
	}
	err := env.Parse(&cfg)
	if err == nil {
		port = cfg.Address
	}

	if port == "" {
		flag.StringVar(&port, "a", "0.0.0.0:8080", "Порт на котором будет поднят сервер")
	}
	flag.Parse()

	logger.InitLogger()
	handler := NewMetricHandler()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Route("/", func(r chi.Router) {
		r.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
		r.Get("/value/{mType}/{mName}", handler.GetMetric)
	})

	logger.Info(
		"Start listening",
		zap.String("port", port),
	)
	http.ListenAndServe(port, router)
}
