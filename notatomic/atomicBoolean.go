package notatomic

import "sync/atomic"

// Bool is an atomic boolean wrapper.
type Bool struct {
	v atomic.Uint32
}

// Get returns the current value as a Go bool.
func (b *Bool) Get() bool {
	return b.v.Load() != 0
}

// Set sets the value.
func (b *Bool) Set(val bool) {
	var v uint32
	if val {
		v = 1
	}
	b.v.Store(v)
}

// GetAndSet sets the value and returns the previous value.
func (b *Bool) GetAndSet(val bool) bool {
	var v uint32
	if val {
		v = 1
	}
	old := b.v.Swap(v)
	return old != 0
}

// CompareAndSwap swaps the value if it matches the old value.
// Returns true if the value was set.
func (b *Bool) CompareAndSwap(old, new bool) bool {
	var oldV, newV uint32
	if old {
		oldV = 1
	}
	if new {
		newV = 1
	}
	return b.v.CompareAndSwap(oldV, newV)
}

// SetIfTrue sets the value only if it is currently true.
// Returns true if the value was set.
func (b *Bool) SetIfTrue(val bool) bool {
	return b.CompareAndSwap(true, val)
}

// SetIfFalse sets the value only if it is currently false.
// Returns true if the value was set.
func (b *Bool) SetIfFalse(val bool) bool {
	return b.CompareAndSwap(false, val)
}

// Reset sets the value to false.
func (b *Bool) Reset() {
	b.v.Store(0)
}

// IsZero returns true if the value is false.
func (b *Bool) IsZero() bool {
	return b.v.Load() == 0
}
