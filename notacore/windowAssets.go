package notacore

import (
	"NotaborEngine/notaentity"
	"NotaborEngine/notageometry"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"fmt"
	"path/filepath"
)

type VisualMask uint8

const (
	MaskNone VisualMask = iota
	MaskCircle
)

type VisualOptions struct {
	Width  float32
	Height float32

	Material *notashader.Material

	Mask         VisualMask
	CircleRadius float32
	CircleEdge   float32
}

// SpriteOptions creates a rectangular visual configuration with no masking effects.
func SpriteOptions(width, height float32) VisualOptions {
	return VisualOptions{
		Width:  width,
		Height: height,
	}
}

// CircleSpriteOptions creates a rectangular sprite setup that is visually masked into a circle.
func CircleSpriteOptions(radius float32) VisualOptions {
	return VisualOptions{
		Width:        radius * 2,
		Height:       radius * 2,
		Mask:         MaskCircle,
		CircleRadius: 0.5,
		CircleEdge:   0.01,
	}
}

// LoadTexture loads a texture into the window's texture manager and uploads it to GL.
func (w *Window) LoadTexture(name, path string) (*notatexture.Texture, error) {
	w.MakeContextCurrent()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	tex, err := w.RunTime.TextureMgr.Load(name, absPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture: %w", err)
	}

	return tex, nil
}

func (w *Window) GetTexture(name string) (*notatexture.Texture, error) {
	return w.RunTime.TextureMgr.Get(name)
}

// UnloadTexture removes a texture from the window's texture manager.
func (w *Window) UnloadTexture(name string) error {
	return w.RunTime.TextureMgr.Unload(name)
}

// LoadShader loads or retrieves a shader program from the window's shader manager.
func (w *Window) LoadShader(name, vertexPath, fragmentPath string) (*notashader.Shader, error) {
	w.MakeContextCurrent()
	return w.RunTime.ShaderMgr.Load(name, vertexPath, fragmentPath)
}

// GetShader returns a previously loaded shader by name.
func (w *Window) GetShader(name string) (*notashader.Shader, error) {
	return w.RunTime.ShaderMgr.Get(name)
}

// ReloadShader recompiles a named shader in place.
func (w *Window) ReloadShader(name string) error {
	w.MakeContextCurrent()
	return w.RunTime.ShaderMgr.Reload(name)
}

// UnloadShader deletes a previously loaded shader from the window's shader manager.
func (w *Window) UnloadShader(name string) error {
	w.MakeContextCurrent()
	return w.RunTime.ShaderMgr.Unload(name)
}

// CreateMaterial creates a material instance bound to the provided shader.
func (w *Window) CreateMaterial(shader *notashader.Shader) *notashader.Material {
	return notashader.NewMaterial(shader)
}

// LoadMaterial loads a shader and wraps it in a material instance for draw-time customization.
func (w *Window) LoadMaterial(name, vertexPath, fragmentPath string) (*notashader.Material, error) {
	shader, err := w.LoadShader(name, vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}
	return notashader.NewMaterial(shader), nil
}

// BasicMaterial returns a material built from the engine's default 2D shader pair.
func (w *Window) BasicMaterial() (*notashader.Material, error) {
	return w.LoadMaterial("default-basic", notashader.DefaultVertexShaderPath, notashader.DefaultFragmentShaderPath)
}

// CircleMaterial returns a default material configured with the engine's built-in circle mask effect.
func (w *Window) CircleMaterial(radius, edge float32) (*notashader.Material, error) {
	material, err := w.BasicMaterial()
	if err != nil {
		return nil, err
	}
	return material.CircleMask(radius, edge), nil
}

// CreateSprite creates a sprite from an already loaded texture and geometry polygon.
func (w *Window) CreateSprite(name string, texture *notatexture.Texture, polygon *notageometry.Polygon) (*notatexture.Sprite, error) {
	sprite, err := w.RunTime.SpriteMgr.Create(name, texture)
	if err != nil {
		return nil, err
	}
	sprite.Polygon = polygon
	return sprite, nil
}

// LoadSprite loads a texture and immediately creates a sprite using the supplied polygon.
func (w *Window) LoadSprite(name, texturePath string, polygon *notageometry.Polygon) (*notatexture.Sprite, error) {
	texture, err := w.LoadTexture(name, texturePath)
	if err != nil {
		return nil, err
	}
	return w.CreateSprite(name, texture, polygon)
}

// CreateVisual creates a reusable visual bundle from a texture plus visual options.
func (w *Window) CreateVisual(name string, texture *notatexture.Texture, opts VisualOptions) (*notaentity.Visual, error) {
	polygon := notageometry.CreateRectangle(opts.Width, opts.Height)
	sprite, err := w.CreateSprite(name, texture, polygon)
	if err != nil {
		return nil, err
	}

	material, err := w.materialFromOptions(opts)
	if err != nil {
		return nil, err
	}

	return notaentity.NewVisual(sprite, material), nil
}

// LoadVisual loads a texture and creates a reusable visual bundle from the supplied options.
func (w *Window) LoadVisual(name, texturePath string, opts VisualOptions) (*notaentity.Visual, error) {
	texture, err := w.LoadTexture(name, texturePath)
	if err != nil {
		return nil, err
	}
	return w.CreateVisual(name, texture, opts)
}

// LoadCircleSprite is a compatibility helper that loads a circle-masked sprite and its material together.
func (w *Window) LoadCircleSprite(name, texturePath string, radius float32) (*notatexture.Sprite, *notashader.Material, error) {
	visual, err := w.LoadVisual(name, texturePath, CircleSpriteOptions(radius))
	if err != nil {
		return nil, nil, err
	}
	return visual.Sprite, visual.Material, nil
}

// GetSprite returns a previously created sprite by name.
func (w *Window) GetSprite(name string) (*notatexture.Sprite, error) {
	return w.RunTime.SpriteMgr.Get(name)
}

// RemoveSprite removes a previously created sprite from the window's sprite manager.
func (w *Window) RemoveSprite(name string) error {
	return w.RunTime.SpriteMgr.Remove(name)
}

func (w *Window) materialFromOptions(opts VisualOptions) (*notashader.Material, error) {
	var material *notashader.Material
	if opts.Material != nil {
		material = opts.Material
	} else {
		var err error
		material, err = w.BasicMaterial()
		if err != nil {
			return nil, err
		}
	}

	switch opts.Mask {
	case MaskNone:
		return material, nil
	case MaskCircle:
		if opts.CircleRadius == 0 {
			opts.CircleRadius = 0.5
		}
		if opts.CircleEdge == 0 {
			opts.CircleEdge = 0.01
		}
		return material.Clone().CircleMask(opts.CircleRadius, opts.CircleEdge), nil
	default:
		return nil, fmt.Errorf("unsupported visual mask")
	}
}
