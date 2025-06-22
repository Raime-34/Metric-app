package handler

import (
	"context"
	models "metricapp/internal/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestMemeStorage_UpdateMetrics(t *testing.T) {
	handler := NewMetricHandler()

	type testcase struct {
		name         string
		expectedCode int
		mType        string
		mName        string
		mValue       string
	}

	cases := []testcase{
		{
			name:         "valid gauge case",
			expectedCode: http.StatusOK,
			mType:        models.Gauge,
			mName:        "test",
			mValue:       url.PathEscape("500.0"),
		},
		{
			name:         "valid ciunter case",
			expectedCode: http.StatusOK,
			mType:        models.Counter,
			mName:        "pollCounter",
			mValue:       "5",
		},
		{
			name:         "invalid gauge, without name",
			expectedCode: http.StatusNotFound,
			mType:        models.Gauge,
			mValue:       "500",
		},
		{
			name:         "invalid counter, without name",
			expectedCode: http.StatusNotFound,
			mType:        models.Counter,
			mValue:       "500",
		},
		{
			name:         "invalid gauge, wrong value",
			expectedCode: http.StatusBadRequest,
			mName:        "test",
			mType:        models.Gauge,
			mValue:       "some string",
		},
		{
			name:         "invalid counter, wrong value",
			expectedCode: http.StatusBadRequest,
			mName:        "test",
			mType:        models.Counter,
			mValue:       "some string",
		},
	}

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
	}
}
