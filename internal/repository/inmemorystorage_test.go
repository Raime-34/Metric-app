package repository

import (
	models "metricapp/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_SetField(t *testing.T) {
	ms := NewInMemoryStorage()

	value := 213.0
	testMetric := models.Metrics{
		ID:    "TestGauge",
		MType: models.Gauge,
		Value: &value,
	}
	ms.SetField("TestGauge", testMetric)

	allMetrics := ms.GetFields()
	assert.Equal(t, testMetric, allMetrics["TestGauge"], "Gauge metrics are not the same")

	ms.IncrementCounter()
	allMetrics = ms.GetFields()
	assert.Equal(t, *allMetrics["PollCounter"].Delta, int64(1), "Counter metric are not incrementing")
}
