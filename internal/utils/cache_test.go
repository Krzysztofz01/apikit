package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheShouldCreate(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)
}

func TestCacheShouldAddValues(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)

	key := "example-key"
	expectedValue := 3

	cache.Set(key, expectedValue)
}

func TestCacheShouldTellIfKeyIsSet(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)

	key := "example-key"
	value := 3

	assert.False(t, cache.IsSet(key))

	cache.Set(key, value)

	assert.True(t, cache.IsSet(key))

	assert.True(t, cache.Remove(key))

	assert.False(t, cache.IsSet(key))
}

func TestCacheShouldAddAndGetValues(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)

	key := "example-key"
	expectedValue := 3

	_, ok := cache.Get(key)

	assert.False(t, ok)

	cache.Set(key, expectedValue)

	actualValue, ok := cache.Get(key)

	assert.True(t, ok)
	assert.Equal(t, expectedValue, actualValue)
}

func TestCacheShouldAddAndUpdateValues(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)

	key := "example-key"
	firstExpectedValue := 3

	cache.Set(key, firstExpectedValue)

	actualValue, ok := cache.Get(key)

	assert.True(t, ok)
	assert.Equal(t, firstExpectedValue, actualValue)

	secondExpectedValue := 4

	cache.Set(key, secondExpectedValue)

	actualValue, ok = cache.Get(key)

	assert.True(t, ok)
	assert.Equal(t, secondExpectedValue, actualValue)
}

func TestCacheShouldAddAndRemoveValues(t *testing.T) {
	cache := NewEmptyCache[int]()

	assert.NotNil(t, cache)

	key := "example-key"
	expectedValue := 3

	ok := cache.Remove(key)

	assert.False(t, ok)

	cache.Set(key, expectedValue)

	ok = cache.Remove(key)

	assert.True(t, ok)
}
