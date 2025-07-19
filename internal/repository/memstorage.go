package repository

import (
	"errors"
	"fmt"
	models "metricapp/internal/model"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type MemStorage struct {
	storage  map[string]float64
	counters map[string]int64
	counter  atomic.Int64
	mu       sync.RWMutex
}

func NewMemStorage() MemStorage {
	return MemStorage{
		storage:  make(map[string]float64),
		counters: make(map[string]int64),
	}
}

var (
	ErrMetricIsRequired    = errors.New("metric name is required")
	ErrUnknownMetricType   = errors.New("unknown metric type")
	ErrInvalidGaugeValue   = errors.New("failed to parse gauge value")
	ErrInvalidCounterValue = errors.New("failed to parse counter value")

	ErrUnknownMetric  = errors.New("unknown metric name")
	ErrUnknownCounter = errors.New("unknown counter")
)

func (ms *MemStorage) ProcessMetric(metric models.Metrics) error {
	if metric.ID == "" {
		return ErrMetricIsRequired
	}

	switch metric.MType {
	case models.Gauge:
		ms.SetField(metric.ID, *metric.Value)
	case models.Counter:

		ms.IncrementCounter(struct {
			Name  string
			Delta int64
		}{
			Name:  metric.ID,
			Delta: *metric.Delta,
		})
	default:
		return ErrUnknownMetricType
	}

	return nil
}

func (ms *MemStorage) SetField(key string, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.storage[key] = value
}

func (ms *MemStorage) GetFields() map[string]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	newMap := make(map[string]float64)
	for key, value := range ms.storage {
		newMap[key] = value
	}

	return newMap
}

func (ms *MemStorage) ProcessGetField(mName string, mType string) ([]byte, any, error) {
	switch mType {
	case models.Gauge:
		v, ok := ms.GetField(mName)
		if !ok {
			return nil, nil, ErrUnknownMetric
		}

		s := strconv.FormatFloat(v, 'f', 3, 64)
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
		return []byte(s), v, nil
	case models.Counter:
		counter, ok := ms.GetCounter(mName)
		if !ok {
			return nil, nil, ErrUnknownCounter
		}

		s := strconv.Itoa(int(counter))
		return []byte(s), counter, nil
	}

	return nil, nil, fmt.Errorf("unknown error")
}

func (ms *MemStorage) GetField(name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	value, ok := ms.storage[name]
	return value, ok
}

func (ms *MemStorage) IncrementCounter(n ...struct {
	Name  string
	Delta int64
}) {
	if len(n) == 0 {
		return
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[n[0].Name] = ms.counters[n[0].Name] + n[0].Delta
}

func (ms *MemStorage) GetCounter(name string) (counter int64, ok bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	counter, ok = ms.counters[name]
	return
}
