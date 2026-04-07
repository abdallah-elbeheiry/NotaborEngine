package notatask

import (
	"time"
)

type Task struct {
	fn      func() error              // 8 bytes
	cond    func() bool               // 8 bytes
	runFunc func(now time.Time) error // 8 bytes

	// int64 is used for UnixNano
	lastRun  int64 // 8 bytes
	finishAt int64 // 8 bytes

	interval  int64 // 8 bytes
	timeDelay int64 // 8 bytes

	tickDelay uint32 // 4 bytes
	tickCount uint32 // 4 bytes
	maxRuns   uint32 // 4 bytes

	// total: 68 bytes (rounds to 72 bytes due to fields)
}

// TaskOption configures a Task at creation time.
type TaskOption func(*Task)

// CreateTask creates a task with a dynamically generated runFunc.
// All configuration is done through TaskOption functions.
func CreateTask(fn func() error, opts ...TaskOption) *Task {
	t := &Task{
		fn:      fn,
		lastRun: time.Now().UnixNano(),
	}

	// Apply configuration options
	for _, opt := range opts {
		opt(t)
	}

	// Keep track if initial delay is consumed
	var delayConsumed bool

	t.runFunc = func(now time.Time) error {
		nowUnix := now.UnixNano()

		// 1️FinishAfter: hard stop
		if t.finishAt != 0 && nowUnix >= t.finishAt {
			return ErrDone
		}

		// Initial delay: only skip once
		if !delayConsumed && t.timeDelay != 0 {
			if nowUnix-t.lastRun < t.timeDelay {
				return nil
			}
			delayConsumed = true
			t.lastRun = nowUnix // start counting intervals after the delay
		}

		// tick delay
		if t.tickDelay > 0 {
			t.tickDelay--
			return nil
		}

		// Interval check (regular periodic execution)
		if t.interval != 0 && nowUnix-t.lastRun < t.interval {
			return nil
		}

		// Execute task
		if err := t.fn(); err != nil {
			return err
		}

		t.lastRun = nowUnix
		t.tickCount++

		// Max runs
		if t.maxRuns != 0 && t.tickCount >= t.maxRuns {
			return ErrDone
		}

		// Conditional stop
		if t.cond != nil && t.cond() {
			return ErrDone
		}

		return nil
	}

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
