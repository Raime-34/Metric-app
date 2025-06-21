package repository

import (
	models "metricapp/internal/model"
	"sync"
)

type Repo interface {
	SetField(string, models.Metrics)
	GetFields() map[string]models.Metrics
	IncrementCounter()
}

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
	newCounterValue := *pollCounter.Delta + 1
	pollCounter.Delta = &newCounterValue
	s.metrics["PollCounter"] = pollCounter
}
