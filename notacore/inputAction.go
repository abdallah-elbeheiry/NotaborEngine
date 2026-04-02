package notacore

import (
	"NotaborEngine/notatask"
	"time"
)

type ActionBehavior int

const (
	RunWhileHeld ActionBehavior = iota
	RunWhileToggled
	RunOnceWhenPressed
	RunOnceWhenReleased
	RunWhileIdle
	Ignore
)

type Action struct {
	signal       *InputSignal
	Toggled      bool
	HeldTicks    int64
	LastHeldTime time.Duration
	Behavior     ActionBehavior
	lastHold     time.Time
	lastRelease  time.Time

	Cooldown time.Duration
	lastRun  time.Time

	tasks []*notatask.Task
}

// RunWhenShould checks if the action should run and schedules its tasks
func (a *Action) RunWhenShould(loop *notatask.Loop) bool {
	a.shouldToggle()
	a.updateHoldInformation()

	if time.Since(a.lastRun) < a.Cooldown {
		return false
	}

	var result bool
	switch a.Behavior {
	case RunWhileHeld:
		result = a.signal.Held()
	case RunWhileToggled:
		result = a.Toggled
	case RunOnceWhenPressed:
		result = a.signal.Pressed()
	case RunOnceWhenReleased:
		result = a.signal.Released()
	case RunWhileIdle:
		result = a.signal.Idle()
	case Ignore:
		return false
	}

	if result {
		a.lastRun = time.Now()
		for _, t := range a.tasks {
			loop.Add(t) // schedule task on the loop
		}
	}
	return result
}

func (a *Action) shouldToggle() {
	if a.signal.Pressed() {
		a.Toggled = !a.Toggled
	}
}

func (a *Action) updateHoldInformation() {
	if a.signal.Released() {
		a.lastRelease = time.Now()
		a.HeldTicks = 0
	}

	if a.signal.Held() {
		a.lastHold = time.Now()
		a.HeldTicks++
	}

	a.LastHeldTime = a.lastHold.Sub(a.lastRelease)
}

// AddTask adds a task to this action
func (a *Action) AddTask(t *notatask.Task) {
	a.tasks = append(a.tasks, t)
}

// BindSignal binds an input signal to this action
func (a *Action) BindSignal(sig *InputSignal) {
	a.signal = sig
}
