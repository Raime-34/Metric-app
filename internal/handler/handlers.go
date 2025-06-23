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
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}
}

func (h *MetricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")

	switch mType {
	case models.Gauge:
		v, ok := h.storage.GetField(mName)
		if !ok {
			http.Error(w, "unknown metric name", http.StatusBadRequest)
			return
		}

		s := strconv.Itoa(int(v))
		w.Write([]byte(s))
	case models.Counter:
		s := strconv.Itoa(int(h.storage.GetCounter()))
		w.Write([]byte(s))
	}
}
