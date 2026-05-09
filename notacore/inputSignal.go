package notacore

import "sync"

// InputSignal represents a bindable input action with a backing InputNode
type InputSignal struct {
	node InputNode
	mu   sync.RWMutex
	ctx  *InputContext
}

// NewInputSignal creates a signal from an InputNode
func NewInputSignal(node InputNode, ctx *InputContext) *InputSignal {
	return &InputSignal{
		node: node,
		ctx:  ctx,
	}
}

// State returns the current state of this signal
func (s *InputSignal) State() InputState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.node.Evaluate(s.ctx)
}

// Idle returns true if the signal is in idle state
func (s *InputSignal) Idle() bool {
	return s.State() == StateIdle
}

// Held returns true if the signal is held (pressed or was already held)
func (s *InputSignal) Held() bool {
	state := s.State()
	return state == StateHeld || state == StatePressed
}

// Pressed returns true if the signal just entered active state
func (s *InputSignal) Pressed() bool {
	return s.State() == StatePressed
}

// Released returns true if the signal just left active state
func (s *InputSignal) Released() bool {
	return s.State() == StateReleased
}

// Down is an alias for Held() for backward compatibility
func (s *InputSignal) Down() bool {
	return s.Held()
}

// GetNode returns the underlying InputNode for this signal
func (s *InputSignal) GetNode() InputNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.node
}
