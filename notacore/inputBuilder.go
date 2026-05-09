package notacore

type InputBuilder struct {
	node InputNode
}

// NewInputBuilder starts building from a raw input
func NewInputBuilder(name string, input StateInput) *InputBuilder {
	return &InputBuilder{
		node: NewRawInputNode(name, input),
	}
}

// From starts building from an existing node
func From(node InputNode) *InputBuilder {
	return &InputBuilder{
		node: node,
	}
}

// WithCooldown adds a cooldown modifier
func (b *InputBuilder) WithCooldown(frames int) *InputBuilder {
	b.node = NewCooldownNode(b.node.GetName()+"_cooldown", frames, b.node)
	return b
}

// WithHoldDuration requires the input to be held for a minimum duration
func (b *InputBuilder) WithHoldDuration(frames int) *InputBuilder {
	b.node = NewHoldDurationNode(b.node.GetName()+"_hold", frames, b.node)
	return b
}

// And combines with another input (logical AND)
func (b *InputBuilder) And(name string, other InputNode) *InputBuilder {
	b.node = NewCompositeNode(name, OpAnd, b.node, other)
	return b
}

// Or combines with another input (logical OR)
func (b *InputBuilder) Or(name string, other InputNode) *InputBuilder {
	b.node = NewCompositeNode(name, OpOr, b.node, other)
	return b
}

// Not inverts the current input
func (b *InputBuilder) Not(name string) *InputBuilder {
	b.node = NewCompositeNode(name, OpNot, b.node)
	return b
}

// Xor combines with another input (logical XOR)
func (b *InputBuilder) Xor(name string, other InputNode) *InputBuilder {
	b.node = NewCompositeNode(name, OpXor, b.node, other)
	return b
}

// Sequence creates a sequence detector
func (b *InputBuilder) Sequence(name string, timeoutFrames int, inputs ...InputNode) *InputBuilder {
	allInputs := append([]InputNode{b.node}, inputs...)
	b.node = NewSequenceNode(name, timeoutFrames, allInputs...)
	return b
}

// Build returns the final InputSignal
func (b *InputBuilder) Build(ctx *InputContext) *InputSignal {
	return NewInputSignal(b.node, ctx)
}

// Node returns the underlying node for advanced usage
func (b *InputBuilder) Node() InputNode {
	return b.node
}

// Input creates a signal from a single keyboard input
func Input(name string, key StateInput, ctx *InputContext) *InputSignal {
	return NewInputBuilder(name, key).Build(ctx)
}

// InputCombo creates a signal from multiple keys held together
func InputCombo(name string, ctx *InputContext, keys ...StateInput) *InputSignal {
	if len(keys) == 0 {
		return NewInputSignal(NewRawInputNode(name, 0), ctx)
	}
	if len(keys) == 1 {
		return Input(name, keys[0], ctx)
	}

	nodes := make([]InputNode, len(keys))
	for i, k := range keys {
		nodes[i] = NewRawInputNode(name+"_"+string(rune(i)), k)
	}

	return NewInputSignal(NewCompositeNode(name, OpAnd, nodes...), ctx)
}
