package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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
		http.Error(w, "error to read request body: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// if metrics.ID == "PollCount" {
	// 	metrics.ID = "PollCounter"
	// }
	if err := h.storage.ProcessMetric(metrics); err != nil {
		switch err {
		case repository.ErrMetricIsRequired:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	var v any
	switch metrics.Type {
	case models.Gauge:
		v, _ = h.storage.GetField(metrics.ID)
	case models.Counter:
		v, _ = h.storage.GetCounter(metrics.ID)
	}

	resp := make(map[string]any)
	resp["id"] = metrics.ID
	resp["type"] = metrics.Type
	resp["value"] = v

	b, _ = json.Marshal(resp)
	w.Write(b)

	logger.Info(
		"UPDATE",
		zap.Any("metric", metrics),
	)
}

func (h *MetricHandler) UpdateMetricsWJSONv2(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error to read request body: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	switch metrics.Type {
	case models.Gauge:
		var gMetrics struct {
			ID    string  `json:"id"`
			Type  string  `json:"type"`
			Value float64 `json:"value"`
		}

		err := json.Unmarshal(b, &gMetrics)
		if err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		h.storage.SetField(gMetrics.ID, gMetrics.Value)

	case models.Counter:
		var cMetrics struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Value int64  `json:"delta"`
		}

		err := json.Unmarshal(b, &cMetrics)
		if err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		h.storage.IncrementCounter(struct {
			Name  string
			Delta int64
		}{Name: cMetrics.ID, Delta: cMetrics.Value})
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
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
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

	logger.Info(
		"READ",
		zap.Any("payload", payload),
	)

	// if payload.ID == "PollCount" {
	// 	payload.ID = "PollCounter"
	// }
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

func gzipDecompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	var out bytes.Buffer
	_, err = io.Copy(&out, gr)
	if err != nil {
		return nil, err
	}

	logger.Info(
		"decomressed",
		zap.String("msg", out.String()),
	)
	return out.Bytes(), nil
}

func (h *MetricHandler) GetMetricWJSONv2(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error to read request body", http.StatusInternalServerError)
		return
	}

	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		b, err = gzipDecompress(b)
		if err != nil {
			http.Error(w, "failed to decompress data: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var payload struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	err = json.Unmarshal(b, &payload)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	switch payload.Type {
	case models.Gauge:
		v, ok := h.storage.GetField(payload.ID)
		if !ok {
			http.Error(w, "unknown metric", http.StatusNotFound)
			return
		}

		resp := struct {
			ID    string  `json:"id"`
			Type  string  `json:"type"`
			Value float64 `json:"value"`
		}{
			ID:    payload.ID,
			Type:  payload.Type,
			Value: v,
		}

		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

	case models.Counter:
		v, ok := h.storage.GetCounter(payload.ID)
		if !ok {
			http.Error(w, "unknown metric", http.StatusNotFound)
			return
		}

		resp := struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Value int64  `json:"delta"`
		}{
			ID:    payload.ID,
			Type:  payload.Type,
			Value: v,
		}

		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}
