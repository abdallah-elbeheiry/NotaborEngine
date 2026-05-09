package notacore

type InputContext struct {
	// Hardware state tracking
	PrevHardware map[StateInput]bool // Hardware state from last frame
	CurrHardware map[StateInput]bool // Current hardware state (is key physically pressed)

	// Event tracking for this frame
	KeyDownEvents map[StateInput]bool // Keys that had KeyDown events this frame
	KeyUpEvents   map[StateInput]bool // Keys that had KeyUp events this frame

	// Saved events from current frame before clearing
	frameKeyDownEvents map[StateInput]bool
	frameKeyUpEvents   map[StateInput]bool
}

func NewInputContext() *InputContext {
	return &InputContext{
		PrevHardware:       make(map[StateInput]bool),
		CurrHardware:       make(map[StateInput]bool),
		KeyDownEvents:      make(map[StateInput]bool),
		KeyUpEvents:        make(map[StateInput]bool),
		frameKeyDownEvents: make(map[StateInput]bool),
		frameKeyUpEvents:   make(map[StateInput]bool),
	}
}

// BeginFrame should be called at the start of each frame to update state
func (c *InputContext) BeginFrame() {
	// Save THIS frame's events before clearing for evaluation
	c.frameKeyDownEvents = make(map[StateInput]bool)
	c.frameKeyUpEvents = make(map[StateInput]bool)
	for k, v := range c.KeyDownEvents {
		c.frameKeyDownEvents[k] = v
	}
	for k, v := range c.KeyUpEvents {
		c.frameKeyUpEvents[k] = v
	}

	// Shift hardware state to previous (for state-based fallback)
	for k, v := range c.CurrHardware {
		c.PrevHardware[k] = v
	}
	// Clear event flags for NEXT frame
	c.KeyDownEvents = make(map[StateInput]bool)
	c.KeyUpEvents = make(map[StateInput]bool)
}

// GetState returns the current input state for debugging/inspection
func (c *InputContext) GetState(input StateInput) InputState {
	curr := c.CurrHardware[input]
	prev := c.PrevHardware[input]

	switch {
	case curr && !prev:
		return StatePressed
	default:
		return StateIdle
	}
}

// RecordKeyDown is called when a hardware key down event occurs
func (c *InputContext) RecordKeyDown(input StateInput) {
	c.CurrHardware[input] = true
	c.KeyDownEvents[input] = true
}

// RecordKeyUp is called when a hardware key up event occurs
func (c *InputContext) RecordKeyUp(input StateInput) {
	c.CurrHardware[input] = false
	c.KeyUpEvents[input] = true
}

// IsKeyDownThisFrame returns true if a KeyDown event occurred for this input this frame
func (c *InputContext) IsKeyDownThisFrame(input StateInput) bool {
	return c.frameKeyDownEvents[input]
}

// IsKeyUpThisFrame returns true if a KeyUp event occurred for this input this frame
func (c *InputContext) IsKeyUpThisFrame(input StateInput) bool {
	return c.frameKeyUpEvents[input]
}

// IsKeyHeldThisFrame returns true if the key is physically pressed this frame
func (c *InputContext) IsKeyHeldThisFrame(input StateInput) bool {
	return c.CurrHardware[input]
}

// WasKeyHeldLastFrame returns true if the key was physically pressed last frame
func (c *InputContext) WasKeyHeldLastFrame(input StateInput) bool {
	return c.PrevHardware[input]
}
