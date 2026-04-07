package notacore

import (
	"NotaborEngine/notaentity"
	"NotaborEngine/notasound"
	"NotaborEngine/notatask"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Settings struct {
	Vsync      bool    //Locks TargetFPS to monitor's refresh rate
	Muted      bool    //Completely disables sound
	SoundLevel float32 //Volume of sound (1 = 100%)
}

type Engine struct {
	Windows       []*Window
	WindowManager *windowManager
	InputManager  *InputManager
	SoundManager  *notasound.SoundManager
	EntityManager *notaentity.EntityManager

	settings  *Settings
	inputLoop *notatask.Loop
	running   bool
}

// Run activates all loops within all windows, it's recommended to use at the end of the engine configuration
func (e *Engine) Run() error {
	e.running = true

	if e.inputLoop != nil {
		e.inputLoop.Start(e.EntityManager)
	}
	// Start all logic loops
	for _, w := range e.Windows {
		cfg := w.GetConfig()
		for _, loop := range cfg.Loops {
			loop.Start(e.EntityManager)
		}
		w.GetRuntime().lastRender = time.Now()
	}

	for e.running && !e.AllWindowsClosed() {
		e.WindowManager.PollEvents()

		if e.InputManager != nil {
			e.InputManager.captureInputs(e.Windows)
		}

		now := time.Now()

		for _, win := range e.Windows {
			if win.ShouldClose() {
				continue
			}

			rt := win.GetRuntime()
			elapsed := now.Sub(rt.lastRender)
			if elapsed < rt.targetDt {
				time.Sleep(rt.targetDt - elapsed)
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
		for _, loop := range w.GetConfig().Loops {
			loop.Stop()
		}
	}
	return nil
}

// SetInputFrequency sets the frequency of the input loop, which is responsible for hardware input detection
func (e *Engine) SetInputFrequency(Hz float32) {
	e.inputLoop = notatask.CreateLoop(Hz)
	e.InputManager.loop = e.inputLoop
	e.inputLoop.Add(notatask.CreateTask(func() error {
		if e.InputManager == nil {
			return errors.New("InputManager is not initialized, initialize it first")
		}
		e.InputManager.tick()
		return nil
	}))
}

// AllWindowsClosed returns true if all windows are closed
func (e *Engine) AllWindowsClosed() bool {
	for _, w := range e.Windows {
		if !w.ShouldClose() {
			return false
		}
	}
	return true
}

// Shutdown shuts down the engine and releases all resources
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
	e.EntityManager = notaentity.NewEntityManager()
	return nil
}

// CreateWindow creates a new window with the given configuration
func (e *Engine) CreateWindow(cfg WindowConfig) (*Window, error) {
	win, err := e.WindowManager.Create(cfg)
	if err != nil {
		return nil, err
	}
	win.MakeContextCurrent()
	win.RunTime.backend.Init()
	if e.settings.Vsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
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

// CreateEngine creates a new engine instance with the given settings
// The engine will be initialized with a sound manager, window manager, entity manager, and input manager ready to use
// The engine automatically detects the OS and provides the ideal configuration files
func CreateEngine(settings *Settings) (*Engine, error) {
	audio, err := notasound.NewSoundManager()
	if err != nil {
		return nil, err
	}

	e := &Engine{
		Windows:       []*Window{},
		settings:      settings,
		WindowManager: &windowManager{},
		SoundManager:  audio,
	}
	e.ChangeSettings(settings)
	return e, e.initPlatform()
}

// ChangeSettings changes the engine's settings and updates all engine components appropriately
func (e *Engine) ChangeSettings(settings *Settings) {
	e.settings = settings
	e.SoundManager.Mute = settings.Muted
	e.SoundManager.MasterVolume = settings.SoundLevel

	if len(e.Windows) > 0 {
		if settings.Vsync {
			glfw.SwapInterval(1)
		} else {
			glfw.SwapInterval(0)
		}
	}
}

// GetSettings returns the engine's settings as a copy
func (e *Engine) GetSettings() Settings {
	return *e.settings
}
