package server

import (
	"compress/gzip"
	"flag"
	"io"
	"metricapp/internal/logger"
	"net/http"
	"time"

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
	router.Use(requestLogger)
	router.Use(GzipDecompressMiddleware)
	router.Use(middleware.Compress(1))

	router.Route("/", func(r chi.Router) {
		r.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
		r.Get("/value/{mType}/{mName}", handler.GetMetric)

		r.Post("/update/", handler.UpdateMetricsWJSONv2)
		r.Post("/value/", handler.GetMetricWJSONv2)
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
		msg    string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.msg = string(b)
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(gz)
		}
		next.ServeHTTP(w, r)
	})
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
			zap.String("Resp", responseData.msg),
		)
	}

	return http.HandlerFunc(logFn)
}
