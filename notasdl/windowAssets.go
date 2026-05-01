package notasdl

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

func SpriteOptions(width, height float32) VisualOptions {
	return VisualOptions{
		Width:  width,
		Height: height,
	}
}

func CircleSpriteOptions(radius float32) VisualOptions {
	return VisualOptions{
		Width:        radius * 2,
		Height:       radius * 2,
		Mask:         MaskCircle,
		CircleRadius: 0.5,
		CircleEdge:   0.01,
	}
}

func (w *Window) LoadTexture(name, path string) (*notatexture.Texture, error) {
	w.MakeCurrent()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return w.Runtime.TextureMgr.Load(name, absPath, true)
}

func (w *Window) GetTexture(name string) (*notatexture.Texture, error) {
	return w.Runtime.TextureMgr.Get(name)
}

func (w *Window) UnloadTexture(name string) error {
	return w.Runtime.TextureMgr.Unload(name)
}

func (w *Window) LoadShader(name, vertexPath, fragmentPath string) (*notashader.Shader, error) {
	w.MakeCurrent()
	return w.Runtime.ShaderMgr.Load(name, vertexPath, fragmentPath)
}

func (w *Window) GetShader(name string) (*notashader.Shader, error) {
	return w.Runtime.ShaderMgr.Get(name)
}

func (w *Window) ReloadShader(name string) error {
	w.MakeCurrent()
	return w.Runtime.ShaderMgr.Reload(name)
}

func (w *Window) UnloadShader(name string) error {
	w.MakeCurrent()
	return w.Runtime.ShaderMgr.Unload(name)
}

func (w *Window) CreateMaterial(shader *notashader.Shader) *notashader.Material {
	return notashader.NewMaterial(shader)
}

func (w *Window) LoadMaterial(name, vertexPath, fragmentPath string) (*notashader.Material, error) {
	shader, err := w.LoadShader(name, vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}
	return notashader.NewMaterial(shader), nil
}

func (w *Window) BasicMaterial() (*notashader.Material, error) {
	return w.LoadMaterial(
		"default-basic",
		notashader.DefaultVertexShaderPath,
		notashader.DefaultFragmentShaderPath,
	)
}

func (w *Window) CircleMaterial(radius, edge float32) (*notashader.Material, error) {
	m, err := w.BasicMaterial()
	if err != nil {
		return nil, err
	}
	return m.CircleMask(radius, edge), nil
}
func (w *Window) CreateSprite(name string, texture *notatexture.Texture, polygon *notageometry.Polygon) (*notatexture.Sprite, error) {
	sprite, err := w.Runtime.SpriteMgr.Create(name, texture)
	if err != nil {
		return nil, err
	}
	sprite.Polygon = polygon
	return sprite, nil
}

func (w *Window) LoadSprite(name, texturePath string, polygon *notageometry.Polygon) (*notatexture.Sprite, error) {
	tex, err := w.LoadTexture(name, texturePath)
	if err != nil {
		return nil, err
	}
	return w.CreateSprite(name, tex, polygon)
}

func (w *Window) GetSprite(name string) (*notatexture.Sprite, error) {
	return w.Runtime.SpriteMgr.Get(name)
}

func (w *Window) RemoveSprite(name string) error {
	return w.Runtime.SpriteMgr.Remove(name)
}
func (w *Window) CreateVisual(name string, texture *notatexture.Texture, opts VisualOptions) (*notaentity.Visual, error) {
	polygon := notageometry.CreateRectangle(opts.Width, opts.Height)

	sprite, err := w.CreateSprite(name, texture, polygon)
	if err != nil {
		return nil, err
	}

	mat, err := w.materialFromOptions(opts)
	if err != nil {
		return nil, err
	}

	return notaentity.NewVisual(sprite, mat), nil
}

func (w *Window) LoadVisual(name, texturePath string, opts VisualOptions) (*notaentity.Visual, error) {
	tex, err := w.LoadTexture(name, texturePath)
	if err != nil {
		return nil, err
	}
	return w.CreateVisual(name, tex, opts)
}

func (w *Window) LoadCircleSprite(name, texturePath string, radius float32) (*notatexture.Sprite, *notashader.Material, error) {
	v, err := w.LoadVisual(name, texturePath, CircleSpriteOptions(radius))
	if err != nil {
		return nil, nil, err
	}
	return v.Sprite, v.Material, nil
}

func (w *Window) materialFromOptions(opts VisualOptions) (*notashader.Material, error) {
	var mat *notashader.Material

	if opts.Material != nil {
		mat = opts.Material
	} else {
		var err error
		mat, err = w.BasicMaterial()
		if err != nil {
			return nil, err
		}
	}

	switch opts.Mask {

	case MaskNone:
		return mat, nil

	case MaskCircle:
		r := opts.CircleRadius
		if r == 0 {
			r = 0.5
		}

		e := opts.CircleEdge
		if e == 0 {
			e = 0.01
		}

		return mat.Clone().CircleMask(r, e), nil

	default:
		return nil, fmt.Errorf("unsupported visual mask")
	}
}
