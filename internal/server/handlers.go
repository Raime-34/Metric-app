package server

import (
	"encoding/json"
	"io"
	"metricapp/internal/filemanager"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricHandler struct {
	storage repository.MemStorage
	fm      *filemanager.FManager
}

const (
	errInternal = http.StatusInternalServerError
	errBadReq   = http.StatusBadRequest
)

func NewMetricHandlerWfm(fm *filemanager.FManager, restore bool) *MetricHandler {
	handler := &MetricHandler{
		storage: repository.NewMemStorage(),
		fm:      fm,
	}

	if restore {
		metrics, err := fm.Read()
		if err == nil {
			for _, m := range metrics {
				switch m.MType {
				case models.Gauge:
					handler.storage.SetField(m.ID, *m.Value)
				case models.Counter:
					handler.storage.IncrementCounter(struct {
						Name  string
						Delta int64
					}{
						Name:  m.ID,
						Delta: *m.Delta,
					})
				}
			}
		}
	}

	return handler
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

	if h.getStoreInterval() == 0 {
		h.write(h.storage.GetAllMetrics())
	}
}

func (h *MetricHandler) write(metrics []models.Metrics) error {
	if h.fm != nil {
		return h.fm.Write(metrics)
	}

	return nil
}

func (h *MetricHandler) read() ([]models.Metrics, error) {
	if h.fm != nil {
		return h.fm.Read()
	}

	return nil, nil
}

func (h *MetricHandler) getStoreInterval() int {
	if h.fm == nil {
		return 1_000
	}

	return h.fm.Storeinterval
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
		http.Error(w, http.StatusText(errInternal), errInternal)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Error(
				"failed to close request body",
				zap.Error(err),
			)
		}
	}()

	var metrics struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		http.Error(w, http.StatusText(errBadReq), errBadReq)
		return
	}

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

	if h.getStoreInterval() == 0 {
		h.fm.Write(h.storage.GetAllMetrics())
	}
}

func (h *MetricHandler) UpdateMetricWJSONv2(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(errInternal), errInternal)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Error(
				"failed to close request body",
				zap.Error(err),
			)
		}
	}()

	var metrics struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		http.Error(w, http.StatusText(errBadReq), errBadReq)
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
			http.Error(w, http.StatusText(errBadReq), errBadReq)
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
			http.Error(w, http.StatusText(errBadReq), errBadReq)
			return
		}

		h.storage.IncrementCounter(struct {
			Name  string
			Delta int64
		}{Name: cMetrics.ID, Delta: cMetrics.Value})
	}
}

func (h *MetricHandler) UpdateMultyMetrics(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(errInternal), errInternal)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			logger.Error(
				"failed to close request body",
				zap.Error(err),
			)
		}
	}()

	var metrics []models.Metrics
	err = json.Unmarshal(b, &metrics)
	if err != nil {
		http.Error(w, "failed to parse data", errBadReq)
		return
	}

	err = h.storage.ProcessMultyMetrics(r.Context(), metrics)
	if err != nil {
		logger.Error("failed to update m-metrics", zap.Error(err))
		http.Error(w, "failed to update m-metrics", errInternal)
		return
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
		http.Error(w, http.StatusText(errBadReq), errBadReq)
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

func (h *MetricHandler) GetMetricWJSONv2(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, http.StatusText(errBadReq), errBadReq)
		return
	}

	switch payload.Type {
	case models.Gauge:
		v, ok := h.storage.GetField(payload.ID)
		if !ok {
			http.Error(w, "unknown metric", http.StatusNotFound)
			return
		}

		resp := models.Metrics{
			ID:    payload.ID,
			MType: payload.Type,
			Value: &v,
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

		resp := models.Metrics{
			ID:    payload.ID,
			MType: payload.Type,
			Delta: &v,
		}

		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

func (h *MetricHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	err := repository.Ping()
	if err != nil {
		http.Error(w, "Error: database is not responding", errInternal)
		return
	}
}
