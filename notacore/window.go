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

type Window interface {
	GetConfig() *WindowConfig
	GetRuntime() *WindowBaseRuntime
	MakeContextCurrent()
	SwapBuffers()
	ShouldClose() bool
	RunRenderer()
	SetWindowType(t WindowType) error
	LoadTexture(name, path string) (*notagl.Texture, error)
	GetTexture(name string) (*notagl.Texture, error)
	UnloadTexture(name string) error

	GLFW() *glfw.Window
}

type WindowBaseRuntime struct {
	lastRender time.Time
	targetDt   time.Duration
}

type windowRunTime2D struct {
	WindowBaseRuntime
	backend    *notagl.GLBackend2D
	Renderer   *notagl.Renderer2D
	TextureMgr *notagl.TextureManager
}

type windowRuntime3D struct {
	WindowBaseRuntime
	backend    *notagl.GLBackend3D
	Renderer   *notagl.Renderer3D
	TextureMgr *notagl.TextureManager
}

type GlfwWindow2D struct {
	ID      int
	Handle  *glfw.Window
	Config  WindowConfig
	RunTime windowRunTime2D
	Shaders map[string]*notashader.Shader
}

type GlfwWindow3D struct {
	ID      int
	Handle  *glfw.Window
	Config  WindowConfig
	RunTime windowRuntime3D
	Shaders map[string]*notashader.Shader
}

func (w *GlfwWindow2D) GetConfig() *WindowConfig       { return &w.Config }
func (w *GlfwWindow2D) GetRuntime() *WindowBaseRuntime { return &w.RunTime.WindowBaseRuntime }
func (w *GlfwWindow2D) RunRenderer() {
	w.RunTime.Renderer.Orders = w.RunTime.Renderer.Orders[:0]
	w.Config.RenderLoop.Render()
	w.RunTime.Renderer.Flush(w.RunTime.backend)
}
func (w *GlfwWindow2D) GLFW() *glfw.Window { return w.Handle }
func (w *GlfwWindow3D) GLFW() *glfw.Window { return w.Handle }

func (w *GlfwWindow3D) GetConfig() *WindowConfig       { return &w.Config }
func (w *GlfwWindow3D) GetRuntime() *WindowBaseRuntime { return &w.RunTime.WindowBaseRuntime }
func (w *GlfwWindow3D) RunRenderer() {
	w.RunTime.Renderer.Orders = w.RunTime.Renderer.Orders[:0]
	w.Config.RenderLoop.Render()
	w.RunTime.Renderer.Flush(w.RunTime.backend)
}

type windowManager struct {
	mu        sync.Mutex
	windows2D []*GlfwWindow2D
	windows3D []*GlfwWindow3D
	nextID    int
}

func (wm *windowManager) Create2D(cfg WindowConfig) (*GlfwWindow2D, error) {
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

	win := &GlfwWindow2D{
		ID:     wm.nextID,
		Handle: handle,
		Config: cfg,
		RunTime: windowRunTime2D{
			WindowBaseRuntime: WindowBaseRuntime{
				lastRender: time.Now(),
				targetDt:   time.Second / time.Duration(cfg.RenderLoop.MaxHz),
			},
			backend:    &notagl.GLBackend2D{},
			Renderer:   &notagl.Renderer2D{},
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

func (wm *windowManager) Create3D(cfg WindowConfig) (*GlfwWindow3D, error) {
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

	default:
	}

	handle, err := setupWindow(cfg, monitor)
	if err != nil {
		return nil, err
	}
	win := &GlfwWindow3D{
		ID:     wm.nextID,
		Handle: handle,
		Config: cfg,
		RunTime: windowRuntime3D{
			WindowBaseRuntime: WindowBaseRuntime{
				lastRender: time.Now(),
				targetDt:   time.Second / time.Duration(cfg.RenderLoop.MaxHz),
			},
			backend:    &notagl.GLBackend3D{},
			Renderer:   &notagl.Renderer3D{},
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

func (w *GlfwWindow2D) MakeContextCurrent()  { w.Handle.MakeContextCurrent() }
func (w *GlfwWindow2D) SwapBuffers()         { w.Handle.SwapBuffers() }
func (w *GlfwWindow2D) ShouldClose() bool    { return w.Handle.ShouldClose() }
func (w *GlfwWindow2D) Close()               { w.Handle.SetShouldClose(true) }
func (w *GlfwWindow2D) Size() (int, int)     { return w.Handle.GetSize() }
func (w *GlfwWindow2D) Position() (int, int) { return w.Handle.GetPos() }

func (w *GlfwWindow3D) MakeContextCurrent()  { w.Handle.MakeContextCurrent() }
func (w *GlfwWindow3D) SwapBuffers()         { w.Handle.SwapBuffers() }
func (w *GlfwWindow3D) ShouldClose() bool    { return w.Handle.ShouldClose() }
func (w *GlfwWindow3D) Close()               { w.Handle.SetShouldClose(true) }
func (w *GlfwWindow3D) Size() (int, int)     { return w.Handle.GetSize() }
func (w *GlfwWindow3D) Position() (int, int) { return w.Handle.GetPos() }

func (wm *windowManager) Destroy2D(win *GlfwWindow2D) {
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

func (wm *windowManager) Destroy3D(win *GlfwWindow3D) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	for i, w := range wm.windows3D {
		if w == win {
			w.Close()
			wm.windows3D = append(wm.windows3D[:i], wm.windows3D[i+1:]...)
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
