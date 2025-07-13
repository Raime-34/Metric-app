package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	storage.IncrementCounter(struct {
		Name  string
		Delta int64
	}{Name: "Test", Delta: inc})

	counter, ok := storage.GetCounter("Test")
	require.True(t, ok)

	assert.Equal(t, inc, counter)
}
