package notagl

import (
	"fmt"
	"sync"
)

type TextureManager struct {
	textures map[string]*Texture
	mu       sync.RWMutex
}

func NewTextureManager() *TextureManager {
	return &TextureManager{
		textures: make(map[string]*Texture),
	}
}

// Load Update Load method
func (tm *TextureManager) Load(name, path string, createGL bool) (*Texture, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tex, ok := tm.textures[name]; ok {
		return tex, nil
	}

	tex, err := LoadImageData(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture '%s' from '%s': %w", name, path, err)
	}

	if createGL {
		if err := tex.CreateGLTexture(); err != nil {
			return nil, fmt.Errorf("failed to create OpenGL texture for '%s': %w", name, err)
		}
	}

	tm.textures[name] = tex
	return tex, nil
}

func (tm *TextureManager) CreateGLTextures() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for name, tex := range tm.textures {
		if !tex.Loaded {
			if err := tex.CreateGLTexture(); err != nil {
				return fmt.Errorf("failed to create OpenGL texture for '%s': %w", name, err)
			}
		}
	}
	return nil
}

func (tm *TextureManager) Get(name string) (*Texture, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tex, ok := tm.textures[name]
	if !ok {
		return nil, fmt.Errorf("texture '%s' not found", name)
	}

	return tex, nil
}

func (tm *TextureManager) Unload(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tex, ok := tm.textures[name]
	if !ok {
		return fmt.Errorf("texture '%s' not found", name)
	}

	tex.Delete()
	delete(tm.textures, name)
	return nil
}

func (tm *TextureManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for name, tex := range tm.textures {
		tex.Delete()
		delete(tm.textures, name)
	}
}

func (tm *TextureManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.textures)
}

func (tm *TextureManager) List() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	names := make([]string, 0, len(tm.textures))
	for name := range tm.textures {
		names = append(names, name)
	}
	return names
}
