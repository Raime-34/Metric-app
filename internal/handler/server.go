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
	router.Post("/update/{mType}/{mName}/{mValue}", handler.UpdateMetrics)
	http.ListenAndServe(":8080", router)
}
