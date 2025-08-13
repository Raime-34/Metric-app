package server

import (
	"encoding/json"
	"fmt"
	"io"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"metricapp/internal/server/cfg"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Имело бы смысл засунуть psqlHandler в dbHandler, но пока оставлю пустым
type DBHandler struct{}

func NewDBHandler(dsn string, mPath string) *DBHandler {
	repository.NewPsqlHandler(cfg.Cfg.DSN, cfg.Cfg.MigrationPath)

	return &DBHandler{}
}

func (h *DBHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	name := chi.URLParam(r, "mName")
	value := chi.URLParam(r, "mValue")

	switch mType {
	case models.Gauge:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, "failed to parse mValue", errBadReq)
			return
		}

		err = repository.UpdateGauge(name, v)
		if err != nil {
			http.Error(w, "failed to update gauge", errInternal)
			return
		}

	case models.Counter:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(w, "failed to parse mValue", errBadReq)
			return
		}

		err = repository.IncrementCounter(name, v)
		if err != nil {
			http.Error(w, "failed to update counter", errInternal)
			return
		}
	}

}

func (h *DBHandler) UpdateMultyMetrics(w http.ResponseWriter, r *http.Request) {
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

	if cfg.Cfg.DSN == "" {
		http.Error(w, "db is not initialized", errInternal)
		return
	}

	err = repository.InsertBatch(r.Context(), metrics)
	if err != nil {
		http.Error(w, "failed to make transaction", errInternal)
		return
	}
}

func (h *DBHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")

	metric, err := repository.QueryRow(r.Context(), mType, mName)
	if err != nil {
		http.Error(w, "failed to get metric", errBadReq)
		return
	}

	b, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "failed to marshal data", errInternal)
		return
	}

	w.Write(b)
}

func (h *DBHandler) UpdateMetricsWJSON(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(errInternal), errInternal)
		return
	}
	defer r.Body.Close()

	var metric struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	err = json.Unmarshal(b, &metric)
	if err != nil {
		http.Error(w, http.StatusText(errBadReq), errBadReq)
		return
	}

	switch metric.Type {
	case models.Gauge:
		v := metric.Value.(float64)
		err := repository.UpdateGauge(metric.ID, v)
		if err != nil {
			http.Error(w, "filed to update GAUGE", errInternal)
			return
		}

	case models.Counter:
		v := metric.Value.(int64)
		err := repository.IncrementCounter(metric.ID, v)
		if err != nil {
			http.Error(w, "filed to update COUNTER", errInternal)
			return
		}

	default:
		http.Error(w, fmt.Sprintf("unknown metric type: %s", metric.Type), errBadReq)
	}
}

func (h *DBHandler) UpdateMetricWJSONv2(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(errInternal), errInternal)
		return
	}
	defer r.Body.Close()

	var metric struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	err = json.Unmarshal(b, &metric)
	if err != nil {
		http.Error(w, http.StatusText(errBadReq), errBadReq)
		return
	}

	switch metric.Type {
	case models.Gauge:
		v := metric.Value.(float64)
		err := repository.UpdateGauge(metric.ID, v)
		if err != nil {
			http.Error(w, "filed to update GAUGE", errInternal)
			return
		}

	case models.Counter:
		v := metric.Value.(int64)
		err := repository.IncrementCounter(metric.ID, v)
		if err != nil {
			http.Error(w, "filed to update COUNTER", errInternal)
			return
		}

	default:
		http.Error(w, fmt.Sprintf("unknown metric type: %s", metric.Type), errBadReq)
	}
}

func (h *DBHandler) GetMetricWJSON(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "failed to parse payload", errBadReq)
		return
	}

	metric, err := repository.QueryRow(r.Context(), payload.Type, payload.ID)
	if err != nil {
		http.Error(w, "failed get metric", errInternal)
		return
	}

	b, err = json.Marshal(metric)
	if err != nil {
		http.Error(w, "failed to marshal metric", errInternal)
		return
	}

	w.Write(b)
}

func (h *DBHandler) GetMetricWJSONv2(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "failed to parse payload", errBadReq)
		return
	}

	metric, err := repository.QueryRow(r.Context(), payload.Type, payload.ID)
	if err != nil {
		http.Error(w, "failed get metric", errInternal)
		return
	}

	b, err = json.Marshal(metric)
	if err != nil {
		http.Error(w, "failed to marshal metric", errInternal)
		return
	}

	w.Write(b)
}

func (h *DBHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	err := repository.Ping()
	if err != nil {
		http.Error(w, "Error: database is not responding", errInternal)
		return
	}
}

func (h *DBHandler) getStoreInterval() int {
	return 1_000
}

func (h *DBHandler) GetStorage() *repository.MemStorage {
	return nil
}
