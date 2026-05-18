package notacore

import (
	"NotaborEngine/notatomic"
)

type InputContext struct {
	// Current hardware state (what is physically pressed right now)
	currHardware *notatomic.Pointer[map[StateInput]bool]

	// Snapshot taken at the beginning of the frame (safe for readers)
	snapshot *notatomic.Pointer[inputSnapshot]
}

// inputSnapshot holds a consistent view of input state for one frame
type inputSnapshot struct {
	PrevHardware map[StateInput]bool
	CurrHardware map[StateInput]bool

	KeyDownEvents map[StateInput]bool
	KeyUpEvents   map[StateInput]bool
}

func NewInputContext() *InputContext {
	emptyMap := make(map[StateInput]bool)

	ctx := &InputContext{
		currHardware: notatomic.NewPointer(&emptyMap),
		snapshot:     notatomic.NewPointer(&inputSnapshot{}),
	}

	// Initialize snapshot
	initialSnap := &inputSnapshot{
		PrevHardware:  make(map[StateInput]bool),
		CurrHardware:  make(map[StateInput]bool),
		KeyDownEvents: make(map[StateInput]bool),
		KeyUpEvents:   make(map[StateInput]bool),
	}
	ctx.snapshot.Set(initialSnap)

	return ctx
}

// beginFrame should be called at the start of each frame to update state
// This is the only place that does significant work and should be called from the input loop.
func (c *InputContext) beginFrame() {
	oldSnap := c.snapshot.Get()

	newSnap := &inputSnapshot{
		PrevHardware:  make(map[StateInput]bool),
		CurrHardware:  make(map[StateInput]bool),
		KeyDownEvents: make(map[StateInput]bool),
		KeyUpEvents:   make(map[StateInput]bool),
	}

	curr := *c.currHardware.Get()
	for k, v := range curr {
		newSnap.PrevHardware[k] = v
		newSnap.CurrHardware[k] = v
	}

	if oldSnap != nil {
		for k := range oldSnap.KeyDownEvents {
			newSnap.KeyDownEvents[k] = true
		}
		for k := range oldSnap.KeyUpEvents {
			newSnap.KeyUpEvents[k] = true
		}
	}

	c.snapshot.Set(newSnap)
}

// recordKeyDown is called when a hardware key down event occurs
func (c *InputContext) recordKeyDown(input StateInput) {
	m := *c.currHardware.Get()
	m[input] = true
}

// recordKeyUp is called when a hardware key up event occurs
func (c *InputContext) recordKeyUp(input StateInput) {
	m := *c.currHardware.Get()
	m[input] = false
}

// isKeyHeldThisFrame returns true if the key is physically pressed this frame
func (c *InputContext) isKeyHeldThisFrame(input StateInput) bool {
	snap := c.snapshot.Get()
	if snap == nil {
		return false
	}
	return snap.CurrHardware[input]
}

// wasKeyHeldLastFrame returns true if the key was physically pressed last frame
func (c *InputContext) wasKeyHeldLastFrame(input StateInput) bool {
	snap := c.snapshot.Get()
	if snap == nil {
		return false
	}
	return snap.PrevHardware[input]
}

// isKeyDownThisFrame returns true if a KeyDown event occurred for this input this frame
func (c *InputContext) isKeyDownThisFrame(input StateInput) bool {
	snap := c.snapshot.Get()
	if snap == nil {
		return false
	}
	return snap.KeyDownEvents[input]
}

// isKeyUpThisFrame returns true if a KeyUp event occurred for this input this frame
func (c *InputContext) isKeyUpThisFrame(input StateInput) bool {
	snap := c.snapshot.Get()
	if snap == nil {
		return false
	}
	return snap.KeyUpEvents[input]
}

// GetState returns the current input state for debugging/inspection
func (c *InputContext) GetState(input StateInput) InputState {
	snap := c.snapshot.Get()
	if snap == nil {
		return StateIdle
	}

	curr := snap.CurrHardware[input]
	prev := snap.PrevHardware[input]

	switch {
	case curr && !prev:
		return StatePressed
	case curr && prev:
		return StateHeld
	case !curr && prev:
		return StateReleased
	default:
		return StateIdle
	}
}
