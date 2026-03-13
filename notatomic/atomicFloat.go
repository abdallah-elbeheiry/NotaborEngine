package notatomic

import (
	"math"
	"strconv"
	"sync/atomic"
)

// Float32 is an atomic float32 wrapper.
type Float32 struct {
	v atomic.Uint32
}

func (a *Float32) Get() float32    { return math.Float32frombits(a.v.Load()) }
func (a *Float32) Set(val float32) { a.v.Store(math.Float32bits(val)) }
func (a *Float32) Reset()          { a.v.Store(0) }
func (a *Float32) IsZero() bool    { return a.Get() == 0 }

func (a *Float32) GetAndSet(val float32) float32 { return a.Swap(val) }

func (a *Float32) Swap(val float32) float32 {
	return math.Float32frombits(a.v.Swap(math.Float32bits(val)))
}

// CompareAndSwap replaces old with new if old equals the current value.
// Returns whether the swap happened.
func (a *Float32) CompareAndSwap(old, new float32) bool {
	return a.v.CompareAndSwap(math.Float32bits(old), math.Float32bits(new))
}

func (a *Float32) Add(delta float32) float32 {
	for {
		oldBits := a.v.Load()
		oldVal := math.Float32frombits(oldBits)
		newVal := oldVal + delta
		if a.v.CompareAndSwap(oldBits, math.Float32bits(newVal)) {
			return newVal
		}
	}
}

func (a *Float32) Sub(delta float32) float32 { return a.Add(-delta) }
func (a *Float32) Inc() float32              { return a.Add(1) }
func (a *Float32) Dec() float32              { return a.Add(-1) }

// TryAdd adds delta only if the result would not exceed limit.
func (a *Float32) TryAdd(delta, limit float32) bool {
	for {
		oldBits := a.v.Load()
		old := math.Float32frombits(oldBits)
		newVal := old + delta
		if newVal > limit {
			return false
		}
		if a.v.CompareAndSwap(oldBits, math.Float32bits(newVal)) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *Float32) SetIfGreater(val float32) {
	for {
		oldBits := a.v.Load()
		old := math.Float32frombits(oldBits)
		if val <= old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float32bits(val)) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *Float32) SetIfLess(val float32) {
	for {
		oldBits := a.v.Load()
		old := math.Float32frombits(oldBits)
		if val >= old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float32bits(val)) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *Float32) SetIfEqual(val float32) {
	for {
		oldBits := a.v.Load()
		old := math.Float32frombits(oldBits)
		if val != old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float32bits(val)) {
			return
		}
	}
}

func (a *Float32) String() string {
	return strconv.FormatFloat(float64(a.Get()), 'f', -1, 32)
}

// Float64 is an atomic float64 wrapper.
type Float64 struct {
	v atomic.Uint64
}

func (a *Float64) Get() float64    { return math.Float64frombits(a.v.Load()) }
func (a *Float64) Set(val float64) { a.v.Store(math.Float64bits(val)) }
func (a *Float64) Reset()          { a.v.Store(0) }
func (a *Float64) IsZero() bool    { return a.Get() == 0 }

func (a *Float64) GetAndSet(val float64) float64 { return a.Swap(val) }

func (a *Float64) Swap(val float64) float64 {
	return math.Float64frombits(a.v.Swap(math.Float64bits(val)))
}

// CompareAndSwap replaces old with new if old equals the current value.
// Returns whether the swap happened.
func (a *Float64) CompareAndSwap(old, new float64) bool {
	return a.v.CompareAndSwap(math.Float64bits(old), math.Float64bits(new))
}

func (a *Float64) Add(delta float64) float64 {
	for {
		oldBits := a.v.Load()
		oldVal := math.Float64frombits(oldBits)
		newVal := oldVal + delta
		if a.v.CompareAndSwap(oldBits, math.Float64bits(newVal)) {
			return newVal
		}
	}
}

func (a *Float64) Sub(delta float64) float64 { return a.Add(-delta) }
func (a *Float64) Inc() float64              { return a.Add(1) }
func (a *Float64) Dec() float64              { return a.Add(-1) }

// TryAdd adds delta only if the result would not exceed limit.
func (a *Float64) TryAdd(delta, limit float64) bool {
	for {
		oldBits := a.v.Load()
		old := math.Float64frombits(oldBits)
		newVal := old + delta
		if newVal > limit {
			return false
		}
		if a.v.CompareAndSwap(oldBits, math.Float64bits(newVal)) {
			return true
		}
	}
}

// SetIfGreater updates the value only if val is greater than the current value.
func (a *Float64) SetIfGreater(val float64) {
	for {
		oldBits := a.v.Load()
		old := math.Float64frombits(oldBits)
		if val <= old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float64bits(val)) {
			return
		}
	}
}

// SetIfLess updates the value only if val is less than the current value.
func (a *Float64) SetIfLess(val float64) {
	for {
		oldBits := a.v.Load()
		old := math.Float64frombits(oldBits)
		if val >= old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float64bits(val)) {
			return
		}
	}
}

// SetIfEqual updates the value only if val equals the current value.
func (a *Float64) SetIfEqual(val float64) {
	for {
		oldBits := a.v.Load()
		old := math.Float64frombits(oldBits)
		if val != old {
			return
		}
		if a.v.CompareAndSwap(oldBits, math.Float64bits(val)) {
			return
		}
	}
}

func (a *Float64) String() string {
	return strconv.FormatFloat(a.Get(), 'f', -1, 64)
}
