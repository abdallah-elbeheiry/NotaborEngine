package notasdl

import (
	"NotaborEngine/notaentity"
	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
	"NotaborEngine/notatask"
	"NotaborEngine/notatexture"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
)

type WindowType int

const (
	Windowed WindowType = iota
	Fullscreen
	Borderless
)

type WindowConfig struct {
	X, Y      int
	W, H      int
	Title     string
	Resizable bool
	Type      WindowType

	TargetFPS float32
	Loops     []*notatask.Loop
}

type WindowRuntime struct {
	LastFrame time.Time
	TargetDt  time.Duration

	GLContext sdl.GLContext

	RenderLoop *notatask.Loop

	Backend    *notarender.GLBackend
	Renderer   *notarender.Renderer
	TextureMgr *notatexture.TextureManager
	SpriteMgr  *notatexture.SpriteManager
	ShaderMgr  *notashader.Manager

	Cameras []*Camera2D
}

type Window struct {
	ID            WindowID
	Handle        *sdl.Window
	Config        *WindowConfig
	Runtime       *WindowRuntime
	DefaultCamera *Camera2D

	ShouldClose bool
}

func (w *Window) RenderFrame() {
	now := time.Now()
	dt := float32(now.Sub(w.Runtime.LastFrame).Seconds())
	w.Runtime.LastFrame = now

	for _, cam := range w.Runtime.Cameras {
		cam.Update(dt)
	}

	w.Runtime.Renderer.FrameID.Inc()
	w.Runtime.Renderer.Flush(w.Runtime.Backend)

	_ = sdl.GL_SwapWindow(w.Handle)
}

func (w *Window) MakeCurrent() {
	_ = sdl.GL_MakeCurrent(w.Handle, w.Runtime.GLContext)
}

func (w *Window) GetConfig() *WindowConfig {
	return w.Config
}

func (w *Window) SetVSync(enabled bool) {
	if enabled {
		_ = sdl.GL_SetSwapInterval(1)
	} else {
		_ = sdl.GL_SetSwapInterval(0)
	}
}

// Draw queues entities for rendering using SDL-backed renderer.
// If cam is nil, default camera is used.
func (w *Window) Draw(alpha float32, cam *Camera2D, entities ...*notaentity.Entity) error {
	if cam == nil {
		cam = w.DefaultCamera
	}

	view := cam.ViewMatrix()

	for _, e := range entities {
		if e == nil {
			continue
		}
		if err := e.DrawWithView(w.Runtime.Renderer, view, alpha); err != nil {
			return err
		}
	}

	return nil
}
