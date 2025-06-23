package handler

import (
	"metricapp/internal/logger"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Start() {
	logger.InitLogger()
	router := chi.NewRouter()
	handler := NewMetricHandler()

	router.Route("/", func(r chi.Router) {
		r.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
		r.Get("/value/{mType}/{mName}", handler.GetMetric)
	})

	http.ListenAndServe(":8080", router)
}
