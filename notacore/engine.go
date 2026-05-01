package notacore

import (
	"NotaborEngine/notaentity"
	"NotaborEngine/notasdl"
	"NotaborEngine/notasound"
	"NotaborEngine/notatask"
	"runtime"
	"time"
)

type Settings struct {
	Vsync      bool    //Locks TargetFPS to monitor's refresh rate
	Muted      bool    //Completely disables sound
	SoundLevel float32 //Volume of sound (1 = 100%)
}

type Engine struct {
	Windows       []*notasdl.Window
	Platform      notasdl.Platform
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
		e.inputLoop.Start()
	}

	// Start all logic loops
	for _, w := range e.Windows {
		cfg := w.GetConfig()
		for _, loop := range cfg.Loops {
			loop.Start()
		}
		w.Runtime.LastFrame = time.Now()
	}

	for e.running && !e.AllWindowsClosed() {
		e.Platform.PollEvents()

		//if e.InputManager != nil {
		//	e.InputManager.captureInputs(e.Windows)
		//}

		now := time.Now()

		for _, win := range e.Windows {
			if win.ShouldClose {
				continue
			}

			rt := win.Runtime
			elapsed := now.Sub(rt.LastFrame)
			if elapsed < rt.TargetDt {
				time.Sleep(rt.TargetDt - elapsed)
				continue
			}

			win.MakeCurrent()
			win.RenderFrame()
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

	e.inputLoop.Do(func() {
		if e.InputManager == nil {
			panic("InputManager is not initialized, initialize it first")
		}
		e.InputManager.tick()
	})
}

// AllWindowsClosed returns true if all windows are closed
func (e *Engine) AllWindowsClosed() bool {
	for _, w := range e.Windows {
		if !w.ShouldClose {
			return false
		}
	}
	return true
}

// Shutdown shuts down the engine and releases all resources
func (e *Engine) Shutdown() {
	if e.Platform != nil {
		e.Platform.Shutdown()
	}
}

func (e *Engine) initPlatform() error {
	runtime.LockOSThread()

	p := &notasdl.WindowManager{}
	if err := p.Init(); err != nil {
		return err
	}

	e.Platform = p

	e.InputManager = &InputManager{}
	e.EntityManager = notaentity.NewEntityManager()

	return nil
}

// CreateWindow creates a new window with the given configuration
func (e *Engine) CreateWindow(cfg *notasdl.WindowConfig) (*notasdl.Window, error) {
	win, err := e.Platform.CreateWindow(cfg)
	if err != nil {
		return nil, err
	}

	win.MakeCurrent()
	win.Runtime.Backend.Init()

	if e.settings.Vsync {
		win.SetVSync(true)
	} else {
		win.SetVSync(false)
	}

	e.Windows = append(e.Windows, win)
	return win, nil
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
		Windows:      []*notasdl.Window{},
		settings:     settings,
		Platform:     nil,
		SoundManager: audio,
	}

	e.ChangeSettings(settings)
	return e, e.initPlatform()
}

// ChangeSettings changes the engine's settings and updates all engine components appropriately
func (e *Engine) ChangeSettings(settings *Settings) {
	e.settings = settings
	e.SoundManager.Mute = settings.Muted
	e.SoundManager.MasterVolume = settings.SoundLevel
}

// GetSettings returns the engine's settings as a copy
func (e *Engine) GetSettings() Settings {
	return *e.settings
}
