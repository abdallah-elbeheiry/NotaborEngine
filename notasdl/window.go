package notasdl

import (
	"NotaborEngine/notacore"
	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
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
}

type WindowRuntime struct {
	lastFrame time.Time
	targetDt  time.Duration

	GLContext sdl.GLContext

	Backend    *notarender.GLBackend
	Renderer   *notarender.Renderer
	TextureMgr *notatexture.TextureManager
	SpriteMgr  *notatexture.SpriteManager
	ShaderMgr  *notashader.Manager

	Cameras []*notacore.Camera2D
}

type Window struct {
	ID            WindowID
	Handle        *sdl.Window
	Config        *WindowConfig
	Runtime       *WindowRuntime
	DefaultCamera *notacore.Camera2D

	ShouldClose bool
}

func (w *Window) RenderFrame() {
	now := time.Now()
	dt := float32(now.Sub(w.Runtime.lastFrame).Seconds())
	w.Runtime.lastFrame = now

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
