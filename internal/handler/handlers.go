package handler

import (
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MetricHandler struct {
	storage repository.MemStorage
}

func NewMetricHandler() MetricHandler {
	return MetricHandler{
		storage: repository.NewMemStorage(),
	}
}

func (h *MetricHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	name := chi.URLParam(r, "mName")
	value := chi.URLParam(r, "mValue")

	if name == "" {
		http.Error(w, "metric name is required", http.StatusNotFound)
		return
	}

	switch mType {
	case models.Gauge:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "failed to parse value", http.StatusBadRequest)
			return
		}
		h.storage.SetField(name, parsedValue)
	case models.Counter:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "failed to parse value", http.StatusBadRequest)
			return
		}

		h.storage.IncrementCounter(parsedValue)
	}
}
