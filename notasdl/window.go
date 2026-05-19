package notasdl

import (
	"NotaborEngine/notaentity"
	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
	"NotaborEngine/notatask"
	"NotaborEngine/notatexture"
	"sync"
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
	Hidden      bool
	Minimized   bool
	Occluded    bool

	positionMu         sync.Mutex
	pendingSetPosition bool
	pendingX           int
	pendingY           int
	pendingMoveX       int
	pendingMoveY       int
}

func (w *Window) RenderFrame() {
	now := time.Now()
	dt := float32(now.Sub(w.Runtime.LastFrame).Seconds())
	w.Runtime.LastFrame = now

	if !w.canRender() {
		w.Runtime.Renderer.Clear()
		return
	}

	w.applyPendingWindowPosition()

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

func (w *Window) canRender() bool {
	return !w.ShouldClose &&
		!w.Hidden &&
		!w.Minimized &&
		!w.Occluded &&
		w.Config.W > 0 &&
		w.Config.H > 0
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

	w.positionMu.Lock()
	w.pendingSetPosition = true
	w.pendingX = x
	w.pendingY = y
	w.pendingMoveX = 0
	w.pendingMoveY = 0
	w.positionMu.Unlock()
}

func (w *Window) applyPendingWindowPosition() {
	w.positionMu.Lock()
	if !w.pendingSetPosition && w.pendingMoveX == 0 && w.pendingMoveY == 0 {
		w.positionMu.Unlock()
		return
	}

	x := w.Config.X
	y := w.Config.Y
	if w.pendingSetPosition {
		x = w.pendingX
		y = w.pendingY
	}
	x += w.pendingMoveX
	y += w.pendingMoveY

	w.pendingSetPosition = false
	w.pendingMoveX = 0
	w.pendingMoveY = 0
	w.positionMu.Unlock()

	_ = w.setPositionNow(x, y)
}

func (w *Window) setPositionNow(x, y int) error {
	if w.Handle == nil {
		return nil
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

	if err := w.Handle.SetPosition(int32(x), int32(y)); err != nil {
		return err
	}

	w.setCachedPosition(x, y)
	return nil
}

func (w *Window) setCachedPosition(x, y int) {
	w.positionMu.Lock()
	defer w.positionMu.Unlock()
	w.Config.X = x
	w.Config.Y = y
}

// Move moves the window by delta with boundary checking
func (w *Window) Move(dx, dy int) {
	if w.Handle == nil {
		return
	}

	w.positionMu.Lock()
	w.pendingMoveX += dx
	w.pendingMoveY += dy
	w.positionMu.Unlock()
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
	if w.Handle == nil {
		return w.Config.X, w.Config.Y
	}

	w.positionMu.Lock()
	defer w.positionMu.Unlock()

	x = w.Config.X
	y = w.Config.Y
	if w.pendingSetPosition {
		x = w.pendingX
		y = w.pendingY
	}

	return x + w.pendingMoveX, y + w.pendingMoveY
}
