package utils

import (
	"sync"
	"time"
)

type Cacheable[T any] interface {
	Set(value T)
	SetWithTTL(value T, ttl time.Duration)
	Get() (T, bool)
	IsSet() bool
}

type cacheable[T any] struct {
	value  T
	isSet  bool
	ttlEnd int64
	mu     sync.RWMutex
}

func (c *cacheable[T]) Get() (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isSet {
		return *new(T), false
	}

	if c.ttlEnd != -1 && c.ttlEnd <= time.Now().UTC().UnixMilli() {
		return *new(T), false
	}

	return c.value, true
}

func (c *cacheable[T]) Set(value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = value
	c.ttlEnd = -1
	c.isSet = true
}

func (c *cacheable[T]) SetWithTTL(value T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = value
	c.ttlEnd = time.Now().UTC().Add(ttl).UnixMilli()
	c.isSet = true
}

func (c *cacheable[T]) IsSet() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isSet {
		return false
	}

	if c.ttlEnd != -1 && c.ttlEnd <= time.Now().UTC().UnixMilli() {
		return false
	}

	return true
}

func NewCacheable[T any]() Cacheable[T] {
	return &cacheable[T]{
		value:  *new(T),
		isSet:  false,
		ttlEnd: 0,
		mu:     sync.RWMutex{},
	}
}
