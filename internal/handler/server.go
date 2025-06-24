package handler

import (
	"flag"
	"fmt"
	"metricapp/internal/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Start() {
	var port string
	flag.Func("a", "Порт на котором будет поднят сервер", func(s string) error {
		port = fmt.Sprintf(":%s", s)
		return nil
	})
	flag.Parse()

	if port == "" {
		flag.Set("a", "8080")
	}

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
