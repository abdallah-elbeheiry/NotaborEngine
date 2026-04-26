package notashader

import (
	"fmt"
	"path/filepath"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	shaders map[string]*Shader
}

// NewManager creates a shader manager that caches shaders by name.
func NewManager() *Manager {
	return &Manager{
		shaders: make(map[string]*Shader),
	}
}

// Load loads a shader pair once and returns the cached instance for subsequent calls with the same name.
func (m *Manager) Load(name, vertexPath, fragmentPath string) (*Shader, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if shader, ok := m.shaders[name]; ok {
		return shader, nil
	}

	vertAbs, err := filepath.Abs(vertexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve vertex shader path: %w", err)
	}

	fragAbs, err := filepath.Abs(fragmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve fragment shader path: %w", err)
	}

	shader, err := NewShader(vertAbs, fragAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to load shader '%s': %w", name, err)
	}

	m.shaders[name] = shader
	return shader, nil
}

// Get returns a cached shader by name.
func (m *Manager) Get(name string) (*Shader, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shader, ok := m.shaders[name]
	if !ok {
		return nil, fmt.Errorf("shader '%s' not found", name)
	}

	return shader, nil
}

// Reload recompiles a cached shader in place.
func (m *Manager) Reload(name string) error {
	m.mu.RLock()
	shader, ok := m.shaders[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("shader '%s' not found", name)
	}

	return shader.Reload()
}

// Unload deletes a cached shader and releases its GPU program.
func (m *Manager) Unload(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	shader, ok := m.shaders[name]
	if !ok {
		return fmt.Errorf("shader '%s' not found", name)
	}

	shader.Delete()
	delete(m.shaders, name)
	return nil
}
