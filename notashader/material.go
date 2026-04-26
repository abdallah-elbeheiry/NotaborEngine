package notashader

import "sync"

const (
	DefaultVertexShaderPath   = "notashader/shaders/basic.vert"
	DefaultFragmentShaderPath = "notashader/shaders/basic.frag"
)

type Material struct {
	Shader *Shader

	mu       sync.RWMutex
	uniforms map[string]interface{}
}

// NewMaterial creates a material instance that wraps a shader plus a set of reusable uniform overrides.
func NewMaterial(shader *Shader) *Material {
	return &Material{
		Shader:   shader,
		uniforms: make(map[string]interface{}),
	}
}

// Clone creates a shallow copy of the material that shares the shader but owns an independent uniform set.
func (m *Material) Clone() *Material {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &Material{
		Shader:   m.Shader,
		uniforms: make(map[string]interface{}, len(m.uniforms)),
	}
	for key, value := range m.uniforms {
		clone.uniforms[key] = value
	}
	return clone
}

// Set stores a raw uniform value on the material and returns the material for chaining.
func (m *Material) Set(name string, value interface{}) *Material {
	m.mu.Lock()
	m.uniforms[name] = value
	m.mu.Unlock()
	return m
}

// Bool stores a boolean uniform value on the material.
func (m *Material) Bool(name string, value bool) *Material {
	return m.Set(name, value)
}

// Int stores an integer uniform value on the material.
func (m *Material) Int(name string, value int32) *Material {
	return m.Set(name, value)
}

// Float stores a float uniform value on the material.
func (m *Material) Float(name string, value float32) *Material {
	return m.Set(name, value)
}

// CircleMask enables the built-in circular mask uniforms on the material.
func (m *Material) CircleMask(radius, edge float32) *Material {
	return m.
		Bool(UseCircle, true).
		Float(CircleRadius, radius).
		Float(CircleEdge, edge)
}

// NoCircleMask disables the built-in circular mask on the material.
func (m *Material) NoCircleMask() *Material {
	return m.Bool(UseCircle, false)
}

// Apply binds the shader and pushes the material's uniform state for the current draw.
func (m *Material) Apply(textureEnabled bool) {
	if m == nil || m.Shader == nil {
		return
	}

	m.Shader.Bind()
	m.Shader.SetUniform(UseTexture, textureEnabled)
	m.Shader.SetUniform(TextureBind, int32(0))
	m.Shader.SetUniform(UseCircle, false)

	m.mu.RLock()
	defer m.mu.RUnlock()
	for name, value := range m.uniforms {
		m.Shader.SetUniform(name, value)
	}
}
