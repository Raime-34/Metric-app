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

	router.Route("/update/{mType}/{mName}", func(r chi.Router) {
		r.Post("/{mValue}", handler.UpdateMetrics)
		r.Get("/", nil)
	})

	http.ListenAndServe(":8080", router)
}
