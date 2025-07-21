package server

import (
	"compress/gzip"
	"flag"
	"log"
	"metricapp/internal/filemanager"
	"metricapp/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricServer struct{}

func (ms *MetricServer) Start() {
	var cfg struct {
		Address         string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
	}
	env.Parse(&cfg)

	if cfg.Address == "" {
		flag.StringVar(&cfg.Address, "a", "0.0.0.0:8080", "Порт на котором будет поднят сервер")
	}
	if cfg.StoreInterval == 0 {
		flag.IntVar(&cfg.StoreInterval, "i", 300, "Интервал записи метрик в файл")
	}
	if cfg.FileStoragePath == "" {
		flag.StringVar(&cfg.FileStoragePath, "f", "./metrics.log", "Путь к файлу с сохраненными метрика")
	}
	if !cfg.Restore {
		flag.BoolVar(&cfg.Restore, "r", false, "Флаг для загрузки сохраненных метрик с предыдущего сеанса")
	}
	flag.Parse()

	logger.InitLogger()

	fm, err := filemanager.Open(cfg.FileStoragePath, cfg.StoreInterval)
	if err != nil {
		log.Fatal("failed to open log file: ", err)
	}
	handler := NewMetricHandlerWfm(fm, cfg.Restore)
	if handler.fm.Storeinterval != 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(fm.Storeinterval) * time.Second)
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			for {
				select {
				case <-ticker.C:
					fm.Write(handler.storage.GetAllMetrics())
				case <-sigs:
					fm.Write(handler.storage.GetAllMetrics())
					logger.Info("exiting gracefully")
					fm.Close()
					os.Exit(0)
				}
			}
		}()
	}

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

		r.Post("/update/", handler.UpdateMetricsWJSONv2)
		r.Post("/value/", handler.GetMetricWJSONv2)
	})

	logger.Info(
		"Start listening",
		zap.String("port", cfg.Address),
	)

	http.ListenAndServe(cfg.Address, router)
	fm.Close()
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
