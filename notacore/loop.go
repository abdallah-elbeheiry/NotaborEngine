package notacore

import (
	"NotaborEngine/notatomic"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Runnable func() error

var ErrDone = errors.New("runnable finished")

type job struct {
	runnable Runnable
	result   *error
	wg       *sync.WaitGroup
}

var jobPool = sync.Pool{
	New: func() any {
		return &job{}
	},
}

type FixedHzLoop struct {
	Hz           notatomic.Float32
	monitorEvery notatomic.Int64 // Store as time.Duration (int64)
	lastMonitor  notatomic.Pointer[time.Time]

	lastTick  notatomic.Pointer[time.Time]
	delta     notatomic.Int64 // stored as time.Duration (int64)
	tickCount notatomic.UInt64

	// Double-buffered slices
	Runnables notatomic.Pointer[[]Runnable]

	stop       chan struct{}
	workerDone chan struct{}
	taskQueue  chan *job

	wg         sync.WaitGroup
	workerWg   sync.WaitGroup
	numWorkers int
}

type RenderLoop struct {
	MaxHz     notatomic.Float32
	Runnables notatomic.Pointer[[]Runnable]
	LastTime  notatomic.Pointer[time.Time]
}

func (l *FixedHzLoop) EnableMonitor(interval time.Duration) {
	l.monitorEvery.Set(int64(interval))
	now := time.Now()
	l.lastMonitor.Set(&now)
	l.tickCount.Set(0)
}

func (l *FixedHzLoop) Start() {
	if l.stop != nil {
		return
	}
	l.stop = make(chan struct{})

	if l.Runnables.Get() == nil {
		next := make([]Runnable, 0)
		l.Runnables.Set(&next)
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

func (l *FixedHzLoop) startWorkers() {
	l.taskQueue = make(chan *job, l.numWorkers*4)
	l.workerDone = make(chan struct{})

	for i := 0; i < l.numWorkers; i++ {
		l.workerWg.Add(1)
		go l.worker()
	}
}

func (l *FixedHzLoop) worker() {
	defer l.workerWg.Done()
	for {
		select {
		case j := <-l.taskQueue:
			if j == nil {
				continue
			}
			*j.result = j.runnable()
			j.wg.Done()
		case <-l.workerDone:
			return
		}
	}
}

func (l *FixedHzLoop) Stop() {
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

func (l *FixedHzLoop) Add(r Runnable) {
	const minCap = 32

	for {
		oldPtr := l.Runnables.Get()
		var old []Runnable
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

		newSlice := make([]Runnable, newLen, newCap)
		copy(newSlice, old)
		newSlice[len(old)] = r

		if l.Runnables.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

func (l *FixedHzLoop) Remove(r Runnable) {
	targetPtr := reflect.ValueOf(r).Pointer()
	for {
		oldPtr := l.Runnables.Get()
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

		newSlice := make([]Runnable, 0, len(oldSlice)-1)
		newSlice = append(newSlice, oldSlice[:foundIdx]...)
		newSlice = append(newSlice, oldSlice[foundIdx+1:]...)

		if l.Runnables.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

func (l *FixedHzLoop) Alpha(now time.Time) float32 {
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

func (l *FixedHzLoop) runLoop() {
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

		ptrNext := l.Runnables.Get()
		if ptrNext != nil {
			runnablesToRun := *ptrNext
			empty := make([]Runnable, 0)
			l.Runnables.Set(&empty)

			if len(runnablesToRun) > 0 {
				remainingActive := l.runRunnablesParallel(runnablesToRun)
				l.mergeBack(remainingActive)
			}
		}

		l.tickCount.Add(1)
		l.maybeReportMetrics()

		// Update Interval (in case Hz changed)
		newHz := l.Hz.Get()
		if newHz != hz && newHz > 0 {
			hz = newHz
			interval = time.Duration(float64(time.Second) / float64(hz))
			l.delta.Set(int64(interval))
		}

		nextTick = nextTick.Add(interval)
	}
}

func (l *FixedHzLoop) mergeBack(runnables []Runnable) {
	if len(runnables) == 0 {
		return
	}

	const minCap = 32

	for {
		oldPtr := l.Runnables.Get()
		var old []Runnable
		if oldPtr != nil {
			old = *oldPtr
		}

		newLen := len(old) + len(runnables)
		var newCap int
		if newLen < minCap {
			newCap = minCap
		} else {
			newCap = newLen * 2
		}

		newSlice := make([]Runnable, newLen, newCap)
		copy(newSlice, old)
		copy(newSlice[len(old):], runnables)

		if l.Runnables.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

func (l *FixedHzLoop) waitUntil(nextTick time.Time) {
	remaining := time.Until(nextTick)
	if remaining > 0 {
		time.Sleep(remaining)
	}
}

func (l *FixedHzLoop) runRunnablesParallel(runnables []Runnable) []Runnable {
	count := len(runnables)
	if count == 0 {
		return runnables[:0]
	}

	results := make([]error, count)

	var tickWg sync.WaitGroup
	jobsUsed := make([]*job, 0, count)

	for i, r := range runnables {
		tickWg.Add(1)
		j := jobPool.Get().(*job)
		j.runnable = r
		j.result = &results[i]
		j.wg = &tickWg
		jobsUsed = append(jobsUsed, j)
		l.taskQueue <- j
	}

	tickWg.Wait()

	for _, j := range jobsUsed {
		j.runnable = nil
		j.result = nil
		j.wg = nil
		jobPool.Put(j)
	}

	newIdx := 0
	for i := 0; i < count; i++ {
		err := results[i]

		if err == nil {
			runnables[newIdx] = runnables[i]
			newIdx++
			continue
		}

		if errors.Is(err, ErrDone) {
			continue // remove silently
		}

		fmt.Println("Runnable error:", err)
	}

	for i := newIdx; i < count; i++ {
		runnables[i] = nil
	}

	return runnables[:newIdx]
}

func (l *FixedHzLoop) maybeReportMetrics() {
	monitorEvery := time.Duration(l.monitorEvery.Get())
	lastMonPtr := l.lastMonitor.Get()

	if monitorEvery == 0 || lastMonPtr == nil || time.Since(*lastMonPtr) < monitorEvery {
		return
	}

	elapsed := time.Since(*lastMonPtr)
	count := l.tickCount.GetAndSet(0)
	hz := float64(count) / elapsed.Seconds()
	avgTickMs := (elapsed.Seconds() * 1000.0) / float64(count)

	fmt.Printf("[FixedHzLoop] Target=%.0f Hz, Actual=%.1f Hz, AvgTick=%.4f ms\n", l.Hz.Get(), hz, avgTickMs)

	now := time.Now()
	l.lastMonitor.Set(&now)
}

func (r *RenderLoop) Render() {
	now := time.Now()
	r.LastTime.Set(&now)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	runnablesPtr := r.Runnables.Get()
	if runnablesPtr != nil {
		for _, runnable := range *runnablesPtr {
			if err := runnable(); err != nil {
				fmt.Println("Render error:", err)
			}
		}
	}
}

func (r *RenderLoop) Add(runnable Runnable) {
	const minCap = 32

	for {
		oldPtr := r.Runnables.Get()
		var old []Runnable
		if oldPtr != nil {
			old = *oldPtr
		}

		newLen := len(old) + 1
		var newCap int
		if newLen < minCap {
			newCap = minCap
		} else {
			newCap = newLen * 2
		}

		newSlice := make([]Runnable, newLen, newCap)
		copy(newSlice, old)
		newSlice[len(old)] = runnable

		if r.Runnables.CompareAndSwap(oldPtr, &newSlice) {
			return
		}
	}
}

func CreateFixedHzLoop(Hz float32) *FixedHzLoop {
	loop := &FixedHzLoop{}
	loop.Hz.Set(Hz)
	next := make([]Runnable, 0)
	loop.Runnables.Set(&next)
	now := time.Now()
	loop.lastTick.Set(&now)
	loop.lastMonitor.Set(&now)

	return loop
}

func CreateRenderLoop(Hz float32) *RenderLoop {
	loop := &RenderLoop{}
	loop.MaxHz.Set(Hz)
	r := make([]Runnable, 0)
	loop.Runnables.Set(&r)
	now := time.Now()
	loop.LastTime.Set(&now)

	return loop
}

// Once wraps around a runnable making it run once
func Once(fn Runnable) Runnable {
	return func() error {
		if err := fn(); err != nil {
			return err
		}
		return ErrDone
	}
}

// Delay wraps around the runnable, making it run once after a duration
func Delay(fn Runnable, d time.Duration) Runnable {
	start := time.Now()

	return func() error {
		if time.Since(start) < d {
			return nil
		}

		if err := fn(); err != nil {
			return err
		}
		return ErrDone
	}
}

// Every wraps around the runnable, making it run every duration
func Every(fn Runnable, interval time.Duration) Runnable {
	last := time.Now()

	return func() error {
		if time.Since(last) < interval {
			return nil
		}

		last = last.Add(interval)
		if err := fn(); err != nil {
			return err
		}

		return nil
	}
}

// AfterTicks wraps around the runnable, making it run once after a number of ticks
func AfterTicks(fn Runnable, ticks int) Runnable {
	count := 0

	return func() error {
		count++

		if count < ticks {
			return nil
		}

		if err := fn(); err != nil {
			return err
		}
		return ErrDone
	}
}

// EveryTicks wraps around the runnable, making it run every number of ticks
func EveryTicks(fn Runnable, ticks int) Runnable {
	count := 0

	return func() error {
		count++

		if count < ticks {
			return nil
		}

		count = 0
		if err := fn(); err != nil {
			return err
		}
		return nil
	}
}

// Repeat wraps around the runnable, making it run a number of times
func Repeat(fn Runnable, times int) Runnable {
	count := 0
	return func() error {
		if err := fn(); err != nil {
			return err
		}

		count++

		if count < times {
			return nil
		}
		return ErrDone
	}
}

// RepeatUntil wraps around the runnable, making it run a number of times until a condition is met
func RepeatUntil(fn Runnable, cond func() bool) Runnable {
	return func() error {
		if err := fn(); err != nil {
			return err
		}

		if cond() {
			return ErrDone
		}
		return nil
	}
}

// FinishAfter wraps around the runnable, making it run until a duration has passed
func FinishAfter(fn Runnable, d time.Duration) Runnable {
	start := time.Now()
	return func() error {
		if err := fn(); err != nil {
			return err
		}

		if time.Since(start) < d {
			return nil
		}
		return ErrDone
	}
}
