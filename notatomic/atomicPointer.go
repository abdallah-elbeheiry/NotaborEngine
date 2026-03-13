package notatomic

import "sync/atomic"

// Pointer is an atomic wrapper around *T.
type Pointer[T any] struct {
	v atomic.Pointer[T]
}

// Get returns the current pointer value.
func (p *Pointer[T]) Get() *T {
	return p.v.Load()
}

// Set sets the pointer value.
func (p *Pointer[T]) Set(val *T) {
	p.v.Store(val)
}

// GetAndSet atomically sets the pointer to new and returns the old pointer.
func (p *Pointer[T]) GetAndSet(new *T) *T {
	return p.v.Swap(new)
}

// CompareAndSwap swaps the pointer if it matches old.
func (p *Pointer[T]) CompareAndSwap(old, new *T) bool {
	return p.v.CompareAndSwap(old, new)
}

// IsNil returns true if the pointer is nil.
func (p *Pointer[T]) IsNil() bool {
	return p.v.Load() == nil
}
