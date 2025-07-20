package server

import (
	"encoding/json"
	"io"
	"metricapp/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type MetricHandler struct {
	storage repository.MemStorage
}

func NewMetricHandler() *MetricHandler {
	return &MetricHandler{
		storage: repository.NewMemStorage(),
	}
}

func (h *MetricHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	name := chi.URLParam(r, "mName")
	value := chi.URLParam(r, "mValue")

	metrics := struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}{
		ID:    name,
		Type:  mType,
		Value: value,
	}
	if err := h.storage.ProcessMetric(metrics); err != nil {
		switch err {
		case repository.ErrMetricIsRequired:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func (h *MetricHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")

	b, _, err := h.storage.ProcessGetField(mName, mType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Write(b)
}

func (h *MetricHandler) UpdateMetricsWJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var metrics struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if metrics.ID == "PollCount" {
		metrics.ID = "PollCounter"
	}
	if err := h.storage.ProcessMetric(metrics); err != nil {
		switch err {
		case repository.ErrMetricIsRequired:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func (h *MetricHandler) GetMetricWJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error to read request body", http.StatusInternalServerError)
		return
	}

	var payload struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	err = json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}{
		ID:   payload.ID,
		Type: payload.Type,
	}

	if payload.ID == "PollCount" {
		payload.ID = "PollCounter"
	}
	_, v, err := h.storage.ProcessGetField(payload.ID, payload.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp.Value = v
	b, _ = json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
