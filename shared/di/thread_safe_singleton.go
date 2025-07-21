package di

import (
	"sync"
	"sync/atomic"
)

type ThreadSafeSingleton[T any] struct {
	instance atomic.Value // stores T
	once     sync.Once
	creator  func() T
}

func NewThreadSafeSingleton[T any](creator func() T) *ThreadSafeSingleton[T] {
	return &ThreadSafeSingleton[T]{
		creator: creator,
	}
}

func (ts *ThreadSafeSingleton[T]) Acquire() T {
	if instance := ts.instance.Load(); instance != nil {
		return instance.(T)
	}

	var result T
	ts.once.Do(func() {
		result = ts.creator()
		ts.instance.Store(result)
	})

	return result
}

// Reset allows re-initialization (useful for testing)
func (ts *ThreadSafeSingleton[T]) Reset() {
	ts.instance.Store(nil)
	ts.once = sync.Once{}
}

// SetForTesting allows setting a mock instance for testing
func (ts *ThreadSafeSingleton[T]) SetForTesting(instance T) {
	ts.instance.Store(instance)
}
