package notatask

import (
	"sync/atomic"
	"time"
)

type Task struct {
	fn   func()
	cond func() bool

	// int64 for UnixNano
	lastRun  int64
	finishAt int64

	interval  int64
	timeDelay int64

	tickDelay uint32
	tickCount uint32
	maxRuns   uint32

	canceled      atomic.Uint32
	delayConsumed bool
	done          bool
	err           error // set by run when task is done or errored
}

// TaskOption configures a Task at creation time.
type TaskOption func(*Task)

// CreateTask creates a task.
// All configuration is done through TaskOption functions.
func CreateTask(fn func(), opts ...TaskOption) *Task {
	t := &Task{
		fn:      fn,
		lastRun: time.Now().UnixNano(),
	}

	// Apply configuration options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func Do(fn func()) *Task {
	return CreateTask(fn)
}

func Once(fn func()) *Task {
	return CreateTask(fn, RunOnce())
}

func Every(d time.Duration, fn func()) *Task {
	return CreateTask(fn, RepeatEvery(d))
}

func Times(n uint32, fn func()) *Task {
	return CreateTask(fn, RepeatTimes(n))
}

func After(d time.Duration, fn func()) *Task {
	return CreateTask(fn, WithDelay(d))
}

func AfterTicks(count uint32, fn func()) *Task {
	return CreateTask(fn, StartAfterTicks(count))
}

func (t *Task) run(nowUnix int64) {
	if t.canceled.Load() != 0 {
		t.done = true
		t.err = ErrDone
		return
	}

	if t.finishAt != 0 && nowUnix >= t.finishAt {
		t.done = true
		t.err = ErrDone
		return
	}

	if !t.delayConsumed && t.timeDelay != 0 {
		if nowUnix-t.lastRun < t.timeDelay {
			return
		}
		t.delayConsumed = true
		t.lastRun = nowUnix
	}

	if t.tickDelay > 0 {
		t.tickDelay--
		return
	}

	if t.interval != 0 && nowUnix-t.lastRun < t.interval {
		return
	}

	t.fn()
	t.lastRun = nowUnix
	t.tickCount++

	if t.maxRuns != 0 && t.tickCount >= t.maxRuns {
		t.done = true
		t.err = ErrDone
		return
	}

	if t.cond != nil && t.cond() {
		t.done = true
		t.err = ErrDone
	}
}

func (t *Task) cancel() {
	t.canceled.Store(1)
}

func (t *Task) Every(d time.Duration) *Task {
	t.interval = int64(d)
	return t
}

func (t *Task) Delay(d time.Duration) *Task {
	t.timeDelay = int64(d)
	return t
}

func (t *Task) AfterTicks(count uint32) *Task {
	t.tickDelay = count
	return t
}

func (t *Task) Times(n uint32) *Task {
	t.maxRuns = n
	return t
}

func (t *Task) Once() *Task {
	t.maxRuns = 1
	return t
}

func (t *Task) Until(cond func() bool) *Task {
	t.cond = cond
	return t
}

func (t *Task) FinishAfter(d time.Duration) *Task {
	t.finishAt = time.Now().Add(d).UnixNano()
	return t
}

// RunOnce executes the task only once.
func RunOnce() TaskOption {
	return func(t *Task) {
		t.maxRuns = 1
	}
}

// WithDelay sets an initial delay before the first execution.
func WithDelay(d time.Duration) TaskOption {
	return func(t *Task) {
		t.timeDelay = int64(d)
	}
}

// RepeatEvery sets a periodic interval between executions.
func RepeatEvery(d time.Duration) TaskOption {
	return func(t *Task) {
		t.interval = int64(d)
	}
}

// StartAfterTicks delays execution by a number of ticks.
func StartAfterTicks(count uint32) TaskOption {
	return func(t *Task) {
		t.tickDelay = count
	}
}

// RepeatTimes sets the task to execute a fixed number of times.
func RepeatTimes(n uint32) TaskOption {
	return func(t *Task) {
		t.maxRuns = n
	}
}

// StopWhen sets a conditional function to end the task.
func StopWhen(cond func() bool) TaskOption {
	return func(t *Task) {
		t.cond = cond
	}
}

// FinishAfter sets a hard finish time.
func FinishAfter(d time.Duration) TaskOption {
	return func(t *Task) {
		t.finishAt = time.Now().Add(d).UnixNano()
	}
}
