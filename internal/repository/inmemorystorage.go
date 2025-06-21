package repository

import (
	models "metricapp/internal/model"
	"sync"
)

type InMemoryStorage struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s *InMemoryStorage) SetField(key string, value models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[key] = value
}

func (s *InMemoryStorage) GetFields() map[string]models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Копируем мапу
	newMap := make(map[string]models.Metrics)
	for k, v := range s.metrics {
		newMap[k] = v
	}

	return newMap
}

func (s *InMemoryStorage) IncrementCounter() {
	pollCounter := s.metrics["PollCounter"]
	if pollCounter.Delta == nil {
		var zeroCounter int64
		pollCounter.MType = models.Counter
		pollCounter.Delta = &zeroCounter
	}

	newCounterValue := *pollCounter.Delta + 1
	pollCounter.Delta = &newCounterValue
	s.metrics["PollCounter"] = pollCounter
}
