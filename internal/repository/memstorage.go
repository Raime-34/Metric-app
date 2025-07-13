package repository

import (
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
