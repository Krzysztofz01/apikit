package utils

import (
	"sync"
	"time"
)

// Generic key-value mapping based cache
type Cache[T any] interface {
	// Set or overwrite the given value in the cache represented by a given key with a infinite TTL
	Set(key string, value T)

	// Set or overwrite the given value in the cache represented by a given key with a given TTL
	SetWithTTL(key string, value T, ttl time.Duration)

	// Set or overwiret the given value in the cache represented by a given key with a given TTL end
	SetWithTTLEnd(key string, value T, ttlEnd time.Time)

	// Retrieve a value from the cache represented by a given key. The returned boolean value represents if the operation was successful
	Get(key string) (T, bool)

	// Check if a value represented by a given key is stored in the cache and has valid TTL
	IsSet(key string) bool

	// Remove a value from the cache represented by a given key. The returned boolean value represents if the operation was successful
	Remove(key string) bool
}

const (
	defaultCacheCapacity = 25
)

// Create a new empty in-memory cache instance
func NewEmptyCache[T any]() Cache[T] {
	return &inMemoryMapCache[T]{
		values: make(map[string]*inMemoryMapCacheEntry[T], defaultCacheCapacity),
		mu:     sync.RWMutex{},
	}
}

type inMemoryMapCache[T any] struct {
	values map[string]*inMemoryMapCacheEntry[T]
	mu     sync.RWMutex
}

type inMemoryMapCacheEntry[T any] struct {
	value  T
	ttlEnd int64
}

func (entry *inMemoryMapCacheEntry[T]) CheckTTL() bool {
	return entry.ttlEnd == -1 || entry.ttlEnd <= time.Now().UTC().Unix()
}

func (cache *inMemoryMapCache[T]) Set(key string, value T) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.values[key] = &inMemoryMapCacheEntry[T]{
		value:  value,
		ttlEnd: -1,
	}
}

func (cache *inMemoryMapCache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	cache.SetWithTTLEnd(key, value, time.Now().UTC().Add(ttl))
}

func (cache *inMemoryMapCache[T]) SetWithTTLEnd(key string, value T, ttlEnd time.Time) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if time.Now().UTC().Unix() <= ttlEnd.Unix() {
		return
	}

	cache.values[key] = &inMemoryMapCacheEntry[T]{
		value:  value,
		ttlEnd: ttlEnd.Unix(),
	}
}

func (cache *inMemoryMapCache[T]) Get(key string) (T, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	entry, ok := cache.values[key]
	if !ok || !entry.CheckTTL() {
		return *new(T), false
	}

	return entry.value, true
}

func (cache *inMemoryMapCache[T]) IsSet(key string) bool {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	entry, ok := cache.values[key]

	return ok && entry.CheckTTL()
}

func (cache *inMemoryMapCache[T]) Remove(key string) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if _, ok := cache.values[key]; !ok {
		return false
	}

	delete(cache.values, key)
	return true
}
