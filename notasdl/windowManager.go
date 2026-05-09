package notasdl

import (
	"errors"
	"time"

	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
	"NotaborEngine/notatask"
	"NotaborEngine/notatexture"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
)

type WindowID uint32
type WindowManager struct {
	windows       map[WindowID]*Window
	currId        WindowID
	eventHandlers []EventHandler
	gamepads      map[sdl.JoystickID]*sdl.Gamepad
}

func (wm *WindowManager) Init() error {
	binsdl.Load()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMEPAD); err != nil {
		return err
	}

	if err := sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4); err != nil {
		return err
	}
	if err := sdl.GL_SetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 6); err != nil {
		return err
	}
	if err := sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE); err != nil {
		return err
	}
	if err := sdl.GL_SetAttribute(sdl.GL_DOUBLEBUFFER, 1); err != nil {
		return err
	}

	wm.windows = make(map[WindowID]*Window)
	wm.gamepads = make(map[sdl.JoystickID]*sdl.Gamepad)
	sdl.SetGamepadEventsEnabled(true)
	return nil
}

func (wm *WindowManager) Shutdown() {
	for _, window := range wm.windows {
		window.ShouldClose = true
	}
	for _, gamepad := range wm.gamepads {
		if gamepad != nil {
			gamepad.Close()
		}
	}

	sdl.Quit()
}

func (wm *WindowManager) CreateWindow(cfg *WindowConfig) (*Window, error) {
	// Config validation step
	if cfg.W <= 0 || cfg.H <= 0 {
		return nil, errors.New("invalid window size")
	}
	if cfg.TargetFPS < 0 {
		return nil, errors.New("invalid target FPS")
	}
	flags := sdl.WINDOW_OPENGL

	// Set window flags from config step
	if cfg.Resizable {
		flags |= sdl.WINDOW_RESIZABLE
	}

	switch cfg.Type {
	case Fullscreen:
		flags |= sdl.WINDOW_FULLSCREEN
		break
	case Windowed:
		break
	case Borderless:
		flags |= sdl.WINDOW_BORDERLESS
	}

	//Window creation step
	win, err := sdl.CreateWindow(cfg.Title, cfg.W, cfg.H, flags)
	if err != nil {
		return nil, err
	}
	err = win.SetPosition(int32(cfg.X), int32(cfg.Y))
	if err != nil {
		return nil, err
	}

	ctx, err := sdl.GL_CreateContext(win)
	if err != nil {
		return nil, err
	}

	err = sdl.GL_MakeCurrent(win, ctx)
	if err != nil {
		return nil, err
	}

	// Initialization of OpenGL step
	if err := gl.Init(); err != nil {
		return nil, err
	}
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	//Runtime creation step
	rt := WindowRuntime{
		LastFrame:  time.Now(),
		TargetDt:   time.Duration(float64(time.Second) / float64(cfg.TargetFPS)),
		GLContext:  ctx,
		Backend:    &notarender.GLBackend{},
		Renderer:   &notarender.Renderer{},
		TextureMgr: notatexture.NewTextureManager(),
		ShaderMgr:  notashader.NewManager(),
	}
	rt.SpriteMgr = notatexture.NewSpriteManager(rt.TextureMgr)
	rt.RenderLoop = notatask.CreateLoop(cfg.TargetFPS)

	defaultCam := NewCamera2D()
	rt.Cameras = []*Camera2D{defaultCam}

	w := &Window{
		ID:            wm.currId,
		Handle:        win,
		Config:        cfg,
		Runtime:       &rt,
		DefaultCamera: defaultCam,
	}
	wm.currId++

	w.Runtime.Backend.Init()

	if wm.windows == nil {
		wm.windows = make(map[WindowID]*Window)
	}
	wm.windows[w.ID] = w

	rt.RenderLoop.Start()

	return w, nil
}

