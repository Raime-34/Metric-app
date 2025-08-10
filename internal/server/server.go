package server

import (
	"compress/gzip"
	"log"
	"metricapp/internal/filemanager"
	"metricapp/internal/logger"
	"metricapp/internal/repository"
	"metricapp/internal/server/cfg"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricServer struct{}

func (ms *MetricServer) Start() {
	cfg.LoadConfig()
	logger.InitLogger()

	repository.NewPsqlHandler(cfg.Cfg.DSN, cfg.Cfg.MigrationPath)

	fm, err := filemanager.Open(cfg.Cfg.FileStoragePath, cfg.Cfg.StoreInterval)
	if err != nil {
		log.Fatal("failed to open log file: ", err)
	}
	handler := NewMetricHandlerWfm(fm, cfg.Cfg.Restore)

	router := chi.NewRouter()
	router.Use(gzipHandler)
	router.Use(requestLogger)

	router.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("Some text"))
		})

		r.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
		r.Get("/value/{mType}/{mName}", handler.GetMetric)

		r.Post("/update/", handler.UpdateMetricWJSONv2)
		r.Post("/updates/", handler.UpdateMultyMetrics)
		r.Post("/value/", handler.GetMetricWJSONv2)

		r.Get("/ping", handler.PingDB)
	})

	logger.Info(
		"Start listening",
		zap.String("port", cfg.Cfg.Address),
	)

	go http.ListenAndServe(cfg.Cfg.Address, router)
	defer fm.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var tickerC <-chan time.Time // STORE_INTERVAL = 0 -> tickerC = nil -> не будет тикать
	if handler.getStoreInterval() != 0 {
		ticker := time.NewTicker(time.Duration(handler.getStoreInterval()) * time.Second)
		tickerC = ticker.C
	}

outerLoop:
	for {
		select {
		case <-tickerC:
			fm.Write(handler.storage.GetAllMetrics())
		case <-sigs:
			fm.Write(handler.storage.GetAllMetrics())
			logger.Info("exiting gracefully")
			break outerLoop
		}
	}

	os.Exit(0)
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

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
