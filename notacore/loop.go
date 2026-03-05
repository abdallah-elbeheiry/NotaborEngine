package notacore

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Runnable func() error

type FixedHzLoop struct {
	Hz float32

	mu               sync.Mutex
	Runnables        []Runnable
	OneTimeRunnables []Runnable

	stop chan struct{}
	wg   sync.WaitGroup

	lastTick time.Time
	delta    time.Duration

	monitorEvery time.Duration
	lastMonitor  time.Time
	tickCount    uint64
}

func (l *FixedHzLoop) EnableMonitor(interval time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.monitorEvery = interval
	l.lastMonitor = time.Now()
	l.tickCount = 0
}

type RenderLoop struct {
	MaxHz     float32
	Runnables []Runnable
	LastTime  time.Time
}

func (l *FixedHzLoop) Start() {

	l.mu.Lock()
	if l.stop != nil {
		l.mu.Unlock()
		return
	}
	l.stop = make(chan struct{})
	l.mu.Unlock()

	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		interval := time.Duration(float64(time.Second) / float64(l.Hz))

		l.mu.Lock()
		l.delta = interval
		l.lastTick = time.Now()
		l.mu.Unlock()

		nextTick := time.Now().Add(interval)

		for {
			select {
			case <-l.stop:
				return
			default:
			}

			for {
				now := time.Now()
				remaining := nextTick.Sub(now)

				if remaining <= 0 {
					break
				}

				if remaining > 100*time.Microsecond {
					time.Sleep(remaining - 100*time.Microsecond)
				} else {

				}
			}

			now := time.Now()

			l.mu.Lock()
			l.lastTick = now

			otr := l.OneTimeRunnables
			l.OneTimeRunnables = nil

			atr := append([]Runnable(nil), l.Runnables...)

			monitorEvery := l.monitorEvery
			lastMonitor := l.lastMonitor

			l.mu.Unlock()

			for _, r := range otr {
				if err := r(); err != nil {
					fmt.Println(err)
				}
			}

			newRunnables := atr[:0]
			for _, r := range atr {
				if err := r(); err != nil {
					fmt.Println(err)
				} else {
					newRunnables = append(newRunnables, r)
				}
			}

			l.mu.Lock()
			l.Runnables = newRunnables

			l.tickCount++

			if monitorEvery > 0 && time.Since(lastMonitor) >= monitorEvery {
				elapsed := time.Since(lastMonitor)
				hz := float64(l.tickCount) / elapsed.Seconds()
				avgTick := elapsed.Seconds() * 1000.0 / float64(l.tickCount)

				fmt.Printf("[FixedHzLoop] actual=%.1f Hz, avg=%.2f ms\n", hz, avgTick)

				l.lastMonitor = time.Now()
				l.tickCount = 0
			}

			l.mu.Unlock()

			nextTick = nextTick.Add(interval)

			if time.Since(nextTick) > interval {
				nextTick = time.Now().Add(interval)
			}
		}
	}()
}

func (l *FixedHzLoop) Stop() {
	close(l.stop)
	l.wg.Wait()
}

func (l *FixedHzLoop) Remove(i int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	last := len(l.Runnables) - 1
	if i < 0 || i > last {
		return
	}

	l.Runnables[i] = l.Runnables[last]
	l.Runnables[last] = nil
	l.Runnables = l.Runnables[:last]
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

func (r *RenderLoop) Add(runnable Runnable) {
	r.Runnables = append(r.Runnables, runnable)
}

func (l *FixedHzLoop) Add(runnable Runnable) {
	l.mu.Lock()
	l.Runnables = append(l.Runnables, runnable)
	l.mu.Unlock()
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

	if alpha < 0 {
		return 0
	}
	if alpha > 1 {
		return 1
	}
	return alpha
}
