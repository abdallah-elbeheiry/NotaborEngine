package notaobject

import (
	"fmt"
	"sync"
)

type Sprite struct {
	Texture  *Texture
	Name     string
	Polygon  *Polygon // reusable quad
	Material *Material
}

type SpriteManager struct {
	sprites  map[string]*Sprite
	textures *TextureManager
	mu       sync.RWMutex
}

func NewSpriteManager(textureManager *TextureManager) *SpriteManager {
	return &SpriteManager{
		sprites:  make(map[string]*Sprite),
		textures: textureManager,
	}
}

// Create creates a new sprite from a loaded texture
func (sm *SpriteManager) Create(name string, texture *Texture) (*Sprite, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sprites[name]; exists {
		return nil, fmt.Errorf("sprite '%s' already exists", name)
	}

	sprite := &Sprite{
		Texture: texture,
		Name:    name,
	}

	sm.sprites[name] = sprite
	return sprite, nil
}

// LoadAndCreate loads a texture and creates a sprite from it
func (sm *SpriteManager) LoadAndCreate(spriteName, texturePath string) (*Sprite, error) {
	texture, err := sm.textures.Load(spriteName, texturePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture: %w", err)
	}

	return sm.Create(spriteName, texture)
}

// Get retrieves a sprite by name
func (sm *SpriteManager) Get(name string) (*Sprite, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sprite, exists := sm.sprites[name]
	if !exists {
		return nil, fmt.Errorf("sprite '%s' not found", name)
	}

	return sprite, nil
}

// Remove removes a sprite (does not delete the texture)
func (sm *SpriteManager) Remove(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sprites[name]; !exists {
		return fmt.Errorf("sprite '%s' not found", name)
	}

	delete(sm.sprites, name)
	return nil
}

// Clear removes all sprites (does not delete textures)
func (sm *SpriteManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sprites = make(map[string]*Sprite)
}

// Count returns the number of sprites
func (sm *SpriteManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sprites)
}
