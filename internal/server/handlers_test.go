package server

import (
	"bytes"
	"context"
	"encoding/json"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestMemeStorage_UpdateMetrics(t *testing.T) {
	logger.InitLogger()
	handler := NewMetricHandler()

	for _, c := range cases {
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("mType", c.mType)
		routeCtx.URLParams.Add("mName", c.mName)
		routeCtx.URLParams.Add("mValue", c.mValue)

		request := httptest.NewRequest(http.MethodPost, "/update", nil)
		request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		handler.UpdateMetrics(w, request)

		res := w.Result()
		assert.Equal(t, c.expectedCode, res.StatusCode)
		res.Body.Close()
	}
}

func TestMetricHandler_UpdateMetricsWJSON(t *testing.T) {
	logger.InitLogger()
	handler := NewMetricHandler()

	for _, c := range cases {
		// Создаем тело запроса, сначала как мапу
		metrics := make(map[string]any, 0)
		metrics["id"] = c.mName
		metrics["type"] = c.mType
		switch c.mType {
		case models.Gauge:
			parsedValue, _ := strconv.ParseFloat(c.mValue, 64)
			metrics["value"] = parsedValue
		case models.Counter:
			parsedValue, _ := strconv.ParseInt(c.mValue, 10, 64)
			metrics["delta"] = parsedValue
		}

		//Делаем из нее ридер
		b, _ := json.Marshal(metrics)
		r := bytes.NewReader(b)

		request := httptest.NewRequest(http.MethodPost, "/update", r)
		w := httptest.NewRecorder()
		handler.UpdateMetricsWJSON(w, request)

		res := w.Result()
		assert.Equal(t, c.expectedCode, res.StatusCode)
		res.Body.Close()
	}
}

type testcase struct {
	name         string
	expectedCode int
	mType        string
	mName        string
	mValue       string
}

var cases = []testcase{
	{
		name:         "valid gauge case",
		expectedCode: http.StatusOK,
		mType:        models.Gauge,
		mName:        "test",
		mValue:       url.PathEscape("500.0"),
	},
	// {
	// 	name:         "valid counter case",
	// 	expectedCode: http.StatusOK,
	// 	mType:        models.Counter,
	// 	mName:        "pollCounter",
	// 	mValue:       "5",
	// },
	// {
	// 	name:         "invalid gauge, without name",
	// 	expectedCode: http.StatusNotFound,
	// 	mType:        models.Gauge,
	// 	mValue:       "500",
	// },
	// {
	// 	name:         "invalid counter, without name",
	// 	expectedCode: http.StatusNotFound,
	// 	mType:        models.Counter,
	// 	mValue:       "500",
	// },
	// {
	// 	name:         "invalid gauge, wrong value",
	// 	expectedCode: http.StatusBadRequest,
	// 	mName:        "test",
	// 	mType:        models.Gauge,
	// 	mValue:       "some string",
	// },
	// {
	// 	name:         "invalid counter, wrong value",
	// 	expectedCode: http.StatusBadRequest,
	// 	mName:        "test",
	// 	mType:        models.Counter,
	// 	mValue:       "some string",
	// },
	// {
	// 	name:         "invalid metric",
	// 	expectedCode: http.StatusBadRequest,
	// 	mName:        "test",
	// 	mType:        "some string",
	// 	mValue:       "500",
	// },
}
