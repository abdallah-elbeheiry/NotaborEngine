package notacore

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Runnable func() error

type FixedHzLoop struct {
	Hz float32

	mu sync.Mutex
	// double buffer for runnables
	activeRunnables  []Runnable
	nextRunnables    []Runnable
	oneTimeRunnables []Runnable

	stop chan struct{}
	wg   sync.WaitGroup

	lastTick time.Time
	delta    time.Duration

	monitorEvery time.Duration
	lastMonitor  time.Time
	tickCount    uint64

	workers    []chan job
	workerDone chan struct{}
	workerWg   sync.WaitGroup
	numWorkers int
}

type job struct {
	runnable Runnable
	result   *error
	wg       *sync.WaitGroup
}

type RenderLoop struct {
	MaxHz     float32
	Runnables []Runnable
	LastTime  time.Time
}

const (
	// Windows can't reliably sleep below 2ms
	minReliableSleep = 2 * time.Millisecond
	// Spin the last 500µs for precision
	spinWindow = 500 * time.Microsecond
)

func (l *FixedHzLoop) EnableMonitor(interval time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.monitorEvery = interval
	l.lastMonitor = time.Now()
	l.tickCount = 0
}

func (l *FixedHzLoop) Start() {
	l.mu.Lock()
	if l.stop != nil {
		l.mu.Unlock()
		return
	}
	l.stop = make(chan struct{})
	l.mu.Unlock()

	l.numWorkers = runtime.NumCPU()
	if l.numWorkers < 1 {
		l.numWorkers = 1
	}

	l.startWorkers()
	l.wg.Add(1)
	go l.runLoop()
}

func (l *FixedHzLoop) startWorkers() {
	l.workers = make([]chan job, l.numWorkers)
	l.workerDone = make(chan struct{})

	for i := 0; i < l.numWorkers; i++ {
		ch := make(chan job, 1)
		l.workers[i] = ch
		l.workerWg.Add(1)
		go l.worker(i, ch)
	}
}

func (l *FixedHzLoop) worker(id int, jobs <-chan job) {
	defer l.workerWg.Done()
	for {
		select {
		case j := <-jobs:
			*j.result = j.runnable()
			j.wg.Done()
		case <-l.workerDone:
			return
		}
	}
}

func (l *FixedHzLoop) Stop() {
	close(l.stop)
	l.wg.Wait()

	close(l.workerDone)
	l.workerWg.Wait()
}

func (l *FixedHzLoop) Add(r Runnable) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nextRunnables = append(l.nextRunnables, r)
}

func (l *FixedHzLoop) Remove(i int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	last := len(l.activeRunnables) - 1
	if i < 0 || i > last {
		return
	}

	l.activeRunnables[i] = l.activeRunnables[last]
	l.activeRunnables[last] = nil
	l.activeRunnables = l.activeRunnables[:last]
}

func (r *RenderLoop) Render() {
	now := time.Now()

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	for _, runnable := range r.Runnables {
		if err := runnable(); err != nil {
			fmt.Println("Render error:", err)
		}
	}
	r.LastTime = now
}

func (r *RenderLoop) Add(rn Runnable) {
	r.Runnables = append(r.Runnables, rn)
}

func (l *FixedHzLoop) Alpha(now time.Time) float32 {
	l.mu.Lock()
	last := l.lastTick
	delta := l.delta
	l.mu.Unlock()

	if delta <= 0 {
		return 1
	}

	alpha := float32(now.Sub(last).Seconds() / delta.Seconds())
	switch {
	case alpha < 0:
		return 0
	case alpha > 1:
		return 1
	default:
		return alpha
	}
}

func (l *FixedHzLoop) runLoop() {
	defer l.wg.Done()

	interval := time.Duration(float64(time.Second) / float64(l.Hz))

	l.mu.Lock()
	l.delta = interval
	l.lastTick = time.Now()
	l.mu.Unlock()

	nextTick := time.Now().Add(interval)
	usePureSpin := interval < minReliableSleep

	for {
		select {
		case <-l.stop:
			return
		default:
		}

		l.waitUntil(nextTick, usePureSpin)

		now := time.Now()
		atr, otr, monitorEvery, lastMonitor := l.swapBuffers(now)

		l.runOneTimeRunnables(otr)
		newActive := l.runRunnablesParallel(atr)

		l.mu.Lock()
		l.nextRunnables = append(l.nextRunnables, newActive...)
		l.tickCount++
		l.maybeReportMetrics(monitorEvery, lastMonitor)
		l.mu.Unlock()

		nextTick = l.scheduleNextTick(nextTick, interval)
	}
}

func (l *FixedHzLoop) waitUntil(nextTick time.Time, usePureSpin bool) {
	now := time.Now()
	remaining := nextTick.Sub(now)

	if remaining <= 0 {
		return
	}

	if usePureSpin {
		l.busyWait(nextTick)
		return
	}

	if remaining > spinWindow {
		time.Sleep(remaining - spinWindow)
		l.busyWait(nextTick)
	} else {
		l.busyWait(nextTick)
	}
}

func (l *FixedHzLoop) busyWait(nextTick time.Time) {
	for time.Now().Before(nextTick) {
		// Busy wait for precision
	}
}

func (l *FixedHzLoop) swapBuffers(now time.Time) (active, oneTime []Runnable, monitorEvery time.Duration, lastMonitor time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lastTick = now

	oneTime = l.oneTimeRunnables
	l.oneTimeRunnables = nil

	active = l.activeRunnables
	l.activeRunnables = l.nextRunnables
	l.nextRunnables = active[:0]

	return active, oneTime, l.monitorEvery, l.lastMonitor
}

func (l *FixedHzLoop) runOneTimeRunnables(runnables []Runnable) {
	for _, r := range runnables {
		if err := r(); err != nil {
			fmt.Println(err)
		}
	}
}

func (l *FixedHzLoop) runRunnablesParallel(runnables []Runnable) []Runnable {
	if len(runnables) == 0 {
		return runnables[:0]
	}

	var tickWg sync.WaitGroup
	results := make([]error, len(runnables))

	for i, r := range runnables {
		tickWg.Add(1)
		workerIdx := i % l.numWorkers
		l.workers[workerIdx] <- job{
			runnable: r,
			result:   &results[i],
			wg:       &tickWg,
		}
	}
	tickWg.Wait()

	newActive := runnables[:0]
	for i, r := range runnables {
		if results[i] == nil {
			newActive = append(newActive, r)
		} else {
			fmt.Println(results[i])
		}
	}
	return newActive
}

func (l *FixedHzLoop) maybeReportMetrics(monitorEvery time.Duration, lastMonitor time.Time) {
	if monitorEvery == 0 {
		return
	}
	if time.Since(l.lastMonitor) < monitorEvery {
		return
	}

	elapsed := time.Since(lastMonitor)
	hz := float64(l.tickCount) / elapsed.Seconds()
	avgTick := elapsed.Seconds() * 1000.0 / float64(l.tickCount)
	fmt.Printf("[FixedHzLoop] actual=%.1f Hz, avg=%.2f ms\n", hz, avgTick)

	l.lastMonitor = time.Now()
	l.tickCount = 0
}

func (l *FixedHzLoop) scheduleNextTick(nextTick time.Time, interval time.Duration) time.Time {
	nextTick = nextTick.Add(interval)
	if time.Since(nextTick) > interval {
		// We fell behind – reset to now
		return time.Now().Add(interval)
	}
	return nextTick
}