func (wm *WindowManager) PollEvents() {
	var ev sdl.Event

	for sdl.PollEvent(&ev) {
		switch ev.Type {

		case sdl.EVENT_WINDOW_CLOSE_REQUESTED:
			id, _ := ev.Window().ID()

			if w, ok := wm.windows[WindowID(id)]; ok {
				w.ShouldClose = true
			}
			wm.emit(Event{Type: EventWindowClose, WindowID: uint32(id)})

		case sdl.EVENT_WINDOW_RESIZED:
			id, _ := ev.Window().ID()
			if w, ok := wm.windows[WindowID(id)]; ok {
				w.Config.W = int(ev.WindowEvent().Data1)

				w.Config.H = int(ev.WindowEvent().Data2)
				w.MakeCurrent()
				gl.Viewport(0, 0, ev.WindowEvent().Data1, ev.WindowEvent().Data2)
			}

		case sdl.EVENT_QUIT:
			for _, w := range wm.windows {
				w.ShouldClose = true
			}
			wm.emit(Event{Type: EventQuit})

		case sdl.EVENT_KEY_DOWN, sdl.EVENT_KEY_UP:
			key := ev.KeyboardEvent()
			eventType := EventKeyUp
			if key.Down {
				eventType = EventKeyDown
			}
			wm.emit(Event{
				Type:     eventType,
				WindowID: uint32(key.WindowID),
				Key:      Key(key.Key),
			})

		case sdl.EVENT_MOUSE_BUTTON_DOWN, sdl.EVENT_MOUSE_BUTTON_UP:
			btn := ev.MouseButtonEvent()
			eventType := EventMouseUp
			if btn.Down {
				eventType = EventMouseDown
			}
			wm.emit(Event{
				Type:     eventType,
				WindowID: uint32(btn.WindowID),
				MouseBtn: MouseButton(btn.Button),
				X:        btn.X,
				Y:        btn.Y,
			})

		case sdl.EVENT_MOUSE_MOTION:
			motion := ev.MouseMotionEvent()
			wm.emit(Event{
				Type:     EventMouseMove,
				WindowID: uint32(motion.WindowID),
				X:        motion.X,
				Y:        motion.Y,
			})

		case sdl.EVENT_GAMEPAD_BUTTON_DOWN, sdl.EVENT_GAMEPAD_BUTTON_UP:
			btn := ev.GamepadButtonEvent()
			eventType := EventGamepadButtonUp
			if btn.Down {
				eventType = EventGamepadButtonDown
			}
			wm.emit(Event{
				Type:       eventType,
				GamepadBtn: GamepadButton(btn.Button),
			})

		case sdl.EVENT_GAMEPAD_AXIS_MOTION:
			axis := ev.GamepadAxisEvent()
			wm.emit(Event{
				Type:        EventGamepadAxisMotion,
				GamepadAxis: GamepadAxis(axis.Axis),
				AxisValue:   float32(axis.Value) / 32767,
			})

		case sdl.EVENT_GAMEPAD_ADDED:
			device := ev.GamepadDeviceEvent()
			if wm.gamepads == nil {
				wm.gamepads = make(map[sdl.JoystickID]*sdl.Gamepad)
			}
			gamepad, err := device.Which.OpenGamepad()
			if err == nil {
				wm.gamepads[device.Which] = gamepad
			}

		case sdl.EVENT_GAMEPAD_REMOVED:
			device := ev.GamepadDeviceEvent()
			if gamepad := wm.gamepads[device.Which]; gamepad != nil {
				gamepad.Close()
			}
			delete(wm.gamepads, device.Which)
		}
	}
}

func (wm *WindowManager) SubscribeEvents(handler EventHandler) {
	if handler == nil {
		return
	}
	wm.eventHandlers = append(wm.eventHandlers, handler)
}

func (wm *WindowManager) emit(event Event) {
	for _, handler := range wm.eventHandlers {
		handler(event)
	}
}

type Platform interface {
	Init() error
	PollEvents()
	SubscribeEvents(EventHandler)
	CreateWindow(cfg *WindowConfig) (*Window, error)
	Shutdown()
}
