package notatask

import (
	"NotaborEngine/notatomic"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var ErrDone = errors.New("task finished")

type job struct {
	task *Task
	wg   *sync.WaitGroup
	err  *error
}

var jobPool = sync.Pool{
	New: func() any { return &job{} },
}

type Loop struct {
	Hz        notatomic.Float32
	lastTick  notatomic.Pointer[time.Time]
	delta     notatomic.Int64
	tickCount notatomic.UInt64

	Tasks notatomic.Pointer[[]*Task] // double-buffered slice of tasks

	stop       chan struct{}
	workerDone chan struct{}
	taskQueue  chan *job

	wg         sync.WaitGroup
	workerWg   sync.WaitGroup
	numWorkers int

	// Reusable slice buffer to avoid allocating each tick
	taskBuffer []*Task
}

// Start initializes the loop and workers
func (l *Loop) Start() {
	if l.stop != nil {
		return
	}
	l.stop = make(chan struct{})

	if l.Tasks.Get() == nil {
		tasks := make([]*Task, 0, 32)
		l.Tasks.Set(&tasks)
	}

	if l.lastTick.Get() == nil {
		now := time.Now()
		l.lastTick.Set(&now)
	}

	l.numWorkers = runtime.NumCPU() - 1
	if l.numWorkers < 1 {
		l.numWorkers = 1
	}

	l.startWorkers()
	l.wg.Add(1)
	go l.runLoop()
}

func (l *Loop) startWorkers() {
	l.taskQueue = make(chan *job, 256) // large buffer to reduce blocking
	l.workerDone = make(chan struct{})

	for i := 0; i < l.numWorkers; i++ {
		l.workerWg.Add(1)
		go l.worker()
	}
}

func (l *Loop) worker() {
	defer l.workerWg.Done()
	for {
		select {
		case j := <-l.taskQueue:
			if j == nil || j.task == nil {
				continue
			}
			*j.err = j.task.runFunc(time.Now())
			j.wg.Done()
		case <-l.workerDone:
			return
		}
	}
}

// Stop halts the loop and all workers
func (l *Loop) Stop() {
	if l.stop == nil {
		return
	}
	select {
	case <-l.stop:
		return
	default:
		close(l.stop)
	}
	l.wg.Wait()
	close(l.workerDone)
	l.workerWg.Wait()
}

// Add a task to the loop
func (l *Loop) Add(t *Task) {
	const minCap = 32
	for {
		oldPtr := l.Tasks.Get()
		old := [](*Task){}
		if oldPtr != nil {
			old = *oldPtr
		}

		newLen := len(old) + 1
		newCap := newLen * 2
		if newLen < minCap {
			newCap = minCap
		}

		newSlice := make([]*Task, newLen, newCap)
		copy(newSlice, old)
		newSlice[len(old)] = t

		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// Remove a task from the loop (pointer comparison, no reflection)
func (l *Loop) Remove(t *Task) {
	for {
		oldPtr := l.Tasks.Get()
		oldSlice := *oldPtr
		found := -1
		for i, existing := range oldSlice {
			if existing == t { // direct pointer comparison
				found = i
				break
			}
		}
		if found == -1 {
			return
		}

		newSlice := append(oldSlice[:found:found], oldSlice[found+1:]...)
		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// runLoop executes tasks per tick
func (l *Loop) runLoop() {
	defer l.wg.Done()
	hz := l.Hz.Get()
	if hz <= 0 {
		hz = 60
	}
	interval := time.Duration(float64(time.Second) / float64(hz))
	l.delta.Set(int64(interval))
	nextTick := time.Now().Add(interval)

	for {
		select {
		case <-l.stop:
			return
		default:
		}

		l.waitUntil(nextTick)
		now := time.Now()
		l.lastTick.Set(&now)

		ptrTasks := l.Tasks.Get()
		if ptrTasks != nil {
			tasks := *ptrTasks
			if len(l.taskBuffer) < len(tasks) {
				l.taskBuffer = make([]*Task, len(tasks))
			} else {
				l.taskBuffer = l.taskBuffer[:len(tasks)]
			}
			copy(l.taskBuffer, tasks)

			// Clear the shared pointer quickly
			empty := (*[]*Task)(nil)
			l.Tasks.Set(empty)

			if len(l.taskBuffer) > 0 {
				remaining := l.runTasksParallel(l.taskBuffer)
				l.mergeBack(remaining)
			}
		}

		l.tickCount.Add(1)
		newHz := l.Hz.Get()
		if newHz > 0 && newHz != hz {
			hz = newHz
			interval = time.Duration(float64(time.Second) / float64(hz))
			l.delta.Set(int64(interval))
		}
		nextTick = nextTick.Add(interval)
	}
}

// mergeBack efficiently reuses slices
func (l *Loop) mergeBack(tasks []*Task) {
	if len(tasks) == 0 {
		return
	}
	for {
		oldPtr := l.Tasks.Get()
		var old []*Task
		if oldPtr != nil {
			old = *oldPtr
		}
		newSlice := make([]*Task, len(old)+len(tasks))
		copy(newSlice, old)
		copy(newSlice[len(old):], tasks)
		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// runTasksParallel executes tasks via workers and returns remaining
func (l *Loop) runTasksParallel(tasks []*Task) []*Task {
	count := len(tasks)
	if count == 0 {
		return tasks[:0]
	}

	results := make([]error, count)
	var tickWg sync.WaitGroup
	tickWg.Add(count)

	for i, t := range tasks {
		j := jobPool.Get().(*job)
		j.task = t
		j.err = &results[i]
		j.wg = &tickWg
		l.taskQueue <- j
	}

	tickWg.Wait()

	for i, _ := range tasks {
		if results[i] != nil && !errors.Is(results[i], ErrDone) {
			fmt.Println("Task error:", results[i])
		}
	}

	// Keep only active tasks
	newIdx := 0
	for i := 0; i < count; i++ {
		if results[i] == nil {
			tasks[newIdx] = tasks[i]
			newIdx++
		}
	}
	tasks = tasks[:newIdx]

	// Reset jobs
	for i := 0; i < count; i++ {
		jobPool.Put(&job{})
	}

	return tasks
}

// waitUntil sleeps until next tick
func (l *Loop) waitUntil(nextTick time.Time) {
	if remaining := time.Until(nextTick); remaining > 0 {
		time.Sleep(remaining)
	}
}

// CreateLoop returns a new Loop instance
func CreateLoop(Hz float32) *Loop {
	l := &Loop{}
	l.Hz.Set(Hz)
	l.Tasks.Set(&[]*Task{})
	now := time.Now()
	l.lastTick.Set(&now)
	return l
}

// Alpha returns progress through current tick
func (l *Loop) Alpha(now time.Time) float32 {
	last := l.lastTick.Get()
	delta := time.Duration(l.delta.Get())

	if last == nil || delta <= 0 {
		return 1
	}

	alpha := float32(now.Sub(*last).Seconds() / delta.Seconds())
	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}
	return alpha
}
