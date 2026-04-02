package notatask

import (
	"NotaborEngine/notatomic"
	"errors"
	"fmt"
	"reflect"
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
	New: func() any {
		return &job{}
	},
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
}

// Start initializes the loop and workers
func (l *Loop) Start() {
	if l.stop != nil {
		return
	}
	l.stop = make(chan struct{})

	if l.Tasks.Get() == nil {
		tasks := make([]*Task, 0)
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
	l.taskQueue = make(chan *job, l.numWorkers*4)
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
		var old []*Task
		if oldPtr != nil {
			old = *oldPtr
		}

		newLen := len(old) + 1
		var newCap int
		if len(old) < minCap {
			newCap = minCap
		} else {
			newCap = len(old) * 2
		}

		newSlice := make([]*Task, newLen, newCap)
		copy(newSlice, old)
		newSlice[len(old)] = t

		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// Remove a task from the loop
func (l *Loop) Remove(t *Task) {
	targetPtr := reflect.ValueOf(t).Pointer()
	for {
		oldPtr := l.Tasks.Get()
		oldSlice := *oldPtr

		foundIdx := -1
		for i, existing := range oldSlice {
			if reflect.ValueOf(existing).Pointer() == targetPtr {
				foundIdx = i
				break
			}
		}

		if foundIdx == -1 {
			return
		}

		newSlice := append([]*Task{}, oldSlice[:foundIdx]...)
		newSlice = append(newSlice, oldSlice[foundIdx+1:]...)

		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// runLoop executes tasks each tick
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
		l.lastTick.Set(&nextTick)

		ptrTasks := l.Tasks.Get()
		if ptrTasks != nil {
			tasksToRun := *ptrTasks
			empty := make([]*Task, 0)
			l.Tasks.Set(&empty)

			if len(tasksToRun) > 0 {
				remaining := l.runTasksParallel(tasksToRun)
				l.mergeBack(remaining)
			}
		}

		l.tickCount.Add(1)

		// Update interval in case Hz changed
		newHz := l.Hz.Get()
		if newHz != hz && newHz > 0 {
			hz = newHz
			interval = time.Duration(float64(time.Second) / float64(hz))
			l.delta.Set(int64(interval))
		}

		nextTick = nextTick.Add(interval)
	}
}

// mergeBack merges remaining tasks into the loop buffer
func (l *Loop) mergeBack(tasks []*Task) {
	if len(tasks) == 0 {
		return
	}

	const minCap = 32
	for {
		oldPtr := l.Tasks.Get()
		var old []*Task
		if oldPtr != nil {
			old = *oldPtr
		}

		newLen := len(old) + len(tasks)
		var newCap int
		if newLen < minCap {
			newCap = minCap
		} else {
			newCap = newLen * 2
		}

		newSlice := make([]*Task, newLen, newCap)
		copy(newSlice, old)
		copy(newSlice[len(old):], tasks)

		if l.Tasks.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

// runTasksParallel executes tasks via workers and returns active tasks
func (l *Loop) runTasksParallel(tasks []*Task) []*Task {
	count := len(tasks)
	if count == 0 {
		return tasks[:0]
	}

	results := make([]error, count)
	var tickWg sync.WaitGroup
	jobsUsed := make([]*job, 0, count)

	for i, t := range tasks {
		tickWg.Add(1)
		j := jobPool.Get().(*job)
		j.task = t
		j.err = &results[i]
		j.wg = &tickWg
		jobsUsed = append(jobsUsed, j)
		l.taskQueue <- j
	}

	tickWg.Wait()

	for _, j := range jobsUsed {
		j.task = nil
		j.err = nil
		j.wg = nil
		jobPool.Put(j)
	}

	newIdx := 0
	for i := 0; i < count; i++ {
		err := results[i]
		if err == nil {
			tasks[newIdx] = tasks[i]
			newIdx++
			continue
		}
		if errors.Is(err, ErrDone) {
			continue // remove silently
		}
		fmt.Println("Task error:", err)
	}

	for i := newIdx; i < count; i++ {
		tasks[i] = nil
	}

	return tasks[:newIdx]
}

// waitUntil sleeps until the given tick
func (l *Loop) waitUntil(nextTick time.Time) {
	remaining := time.Until(nextTick)
	if remaining > 0 {
		time.Sleep(remaining)
	}
}

// CreateLoop returns a new Loop instance
func CreateLoop(Hz float32) *Loop {
	loop := &Loop{}
	loop.Hz.Set(Hz)
	tasks := make([]*Task, 0)
	loop.Tasks.Set(&tasks)
	now := time.Now()
	loop.lastTick.Set(&now)
	return loop
}

// Alpha returns a value in [0,1] representing how far we are through the current tick
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
