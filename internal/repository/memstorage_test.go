package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_SetField(t *testing.T) {
	storage := NewMemStorage()

	key := "Test"
	var expected float64 = 5
	storage.SetField(key, expected)
	actual := storage.GetFields()[key]
	assert.Equal(t, expected, actual)
}

func TestMemStorage_IncrementCounter(t *testing.T) {
	storage := NewMemStorage()

	var inc int64 = 10
	storage.IncrementCounter(Counter{Name: "Test", Delta: inc})

	assert.Equal(t, inc, storage.counter.Load())
}
