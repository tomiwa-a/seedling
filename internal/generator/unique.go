package generator

import (
	"fmt"
	"sync"
)

type UniqueTracker struct {
	mu     sync.RWMutex
	values map[string]map[any]struct{}
}

func NewUniqueTracker() *UniqueTracker {
	return &UniqueTracker{
		values: make(map[string]map[any]struct{}),
	}
}

func (t *UniqueTracker) TryAdd(key string, value any) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	set, ok := t.values[key]
	if !ok {
		set = make(map[any]struct{})
		t.values[key] = set
	}

	if _, exists := set[value]; exists {
		return false
	}
	set[value] = struct{}{}
	return true
}

func (t *UniqueTracker) Add(key string, value any) error {
	if !t.TryAdd(key, value) {
		return fmt.Errorf("unique constraint violation for %q: %v", key, value)
	}
	return nil
}

func (t *UniqueTracker) Has(key string, value any) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	set, ok := t.values[key]
	if !ok {
		return false
	}
	_, exists := set[value]
	return exists
}

func (t *UniqueTracker) Generate(key string, maxAttempts int, fn func() any) (any, error) {
	for i := 0; i < maxAttempts; i++ {
		val := fn()
		if t.TryAdd(key, val) {
			return val, nil
		}
	}
	return nil, fmt.Errorf("unique tracker: exhausted %d attempts for %q", maxAttempts, key)
}

func (t *UniqueTracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.values, key)
}

func (t *UniqueTracker) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.values = make(map[string]map[any]struct{})
}
