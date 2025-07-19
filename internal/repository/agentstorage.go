package repository

import (
	models "metricapp/internal/model"
	"sync"
)

type AgentMemStorage struct {
	metrics map[string]models.Metrics
	mu      sync.RWMutex
}

func NewAgentMemoryStorage() *AgentMemStorage {
	return &AgentMemStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s *AgentMemStorage) SetField(key string, value models.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[key] = value
}

func (s *AgentMemStorage) GetFields() map[string]models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Копируем мапу
	newMap := make(map[string]models.Metrics)
	for k, v := range s.metrics {
		newMap[k] = v
	}

	// s.refreshPollCounter()
	return newMap
}

func (s *AgentMemStorage) refreshPollCounter() {
	s.metrics["PollCount"] = models.ComposeMetrics("PollCount", models.Counter, 0, 0)
}

func (s *AgentMemStorage) IncrementCounter(n ...struct {
	Name  string
	Delta int64
}) {
	pollCounter := s.metrics["PollCount"]
	if pollCounter.Delta == nil {
		var zeroCounter int64
		pollCounter.MType = models.Counter
		pollCounter.Delta = &zeroCounter
	}

	newCounterValue := *pollCounter.Delta + 1
	pollCounter.Delta = &newCounterValue
	pollCounter.ID = "PollCount"
	s.metrics["PollCount"] = pollCounter
}
