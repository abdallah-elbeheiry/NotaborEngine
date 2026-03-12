package notatomic

import "sync/atomic"

type Int32 struct {
	v atomic.Int32
}

func (a *Int32) Add(delta int32) int32 {
	return a.v.Add(delta)
}

func (a *Int32) Inc() int32 {
	return a.v.Add(1)
}

func (a *Int32) Dec() int32 {
	return a.v.Add(-1)
}

func (a *Int32) Swap(val int32) int32 {
	return a.v.Swap(val)
}

// CompareAndSwap replace old with new if old is equal to the atomic value, returns whether the swap happened or not
func (a *Int32) CompareAndSwap(old, new int32) bool {
	return a.v.CompareAndSwap(old, new)
}

func (a *Int32) Reset() {
	a.v.Store(0)
}

func (a *Int32) IsZero() bool {
	return a.v.Load() == 0
}

// TryAdd tries to add delta but ensures the result does not exceed limit
func (a *Int32) TryAdd(delta int32, limit int32) bool {
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

// Int64

type Int64 struct {
	v atomic.Int64
}

func (a *Int64) Add(delta int64) int64 {
	return a.v.Add(delta)
}

func (a *Int64) Inc() int64 {
	return a.v.Add(1)
}

func (a *Int64) Dec() int64 {
	return a.v.Add(-1)
}

func (a *Int64) Swap(val int64) int64 {
	return a.v.Swap(val)
}

func (a *Int64) CompareAndSwap(old, new int64) bool {
	return a.v.CompareAndSwap(old, new)
}

func (a *Int64) Reset() {
	a.v.Store(0)
}

func (a *Int64) IsZero() bool {
	return a.v.Load() == 0
}

func (a *Int64) TryAdd(delta int64, limit int64) bool {
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

// SetIfGreater updates the value only if val is greater than current
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

// SetIfLess updates the value only if val is smaller than current
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

// UInt32

type UInt32 struct {
	v atomic.Uint32
}

func (a *UInt32) Add(delta uint32) uint32 {
	return a.v.Add(delta)
}

func (a *UInt32) Inc() uint32 {
	return a.v.Add(1)
}

func (a *UInt32) Dec() uint32 {
	return a.v.Add(^uint32(0))
}

func (a *UInt32) Swap(val uint32) uint32 {
	return a.v.Swap(val)
}

func (a *UInt32) CompareAndSwap(old, new uint32) bool {
	return a.v.CompareAndSwap(old, new)
}

func (a *UInt32) Reset() {
	a.v.Store(0)
}

func (a *UInt32) IsZero() bool {
	return a.v.Load() == 0
}

func (a *UInt32) TryAdd(delta uint32, limit uint32) bool {
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

// Or provides Bitwise OR
func (a *UInt32) Or(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old | mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// And provides Bitwise AND
func (a *UInt32) And(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old & mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Clear bits
func (a *UInt32) Clear(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old &^ mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// Toggle bits
func (a *UInt32) Toggle(mask uint32) uint32 {
	for {
		old := a.Get()
		res := old ^ mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// UInt64

type UInt64 struct {
	v atomic.Uint64
}

func (a *UInt64) Add(delta uint64) uint64 {
	return a.v.Add(delta)
}

func (a *UInt64) Inc() uint64 {
	return a.v.Add(1)
}

func (a *UInt64) Dec() uint64 {
	return a.v.Add(^uint64(0))
}

func (a *UInt64) Swap(val uint64) uint64 {
	return a.v.Swap(val)
}

func (a *UInt64) CompareAndSwap(old, new uint64) bool {
	return a.v.CompareAndSwap(old, new)
}

func (a *UInt64) Reset() {
	a.v.Store(0)
}

func (a *UInt64) IsZero() bool {
	return a.v.Load() == 0
}

func (a *UInt64) TryAdd(delta uint64, limit uint64) bool {
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

func (a *UInt64) Or(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old | mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

func (a *UInt64) And(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old & mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

func (a *UInt64) Clear(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old &^ mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

func (a *UInt64) Toggle(mask uint64) uint64 {
	for {
		old := a.Get()
		res := old ^ mask

		if a.CompareAndSwap(old, res) {
			return res
		}
	}
}

// getters and setters

func (a *Int32) Get() int32  { return a.v.Load() }
func (a *Int32) Set(v int32) { a.v.Store(v) }

func (a *Int64) Get() int64  { return a.v.Load() }
func (a *Int64) Set(v int64) { a.v.Store(v) }

func (a *UInt32) Get() uint32  { return a.v.Load() }
func (a *UInt32) Set(v uint32) { a.v.Store(v) }

func (a *UInt64) Get() uint64  { return a.v.Load() }
func (a *UInt64) Set(v uint64) { a.v.Store(v) }
