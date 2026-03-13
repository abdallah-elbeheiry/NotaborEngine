package notatomic

import (
	"strconv"
	"sync/atomic"
)

// Int32 is an atomic int32 wrapper.
type Int32 struct {
	v atomic.Int32
}

func (a *Int32) Add(delta int32) int32 { return a.v.Add(delta) }
func (a *Int32) Sub(delta int32) int32 { return a.v.Add(-delta) }
func (a *Int32) Inc() int32            { return a.v.Add(1) }
func (a *Int32) Dec() int32            { return a.v.Add(-1) }

// CompareAndSwap replaces old with new if old equals the current value.
// Returns whether the swap happened.
func (a *Int32) CompareAndSwap(old, new int32) bool { return a.v.CompareAndSwap(old, new) }

func (a *Int32) Reset()                  { a.v.Store(0) }
func (a *Int32) IsZero() bool            { return a.v.Load() == 0 }
func (a *Int32) Get() int32              { return a.v.Load() }
func (a *Int32) Set(v int32)             { a.v.Store(v) }
func (a *Int32) GetAndSet(v int32) int32 { return a.v.Swap(v) }

// TryAdd adds delta only if the result does not exceed limit.
func (a *Int32) TryAdd(delta, limit int32) bool {
	for {
		old := a.Get()
		res := old + delta
		if res > limit {
			return false
		}
		if a.CompareAndSwap(old, res) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *Int32) SetIfGreater(val int32) {
	for {
		old := a.Get()
		if val <= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *Int32) SetIfLess(val int32) {
	for {
		old := a.Get()
		if val >= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *Int32) SetIfEqual(val int32) {
	for {
		old := a.Get()
		if val != old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

func (a *Int32) String() string {
	return strconv.FormatInt(int64(a.Get()), 10)
}

// Int64 is an atomic int64 wrapper.
type Int64 struct {
	v atomic.Int64
}

func (a *Int64) Add(delta int64) int64 { return a.v.Add(delta) }
func (a *Int64) Sub(delta int64) int64 { return a.v.Add(-delta) }
func (a *Int64) Inc() int64            { return a.v.Add(1) }
func (a *Int64) Dec() int64            { return a.v.Add(-1) }

func (a *Int64) CompareAndSwap(old, new int64) bool { return a.v.CompareAndSwap(old, new) }

func (a *Int64) Reset()                  { a.v.Store(0) }
func (a *Int64) IsZero() bool            { return a.v.Load() == 0 }
func (a *Int64) Get() int64              { return a.v.Load() }
func (a *Int64) Set(v int64)             { a.v.Store(v) }
func (a *Int64) GetAndSet(v int64) int64 { return a.v.Swap(v) }

// TryAdd adds delta only if the result does not exceed limit.
func (a *Int64) TryAdd(delta, limit int64) bool {
	for {
		old := a.Get()
		res := old + delta
		if res > limit {
			return false
		}
		if a.CompareAndSwap(old, res) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *Int64) SetIfGreater(val int64) {
	for {
		old := a.Get()
		if val <= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *Int64) SetIfLess(val int64) {
	for {
		old := a.Get()
		if val >= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *Int64) SetIfEqual(val int64) {
	for {
		old := a.Get()
		if val != old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

func (a *Int64) String() string {
	return strconv.FormatInt(a.Get(), 10)
}

// UInt32 is an atomic uint32 wrapper.
type UInt32 struct {
	v atomic.Uint32
}

func (a *UInt32) Add(delta uint32) uint32 { return a.v.Add(delta) }
func (a *UInt32) Sub(delta uint32) uint32 { return a.v.Add(^uint32(0) - delta) }
func (a *UInt32) Inc() uint32             { return a.v.Add(1) }

// Dec decrements the value. Wraps around on underflow (unsigned semantics).
func (a *UInt32) Dec() uint32 { return a.v.Add(^uint32(0)) }

func (a *UInt32) CompareAndSwap(old, new uint32) bool { return a.v.CompareAndSwap(old, new) }

func (a *UInt32) Reset()                    { a.v.Store(0) }
func (a *UInt32) IsZero() bool              { return a.v.Load() == 0 }
func (a *UInt32) Get() uint32               { return a.v.Load() }
func (a *UInt32) Set(v uint32)              { a.v.Store(v) }
func (a *UInt32) GetAndSet(v uint32) uint32 { return a.v.Swap(v) }

// TryAdd adds delta only if the result would not exceed limit.
func (a *UInt32) TryAdd(delta, limit uint32) bool {
	for {
		old := a.Get()
		res := old + delta
		if res > limit {
			return false
		}
		if a.CompareAndSwap(old, res) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *UInt32) SetIfGreater(val uint32) {
	for {
		old := a.Get()
		if val <= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *UInt32) SetIfLess(val uint32) {
	for {
		old := a.Get()
		if val >= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *UInt32) SetIfEqual(val uint32) {
	for {
		old := a.Get()
		if val != old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// Or atomically applies a bitwise OR with mask and returns the new value.
func (a *UInt32) Or(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old | mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// And atomically applies a bitwise AND with mask and returns the new value.
func (a *UInt32) And(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old & mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Clear atomically clears the bits specified by mask and returns the new value.
func (a *UInt32) Clear(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old &^ mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Toggle atomically toggles the bits specified by mask and returns the new value.
func (a *UInt32) Toggle(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old ^ mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

func (a *UInt32) String() string {
	return strconv.FormatUint(uint64(a.Get()), 10)
}

// UInt64 is an atomic uint64 wrapper.
type UInt64 struct {
	v atomic.Uint64
}

func (a *UInt64) Add(delta uint64) uint64 { return a.v.Add(delta) }
func (a *UInt64) Sub(delta uint64) uint64 { return a.v.Add(^uint64(0) - delta) }
func (a *UInt64) Inc() uint64             { return a.v.Add(1) }

// Dec decrements the value. Wraps around on underflow (unsigned semantics).
func (a *UInt64) Dec() uint64 { return a.v.Add(^uint64(0)) }

func (a *UInt64) CompareAndSwap(old, new uint64) bool { return a.v.CompareAndSwap(old, new) }

func (a *UInt64) Reset()                    { a.v.Store(0) }
func (a *UInt64) IsZero() bool              { return a.v.Load() == 0 }
func (a *UInt64) Get() uint64               { return a.v.Load() }
func (a *UInt64) Set(v uint64)              { a.v.Store(v) }
func (a *UInt64) GetAndSet(v uint64) uint64 { return a.v.Swap(v) }

// TryAdd adds delta only if the result does not exceed limit.
func (a *UInt64) TryAdd(delta, limit uint64) bool {
	for {
		old := a.Get()
		res := old + delta
		if res > limit {
			return false
		}
		if a.CompareAndSwap(old, res) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *UInt64) SetIfGreater(val uint64) {
	for {
		old := a.Get()
		if val <= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *UInt64) SetIfLess(val uint64) {
	for {
		old := a.Get()
		if val >= old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *UInt64) SetIfEqual(val uint64) {
	for {
		old := a.Get()
		if val != old {
			return
		}
		if a.CompareAndSwap(old, val) {
			return
		}
	}
}

// Or atomically applies a bitwise OR with mask and returns the new value.
func (a *UInt64) Or(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old | mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// And atomically applies a bitwise AND with mask and returns the new value.
func (a *UInt64) And(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old & mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Clear atomically clears the bits specified by mask and returns the new value.
func (a *UInt64) Clear(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old &^ mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Toggle atomically toggles the bits specified by mask and returns the new value.
func (a *UInt64) Toggle(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old ^ mask
		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

func (a *UInt64) String() string {
	return strconv.FormatUint(uint64(a.Get()), 10)
}
