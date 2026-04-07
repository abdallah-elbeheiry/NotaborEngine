package notacore

// InputSignal tracks a binary input across frames/ticks.
// Update State during the frame, then call Snapshot() once per tick to advance LastState.
type InputSignal struct {
	State     bool
	LastState bool
}

// Set updates the current state for this tick/frame.
func (s *InputSignal) Set(state bool) {
	s.State = state
}

// Snapshot advances LastState to match the current State.
func (s *InputSignal) Snapshot() {
	s.LastState = s.State
}

// Down reports whether the input is currently down.
func (s *InputSignal) Down() bool {
	return s.State
}

// Changed reports whether the input state has changed since the last snapshot.
func (s *InputSignal) Changed() bool {
	return s.State != s.LastState
}

// Held reports whether the input is currently held down.
func (s *InputSignal) Held() bool {
	return s.State && s.LastState
}

// Released reports whether the input was released since the last snapshot.
func (s *InputSignal) Released() bool {
	return !s.State && s.LastState
}

// Pressed reports whether the input was pressed since the last snapshot.
func (s *InputSignal) Pressed() bool {
	return s.State && !s.LastState
}

// Idle reports whether the input is neither pressed nor released.
func (s *InputSignal) Idle() bool {
	return !s.State && !s.LastState
}

// Clone returns a copy of the InputSignal.
func (s *InputSignal) Clone() InputSignal {
	return InputSignal{
		State:     s.State,
		LastState: s.LastState,
	}
}
