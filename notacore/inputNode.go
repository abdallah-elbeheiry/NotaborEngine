package notacore

import "sync"

// InputState represents the logical state of an input signal
type InputState int

const (
	StateIdle     InputState = iota // StateInput is inactive (not pressed)
	StateHeld                       // StateInput is active and was active last frame
	StatePressed                    // StateInput became active this frame
	StateReleased                   // StateInput became inactive this frame
)

// InputNode is the base interface for all nodes in the input DAG
type InputNode interface {
	// Evaluate returns the current state based on input context
	Evaluate(ctx *InputContext) InputState

	// AddDependency registers a node that this node depends on
	AddDependency(node InputNode)

	// GetDependencies returns all nodes this node depends on
	GetDependencies() []InputNode

	// GetName returns the node's identifier
	GetName() string
}

// BaseNode provides common functionality for all input nodes
type BaseNode struct {
	name         string
	dependencies []InputNode
	mu           sync.RWMutex
}

func NewBaseNode(name string) BaseNode {
	return BaseNode{
		name:         name,
		dependencies: []InputNode{},
	}
}

func (b *BaseNode) AddDependency(node InputNode) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dependencies = append(b.dependencies, node)
}

func (b *BaseNode) GetDependencies() []InputNode {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]InputNode{}, b.dependencies...)
}

func (b *BaseNode) GetName() string {
	return b.name
}

// RawInputNode reads directly from hardware input
type RawInputNode struct {
	BaseNode
	input StateInput
}

func NewRawInputNode(name string, input StateInput) *RawInputNode {
	return &RawInputNode{
		BaseNode: NewBaseNode(name),
		input:    input,
	}
}

func (n *RawInputNode) Evaluate(ctx *InputContext) InputState {
	// Use events directly for immediate feedback, fall back to state transitions
	if ctx.isKeyDownThisFrame(n.input) {
		return StatePressed
	}

	if ctx.isKeyUpThisFrame(n.input) {
		return StateReleased
	}

	// For states between events, use hardware state tracking
	isHeldNow := ctx.isKeyHeldThisFrame(n.input)
	wasHeldBefore := ctx.wasKeyHeldLastFrame(n.input)

	switch {
	case isHeldNow && wasHeldBefore:
		return StateHeld
	case !isHeldNow && !wasHeldBefore:
		return StateIdle
	}
	return StateIdle
}

// CompositeNode combines multiple child nodes with a logical operator
type CompositeNode struct {
	BaseNode
	children []InputNode
	op       CompositeOp
}

type CompositeOp int

const (
	OpAnd CompositeOp = iota
	OpOr
	OpNot
	OpXor
)

func NewCompositeNode(name string, op CompositeOp, children ...InputNode) *CompositeNode {
	return &CompositeNode{
		BaseNode: NewBaseNode(name),
		children: children,
		op:       op,
	}
}

func (n *CompositeNode) Evaluate(ctx *InputContext) InputState {
	if len(n.children) == 0 {
		return StateIdle
	}

	switch n.op {
	case OpAnd:
		return n.evaluateAnd(ctx)
	case OpOr:
		return n.evaluateOr(ctx)
	case OpNot:
		return n.evaluateNot(ctx)
	case OpXor:
		return n.evaluateXor(ctx)
	default:
		return StateIdle
	}
}

func (n *CompositeNode) evaluateAnd(ctx *InputContext) InputState {
	// AND: all children must be pressed/held for result to be pressed/held
	allHeld := true
	anyPressed := false

	for _, child := range n.children {
		state := child.Evaluate(ctx)
		if state == StateIdle || state == StateReleased {
			allHeld = false
		}
		if state == StatePressed {
			anyPressed = true
		}
	}

	if !allHeld {
		return StateIdle
	}

	if anyPressed {
		return StatePressed
	}

	return StateHeld
}

func (n *CompositeNode) evaluateOr(ctx *InputContext) InputState {
	// OR: any child pressed/held makes result pressed/held
	anyPressed := false

	for _, child := range n.children {
		state := child.Evaluate(ctx)
		if state == StatePressed {
			anyPressed = true
		}
		if state == StateHeld || state == StatePressed {
			if anyPressed {
				return StatePressed
			}
			return StateHeld
		}
	}

	return StateIdle
}

