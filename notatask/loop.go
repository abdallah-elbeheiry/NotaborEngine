package notatask

import (
	"NotaborEngine/notatomic"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var ErrDone = errors.New("task finished")

type worker struct {
	jobs chan workerJob
}

type workerJob struct {
	tasks   []*Task
	nowUnix int64
	done    *sync.WaitGroup
}

type Loop struct {
	Hz           notatomic.Float32
	lastTickNano notatomic.Int64
	delta        notatomic.Int64
	tickCount    notatomic.UInt64

	tasks []*Task

	pendingMu    sync.Mutex
	pendingTasks []*Task
	hasPending   atomic.Uint32

	stop     chan struct{}
	workerWg sync.WaitGroup

	numWorkers        int
	parallelThreshold int
	workers           []*worker
}

func CreateLoop(Hz float32) *Loop {
	l := &Loop{}
	l.Hz.Set(Hz)
	l.tasks = make([]*Task, 0, 64)
	l.pendingTasks = make([]*Task, 0, 16)
	l.lastTickNano.Set(time.Now().UnixNano())
	return l
}

func (l *Loop) Start() {
	if l.stop != nil {
		return
	}
	l.stop = make(chan struct{})

	if l.lastTickNano.Get() == 0 {
		l.lastTickNano.Set(time.Now().UnixNano())
	}

	totalWorkers := runtime.GOMAXPROCS(0)
	if totalWorkers < 1 {
		totalWorkers = 1
	}

	l.numWorkers = totalWorkers - 1
	l.parallelThreshold = maxInt(64, totalWorkers*8)

	l.startWorkers()
	go l.runLoop()
}

func (l *Loop) startWorkers() {
	if l.numWorkers <= 0 {
		l.workers = nil
		return
	}

	l.workers = make([]*worker, l.numWorkers)
	for i := 0; i < l.numWorkers; i++ {
		l.workers[i] = &worker{jobs: make(chan workerJob, 1)}
		l.workerWg.Add(1)
		go l.workerLoop(l.workers[i])
	}
}

func (l *Loop) workerLoop(w *worker) {
	defer l.workerWg.Done()

	for {
		select {
		case <-l.stop:
			return
		case job := <-w.jobs:
			runTaskSlice(job.tasks, job.nowUnix)
			job.done.Done()
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
	l.workerWg.Wait()
}

func (l *Loop) Add(t *Task) {
	l.pendingMu.Lock()
	l.pendingTasks = append(l.pendingTasks, t)
	l.pendingMu.Unlock()
	l.hasPending.Store(1)
}

func (l *Loop) Do(fn func()) *Task {
	task := Do(fn)
	l.Add(task)
	return task
}

func (l *Loop) Once(fn func()) *Task {
	task := Once(fn)
	l.Add(task)
	return task
}

func (l *Loop) Every(d time.Duration, fn func()) *Task {
	task := Every(d, fn)
	l.Add(task)
	return task
}

func (l *Loop) Times(n uint32, fn func()) *Task {
	task := Times(n, fn)
	l.Add(task)
	return task
}

func (l *Loop) After(d time.Duration, fn func()) *Task {
	task := After(d, fn)
	l.Add(task)
	return task
}

func (l *Loop) AfterTicks(count uint32, fn func()) *Task {
	task := AfterTicks(count, fn)
	l.Add(task)
	return task
}

func (l *Loop) Remove(t *Task) {
	t.cancel()
}

func (l *Loop) runLoop() {
	hz := l.Hz.Get()
	if hz <= 0 {
		hz = 60
	}
	interval := time.Duration(float64(time.Second) / float64(hz))
	intervalNano := int64(interval)
	l.delta.Set(intervalNano)

	timer := time.NewTimer(interval)
	defer timer.Stop()

	nextTickNano := time.Now().UnixNano() + intervalNano
	const maxCatchUpTicks = 256

	for {
		select {
		case <-l.stop:
			return
		case <-timer.C:
		}

		nowUnix := time.Now().UnixNano()
		if nowUnix < nextTickNano {
			timer.Reset(time.Duration(nextTickNano - nowUnix))
			continue
		}

		catchUpTicks := 0
		for nowUnix >= nextTickNano {
			l.runTick(nowUnix)
			nextTickNano += intervalNano
			catchUpTicks++
			if catchUpTicks >= maxCatchUpTicks {
				break
			}

			newHz := l.Hz.Get()
			if newHz > 0 && newHz != hz {
				hz = newHz
				interval = time.Duration(float64(time.Second) / float64(hz))
				intervalNano = int64(interval)
				l.delta.Set(intervalNano)
			}

			nowUnix = time.Now().UnixNano()
		}

		if catchUpTicks >= maxCatchUpTicks && nowUnix >= nextTickNano {
			nextTickNano = nowUnix + intervalNano
		}

		sleepFor := nextTickNano - time.Now().UnixNano()
		if sleepFor < 0 {
			sleepFor = 0
		}
		timer.Reset(time.Duration(sleepFor))
	}
}

func (l *Loop) runTick(nowUnix int64) {
	l.lastTickNano.Set(nowUnix)

	l.drainPending()

	if len(l.tasks) > 0 {
		l.runTasks(l.tasks, nowUnix)
		l.compactTasks()
	}

	l.tickCount.Add(1)
}

func (l *Loop) runTasks(tasks []*Task, nowUnix int64) {
	if len(tasks) == 0 {
		return
	}

	if l.numWorkers == 0 || len(tasks) < l.parallelThreshold {
		runTaskSlice(tasks, nowUnix)
		return
	}

	activeLanes := minInt(len(tasks), l.numWorkers+1)
	if activeLanes <= 1 {
		runTaskSlice(tasks, nowUnix)
		return
	}

	baseChunk := len(tasks) / activeLanes
	extra := len(tasks) % activeLanes
	workerJobs := activeLanes - 1

	var done sync.WaitGroup
	if workerJobs > 0 {
		done.Add(workerJobs)
	}

	start := 0
	for lane := 0; lane < activeLanes; lane++ {
		chunkLen := baseChunk
		if lane < extra {
			chunkLen++
		}
		end := start + chunkLen
		chunk := tasks[start:end]

		if lane < workerJobs {
			l.workers[lane].jobs <- workerJob{
				tasks:   chunk,
				nowUnix: nowUnix,
				done:    &done,
			}
		} else {
			runTaskSlice(chunk, nowUnix)
		}
		start = end
	}

	if workerJobs > 0 {
		done.Wait()
	}
}

func (l *Loop) compactTasks() {
	kept := 0
	for _, t := range l.tasks {
		if !t.done && t.canceled.Load() == 0 && t.err == nil {
			l.tasks[kept] = t
			kept++
		} else if !errors.Is(t.err, ErrDone) {
			fmt.Println("Task error:", t.err)
		}
		t.done = false
		t.err = nil
	}
	l.tasks = l.tasks[:kept]
}

func (l *Loop) Alpha(now time.Time) float32 {
	lastNano := l.lastTickNano.Get()
	deltaNano := l.delta.Get()
	if lastNano == 0 || deltaNano <= 0 {
		return 1
	}

	alpha := float32(float64(now.UnixNano()-lastNano) / float64(deltaNano))
	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}
	return alpha
}

func (l *Loop) TickCount() uint64 {
	return l.tickCount.Get()
}

func runTaskSlice(tasks []*Task, nowUnix int64) {
	for _, t := range tasks {
		t.run(nowUnix)
	}
}

func (l *Loop) drainPending() {
	if l.hasPending.Load() == 0 {
		return
	}

	l.pendingMu.Lock()
	if len(l.pendingTasks) > 0 {
		l.tasks = append(l.tasks, l.pendingTasks...)
		l.pendingTasks = l.pendingTasks[:0]
	}
	l.hasPending.Store(0)
	l.pendingMu.Unlock()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
