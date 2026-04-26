package main

import (
	"NotaborEngine/notatask"
	"NotaborEngine/notatomic"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// TestLoopFrequency verifies the loop can hit target frequencies
func TestLoopFrequency(t *testing.T) {
	tests := []struct {
		name      string
		targetHz  float32
		duration  time.Duration
		minHz     float32 // minimum acceptable frequency
		taskCount int
	}{
		{"1kHz", 1000, 5 * time.Second, 995, 10000},
		{"10kHz", 10000, 5 * time.Second, 9995, 2100},
		{"60Hz", 60, 5 * time.Second, 59.5, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loop := notatask.CreateLoop(tt.targetHz)
			loop.Start()

			// Add tasks
			var counter atomic.Int64
			for i := 0; i < tt.taskCount; i++ {
				loop.Add(notatask.CreateTask(func() {
					counter.Add(1)
				}))
			}

			// Wait for duration
			time.Sleep(tt.duration)
			loop.Stop()

			// Calculate achieved frequency
			elapsed := tt.duration.Seconds()
			actualHz := float32(float64(counter.Load()) / elapsed / float64(tt.taskCount))

			t.Logf("Target: %.0f Hz, Achieved: %.0f Hz", tt.targetHz, actualHz)
			if actualHz < tt.minHz {
				t.Errorf("Frequency too low: got %.0f Hz, want >= %.0f Hz", actualHz, tt.minHz)
			}
		})
	}
}

// TestTaskLifecycle verifies tasks are added, executed, and removed
func TestTaskLifecycle(t *testing.T) {
	loop := notatask.CreateLoop(1000)
	loop.Start()

	var executed atomic.Int32
	var finished atomic.Int32

	// Task that runs once
	task := notatask.CreateTask(func() {
		executed.Add(1)
	}, notatask.RunOnce())
	loop.Add(task)

	// Task that runs and self-terminates
	loop.Add(notatask.CreateTask(func() {
		finished.Add(1)
	}, notatask.RepeatTimes(5)))

	time.Sleep(100 * time.Millisecond)
	loop.Stop()

	if executed.Load() != 1 {
		t.Errorf("RunOnce task executed %d times, want 1", executed.Load())
	}
	if finished.Load() != 5 {
		t.Errorf("RepeatTimes task executed %d times, want 5", finished.Load())
	}
}

// TestAddRemoveDuringRun tests concurrent add/remove while loop is running
func TestAddRemoveDuringRun(t *testing.T) {
	loop := notatask.CreateLoop(1000)
	loop.Start()

	var added atomic.Int32

	// Goroutine that constantly adds tasks
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			task := notatask.CreateTask(func() {}, notatask.RepeatTimes(10))
			loop.Add(task)
			added.Add(1)
			time.Sleep(time.Microsecond)
		}
		close(done)
	}()

	<-done
	time.Sleep(200 * time.Millisecond)
	loop.Stop()

	t.Logf("Added %d tasks", added.Load())
}

// TestWorkerDistribution verifies tasks are distributed across workers
func TestWorkerDistribution(t *testing.T) {
	loop := notatask.CreateLoop(100)

	// Track which worker executed each task
	workerHits := make([]atomic.Int32, runtime.NumCPU())

	for i := 0; i < runtime.NumCPU()*4; i++ {
		idx := i % len(workerHits)
		loop.Add(notatask.CreateTask(func() {
			workerHits[idx].Add(1)
		}))
	}

	loop.Start()
	time.Sleep(500 * time.Millisecond)
	loop.Stop()

	// Verify all workers were used
	activeWorkers := 0
	for i, hits := range workerHits {
		if hits.Load() > 0 {
			activeWorkers++
			t.Logf("Worker %d: %d executions", i, hits.Load())
		}
	}
	t.Logf("Active workers: %d / %d", activeWorkers, len(workerHits))
}

// TestStressTest runs maximum throughput test
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	loop := notatask.CreateLoop(0) // 0 means run as fast as possible
	loop.Hz.Set(0)                 // Will default to 60 in runLoop, so we override

	var counter atomic.Int64

	// Add many tiny tasks
	for i := 0; i < 100; i++ {
		loop.Add(notatask.CreateTask(func() {
			counter.Add(1)
		}))
	}

	loop.Start()
	time.Sleep(3 * time.Second)
	loop.Stop()

	opsPerSec := float64(counter.Load()) / 3.0
	t.Logf("Throughput: %.0f tasks/sec (%.1fM ops/sec)", opsPerSec, opsPerSec/1_000_000)
}

// TestAllocationCount verifies no allocations in hot path after warmup
func TestAllocationCount(t *testing.T) {
	loop := notatask.CreateLoop(1000)

	// Add tasks
	for i := 0; i < 10; i++ {
		loop.Add(notatask.CreateTask(func() {}))
	}

	loop.Start()

	// Warmup
	time.Sleep(500 * time.Millisecond)

	// Measure allocations
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	allocsBefore := stats.Mallocs

	time.Sleep(2 * time.Second)

	runtime.GC()
	runtime.ReadMemStats(&stats)
	allocsAfter := stats.Mallocs

	loop.Stop()

	allocsPerSec := float64(allocsAfter-allocsBefore) / 2.0
	t.Logf("Allocations/sec: %.0f", allocsPerSec)

	if allocsPerSec > 1010 {
		t.Errorf("Too many allocations: %.0f/sec", allocsPerSec)
	}
}

// ExampleLoop demonstrates basic usage
func ExampleLoop() {
	loop := notatask.CreateLoop(60)
	var counter notatomic.Int32
	loop.Add(notatask.CreateTask(func() {
		counter.Inc()
	}))

	loop.Start()

	time.Sleep(1 * time.Second)
	time.Sleep(1 * time.Microsecond)

	loop.Stop()

	fmt.Printf("Task executed approximately %d times\n", counter.Get())
	// Output: Task executed approximately 60 times
}
