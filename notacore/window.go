package notacore

import (
	"NotaborEngine/notagl"
	"NotaborEngine/notashader"
	"errors"
	"sync"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type WindowConfig struct {
	X, Y       int
	W, H       int
	Title      string
	Resizable  bool
	Type       WindowType
	LogicLoops []*FixedHzLoop
	RenderLoop *RenderLoop
}

type WindowBaseRuntime struct {
	lastRender time.Time
	targetDt   time.Duration
}

type windowRunTime struct {
	WindowBaseRuntime
	backend    *notagl.GLBackend
	Renderer   *notagl.Renderer
	TextureMgr *notagl.TextureManager
}

type Window struct {
	ID      int
	Handle  *glfw.Window
	Config  WindowConfig
	RunTime windowRunTime
	Shaders map[string]*notashader.Shader
}

func (w *Window) GetConfig() *WindowConfig       { return &w.Config }
func (w *Window) GetRuntime() *WindowBaseRuntime { return &w.RunTime.WindowBaseRuntime }
func (w *Window) RunRenderer() {
	w.RunTime.Renderer.Orders = w.RunTime.Renderer.Orders[:0]
	w.Config.RenderLoop.Render()
	w.RunTime.Renderer.Flush(w.RunTime.backend)
}
func (w *Window) GLFW() *glfw.Window { return w.Handle }

type windowManager struct {
	mu        sync.Mutex
	windows2D []*Window
	nextID    int
}

func (wm *windowManager) Create(cfg WindowConfig) (*Window, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if cfg.W <= 0 || cfg.H <= 0 {
		return nil, errors.New("invalid window size")
	}

	var monitor *glfw.Monitor
	switch cfg.Type {
	case Fullscreen:
		monitor = glfw.GetPrimaryMonitor()
		if monitor == nil {
			return nil, errors.New("no primary monitor found")
		}
		videoMode := monitor.GetVideoMode()
		if videoMode == nil {
			return nil, errors.New("could not get video mode")
		}

	case Borderless:
		monitor = glfw.GetPrimaryMonitor()
		if monitor == nil {
			return nil, errors.New("no primary monitor found")
		}
		videoMode := monitor.GetVideoMode()
		if videoMode == nil {
			return nil, errors.New("could not get video mode")
		}

	default: // Windowed
	}

	handle, err := setupWindow(cfg, monitor)
	if err != nil {
		return nil, err
	}

	win := &Window{
		ID:     wm.nextID,
		Handle: handle,
		Config: cfg,
		RunTime: windowRunTime{
			WindowBaseRuntime: WindowBaseRuntime{
				lastRender: time.Now(),
				targetDt:   time.Second / time.Duration(cfg.RenderLoop.MaxHz),
			},
			backend:    &notagl.GLBackend{},
			Renderer:   &notagl.Renderer{},
			TextureMgr: notagl.NewTextureManager(),
		},
	}

	win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return nil, err
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	win.RunTime.backend.Init()

	return win, nil
}

func (wm *windowManager) PollEvents() {
	glfw.PollEvents()
}

func (w *Window) MakeContextCurrent()  { w.Handle.MakeContextCurrent() }
func (w *Window) SwapBuffers()         { w.Handle.SwapBuffers() }
func (w *Window) ShouldClose() bool    { return w.Handle.ShouldClose() }
func (w *Window) Close()               { w.Handle.SetShouldClose(true) }
func (w *Window) Size() (int, int)     { return w.Handle.GetSize() }
func (w *Window) Position() (int, int) { return w.Handle.GetPos() }

func (wm *windowManager) Destroy2D(win *Window) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for i, w := range wm.windows2D {
		if w == win {
			w.Close()
			wm.windows2D = append(wm.windows2D[:i], wm.windows2D[i+1:]...)
			break
		}
	}
}

func setupWindow(cfg WindowConfig, monitor *glfw.Monitor) (*glfw.Window, error) {
	if cfg.Type == Borderless {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	} else {
		glfw.WindowHint(glfw.Decorated, glfw.True)
	}

	if cfg.Resizable {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	var handle *glfw.Window
	var err error

	if cfg.Type == Fullscreen {
		handle, err = glfw.CreateWindow(cfg.W, cfg.H, cfg.Title, monitor, nil)
	} else {
		handle, err = glfw.CreateWindow(cfg.W, cfg.H, cfg.Title, nil, nil)
	}

	if err != nil {
		glfw.DefaultWindowHints()
		return nil, err
	}

	glfw.DefaultWindowHints()

	if cfg.Type == Borderless || cfg.Type == Windowed {
		handle.SetPos(cfg.X, cfg.Y)
	}

	handle.MakeContextCurrent()
	handle.Show()
	return handle, nil
}