func (n *CompositeNode) evaluateNot(ctx *InputContext) InputState {
	// NOT: invert the single child state
	if len(n.children) != 1 {
		return StateIdle
	}

	state := n.children[0].Evaluate(ctx)
	switch state {
	case StateIdle:
		return StateHeld // Held when child is idle (but not "pressed" at start)
	case StateHeld:
		return StateIdle
	case StatePressed:
		return StateReleased // Inverted press
	case StateReleased:
		return StatePressed // Inverted release
	default:
		return StateIdle
	}
}

func (n *CompositeNode) evaluateXor(ctx *InputContext) InputState {
	// XOR: true when odd number of children are pressed/held
	activeCount := 0
	anyPressed := false

	for _, child := range n.children {
		state := child.Evaluate(ctx)
		if state == StatePressed || state == StateHeld {
			activeCount++
			if state == StatePressed {
				anyPressed = true
			}
		}
	}

	isActive := activeCount%2 == 1

	if !isActive {
		return StateIdle
	}

	if anyPressed {
		return StatePressed
	}

	return StateHeld
}

// SequenceNode detects a sequence of inputs in order
type SequenceNode struct {
	BaseNode
	sequence     []InputNode
	currentIndex int
	frameTimeout int
	frameCount   int
}

func NewSequenceNode(name string, frameTimeout int, inputs ...InputNode) *SequenceNode {
	return &SequenceNode{
		BaseNode:     NewBaseNode(name),
		sequence:     inputs,
		currentIndex: 0,
		frameTimeout: frameTimeout,
		frameCount:   0,
	}
}

func (n *SequenceNode) Evaluate(ctx *InputContext) InputState {
	if len(n.sequence) == 0 {
		return StateIdle
	}

	n.frameCount++

	// Reset on timeout
	if n.frameCount > n.frameTimeout {
		n.currentIndex = 0
		n.frameCount = 0
	}

	currentNode := n.sequence[n.currentIndex]
	state := currentNode.Evaluate(ctx)

	if state == StatePressed {
		n.currentIndex++
		n.frameCount = 0

		// Sequence complete
		if n.currentIndex >= len(n.sequence) {
			n.currentIndex = 0
			return StatePressed
		}

		return StateIdle
	}

	return StateIdle
}

// CooldownNode prevents a signal from repeating within a cooldown period
type CooldownNode struct {
	BaseNode
	child        InputNode
	cooldownTime int // in frames
	frameCount   int
	lastPressed  bool
}

func NewCooldownNode(name string, cooldownFrames int, child InputNode) *CooldownNode {
	n := &CooldownNode{
		BaseNode:     NewBaseNode(name),
		child:        child,
		cooldownTime: cooldownFrames,
		frameCount:   0,
		lastPressed:  false,
	}
	n.AddDependency(child)
	return n
}

func (n *CooldownNode) Evaluate(ctx *InputContext) InputState {
	n.frameCount++

	if n.frameCount > n.cooldownTime {
		n.frameCount = 0
	}

	state := n.child.Evaluate(ctx)

	// If cooldown active, suppress input
	if n.frameCount < n.cooldownTime {
		return StateIdle
	}

	// Detect press and start cooldown
	if state == StatePressed && !n.lastPressed {
		n.lastPressed = true
		n.frameCount = 1 // Start cooldown countdown
		return StatePressed
	}

	n.lastPressed = state == StateHeld || state == StatePressed

	return StateIdle
}

// HoldDurationNode fires when held for a minimum duration
type HoldDurationNode struct {
	BaseNode
	child      InputNode
	holdFrames int // minimum frames to hold
	frameCount int
	wasHeld    bool
}

func NewHoldDurationNode(name string, holdFrames int, child InputNode) *HoldDurationNode {
	n := &HoldDurationNode{
		BaseNode:   NewBaseNode(name),
		child:      child,
		holdFrames: holdFrames,
		frameCount: 0,
		wasHeld:    false,
	}
	n.AddDependency(child)
	return n
}

func (n *HoldDurationNode) Evaluate(ctx *InputContext) InputState {
	state := n.child.Evaluate(ctx)

	switch state {
	case StatePressed:
		n.frameCount = 1
		return StateIdle
	case StateHeld:
		n.frameCount++
		if n.frameCount >= n.holdFrames && !n.wasHeld {
			n.wasHeld = true
			return StatePressed // Fire once when threshold reached
		}
		if n.frameCount >= n.holdFrames {
			return StateHeld
		}
		return StateIdle
	case StateReleased:
		n.frameCount = 0
		n.wasHeld = false
		return StateReleased
	default:
		n.frameCount = 0
		n.wasHeld = false
		return StateIdle
	}
}
