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

	RenderLoop *notatask.Loop

	Backend    *notarender.Backend
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

	// Acquire command buffer for this frame
	cmdBuf, err := w.Runtime.Backend.BeginFrame()
	if err != nil {
		return
	}

	// Flush render queue to GPU (this also submits the command buffer)
	if err := w.Runtime.Renderer.Flush(w.Runtime.Backend, cmdBuf, w.Handle); err != nil {
		return
	}
}

func (w *Window) MakeCurrent() {
	// SDL3 GPU doesn't require making a context current
	// This is a no-op for GPU rendering
}

func (w *Window) GetConfig() *WindowConfig {
	return w.Config
}

func (w *Window) SetVSync(enabled bool) {
	// SDL3 GPU handles vertical sync automatically
	// This is managed at the presentation level
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

func (w *Window) SetPosition(x, y int) {
	if w.Handle == nil {
		return
	}

	bounds := w.getCurrentDisplayBounds()

	minX := -w.Config.W + 50
	minY := -w.Config.H + 50
	maxX := int(bounds.W - 50)
	maxY := int(bounds.H - 50)

	if x < minX {
		x = minX
	}
	if x > maxX {
		x = maxX
	}
	if y < minY {
		y = minY
	}
	if y > maxY {
		y = maxY
	}

	_ = w.Handle.SetPosition(int32(x), int32(y))

	w.Config.X = x
	w.Config.Y = y
}

// Move moves the window by delta with boundary checking
func (w *Window) Move(dx, dy int) {
	if w.Handle == nil {
		return
	}
	x, y := w.GetPosition()
	w.SetPosition(x+dx, y+dy)
}

// Helper to get current display bounds

func (w *Window) getCurrentDisplayBounds() *sdl.Rect {
	if w.Handle == nil {
		panic("Window handle is nil, initialize window before calling getCurrentDisplayBounds")
	}

	bounds, err := sdl.GetDisplayForWindow(w.Handle).Bounds()
	if err != nil {
		panic(err)
	}

	return bounds
}

// GetPosition returns current window position
func (w *Window) GetPosition() (x, y int) {
	if w.Handle != nil {
		px, py, _ := w.Handle.Position()
		return int(px), int(py)
	}
	return w.Config.X, w.Config.Y
}
