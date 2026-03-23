package notacore

import (
	"NotaborEngine/notasound"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Settings struct {
	Vsync      bool
	Muted      bool
	SoundLevel float32
}

type Engine struct {
	Windows       []*Window
	Settings      *Settings
	WindowManager *windowManager
	InputManager  *InputManager
	SoundManager  *notasound.SoundManager

	inputLoop *FixedHzLoop
	running   bool
}

func (e *Engine) Run() error {
	e.running = true

	if e.inputLoop != nil {
		e.inputLoop.Start()
	}
	// Start all logic loops
	for _, w := range e.Windows {
		cfg := w.GetConfig()
		for _, loop := range cfg.LogicLoops {
			loop.Start()
		}
		w.GetRuntime().lastRender = time.Now()
	}

	for e.running && !e.AllWindowsClosed() {
		e.WindowManager.PollEvents()

		if e.InputManager != nil {
			e.InputManager.CaptureInputs(e.Windows)
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

func (e *Engine) SetInputFrequency(Hz float32) {
	e.inputLoop = CreateFixedHzLoop(Hz)
	e.inputLoop.Add(func() error {
		if e.InputManager == nil {
			return errors.New("InputManager is not initialized, initialize it first")
		}
		e.InputManager.Tick()
		return nil
	})
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

func (e *Engine) initPlatform() error {
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
		windows2D: []*Window{},
		nextID:    0,
	}

	e.WindowManager = wm
	e.InputManager = &InputManager{}
	return nil
}

func (e *Engine) CreateWindow(cfg WindowConfig) (*Window, error) {
	win, err := e.WindowManager.Create(cfg)
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

func CreateEngine(settings *Settings) (*Engine, error) {
	audio, err := notasound.NewSoundManager()
	if err != nil {
		return nil, err
	}

	e := &Engine{
		Windows:       []*Window{},
		Settings:      settings,
		WindowManager: &windowManager{},
		SoundManager:  audio,
	}
	return e, e.initPlatform()
}
