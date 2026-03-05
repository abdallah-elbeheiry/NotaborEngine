package notacore

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Settings struct {
	Vsync bool
}

type Engine struct {
	Windows       []Window
	Settings      *Settings
	WindowManager *windowManager
	Input         *InputManager

	running bool
}

func (e *Engine) Run() error {
	e.running = true

	// Start all logic loops
	for _, w := range e.Windows {
		cfg := w.GetConfig()
		for _, loop := range cfg.LogicLoops {
			if e.Input != nil {
				update := func() error {
					e.Input.UpdateSignals()
					return nil
				}
				loop.activeRunnables = append([]Runnable{update}, loop.activeRunnables...)
			}
			loop.Start()
		}
		w.GetRuntime().lastRender = time.Now()
	}

	for e.running && !e.AllWindowsClosed() {
		e.WindowManager.PollEvents()

		if e.Input != nil {
			e.Input.CaptureInputs(e.Windows)
		}

		now := time.Now()

		for _, win := range e.Windows {
			if win.ShouldClose() {
				continue
			}

			rt := win.GetRuntime()
			elapsed := now.Sub(rt.lastRender)
			if elapsed < rt.targetDt {
				continue
			}
			rt.lastRender = now

			win.MakeContextCurrent()
			win.RunRenderer()
			win.SwapBuffers()
		}
	}

	// Stop logic loops
	for _, w := range e.Windows {
		for _, loop := range w.GetConfig().LogicLoops {
			loop.Stop()
		}
	}
	return nil
}

func (e *Engine) AllWindowsClosed() bool {
	for _, w := range e.Windows {
		if !w.ShouldClose() {
			return false
		}
	}
	return true
}

func (e *Engine) Shutdown() {
	glfw.Terminate()
}

func (e *Engine) InitPlatform() error {
	runtime.LockOSThread()

	if err := addNativeDLLPath(); err != nil {
		return err
	}

	if err := glfw.Init(); err != nil {
		return err
	}

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	wm := &windowManager{
		windows2D: []*GlfwWindow2D{},
		windows3D: []*GlfwWindow3D{},
		nextID:    0,
	}

	e.WindowManager = wm
	e.Input = &InputManager{}
	return nil
}

func (e *Engine) CreateWindow2D(cfg WindowConfig) (*GlfwWindow2D, error) {
	win, err := e.WindowManager.Create2D(cfg)
	if err != nil {
		return nil, err
	}
	win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return nil, err
	}
	win.RunTime.backend.Init()
	e.Windows = append(e.Windows, win)
	return win, nil
}

func (e *Engine) CreateWindow3D(cfg WindowConfig) (*GlfwWindow3D, error) {
	win, err := e.WindowManager.Create3D(cfg)
	if err != nil {
		return nil, err
	}
	win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return nil, err
	}
	win.RunTime.backend.Init()
	e.Windows = append(e.Windows, win)
	return win, nil
}

func addNativeDLLPath() error {
	switch runtime.GOOS {
	case "windows":
		exeDir, err := os.Getwd()
		if err != nil {
			return err
		}
		dllDir := filepath.Join(exeDir, "notacore", "native", "windows")
		err = os.Setenv("PATH", dllDir+";"+os.Getenv("PATH"))
		if err != nil {
			return err
		}

	case "linux":
		// set linux paths later

	case "darwin":
	// set mac paths later
	default:
		// return unsupported platform error
	}
	return nil
}
