package notacore

import (
	"NotaborEngine/notaobject"
	"fmt"
	"path/filepath"
)

// LoadTexture loads a texture and creates OpenGL texture (with context)
func (w *Window) LoadTexture(name, path string) (*notaobject.Texture, error) {
	w.MakeContextCurrent()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	tex, err := w.RunTime.TextureMgr.Load(name, absPath, true) // true = create GL texture
	if err != nil {
		return nil, fmt.Errorf("failed to load texture: %w", err)
	}

	return tex, nil
}

// GetTexture retrieves a loaded texture
func (w *Window) GetTexture(name string) (*notaobject.Texture, error) {
	return w.RunTime.TextureMgr.Get(name)
}

// UnloadTexture removes a texture
func (w *Window) UnloadTexture(name string) error {
	return w.RunTime.TextureMgr.Unload(name)
}
