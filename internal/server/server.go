package server

import (
	"flag"
	"metricapp/internal/logger"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
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
	router.Use(requestLogger)

	router.Route("/", func(r chi.Router) {
		r.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
		r.Get("/value/{mType}/{mName}", handler.GetMetric)

		r.Post("/update/", handler.UpdateMetricsWJSON)
		r.Post("/value/", handler.GetMetricWJSON)
	})

	logger.Info(
		"Start listening",
		zap.String("port", port),
	)
	http.ListenAndServe(port, router)
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func requestLogger(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method

		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Info(
			"Request log",
			zap.String("URI", uri),
			zap.String("Method", method),
			zap.Duration("Duration", duration),
			zap.Int("Status", responseData.status),
			zap.Int("Response size", responseData.size),
		)
	}

	return http.HandlerFunc(logFn)
}
