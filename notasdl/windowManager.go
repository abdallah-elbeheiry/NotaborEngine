package notasdl

import (
	"NotaborEngine/notacore"
	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"errors"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
)

type WindowID uint32
type WindowManager struct {
	windows map[WindowID]*Window
	currId  WindowID
	freeIds []WindowID
}

func (wm *WindowManager) Create(cfg *WindowConfig) (*Window, error) {
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

	//Runtime creation step
	rt := WindowRuntime{
		lastFrame:  time.Now(),
		targetDt:   time.Duration(float64(time.Second) / float64(cfg.TargetFPS)),
		GLContext:  ctx,
		Backend:    &notarender.GLBackend{},
		Renderer:   &notarender.Renderer{},
		TextureMgr: notatexture.NewTextureManager(),
		ShaderMgr:  notashader.NewManager(),
	}
	rt.SpriteMgr = notatexture.NewSpriteManager(rt.TextureMgr)

	defaultCam := notacore.NewCamera2D()
	rt.Cameras = []*notacore.Camera2D{defaultCam}

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

	return w, nil
}
