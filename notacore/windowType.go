package notacore

import (
	"errors"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type WindowType int

const (
	Windowed WindowType = iota
	Fullscreen
	Borderless
)

func (w *Window) SetWindowType(t WindowType) error {
	return setWindowType(w.Config, w.Handle, t)
}

func setWindowType(config WindowConfig, handle *glfw.Window, t WindowType) error {
	switch t {
	case Fullscreen:
		monitor := glfw.GetPrimaryMonitor()
		if monitor == nil {
			return errors.New("no primary monitor found")
		}

		videoMode := monitor.GetVideoMode()
		if videoMode == nil {
			return errors.New("could not get video mode")
		}
		handle.SetMonitor(monitor, 0, 0, videoMode.Width, videoMode.Height, videoMode.RefreshRate)
		updateViewport(videoMode.Width, videoMode.Height, config.W, config.H)
		config.Type = Fullscreen
	case Borderless:
		handle.SetAttrib(glfw.Decorated, glfw.False)
		config.Type = Borderless
	case Windowed:
		handle.SetMonitor(nil, config.X, config.Y, config.W, config.H, glfw.DontCare)
		handle.SetAttrib(glfw.Decorated, glfw.True)
		updateViewport(config.W, config.H, config.W, config.H)
		config.Type = Windowed
	default:
		return errors.New("invalid window type")
	}
	return nil
}

func updateViewport(winW, winH, targetW, targetH int) {
	targetAspect := float32(targetW) / float32(targetH)
	windowAspect := float32(winW) / float32(winH)

	var viewW, viewH int32
	var viewX, viewY int32

	if windowAspect > targetAspect {
		// Window is wider than target aspect ratio
		viewH = int32(winH)
		viewW = int32(float32(winH) * targetAspect)
		viewX = (int32(winW) - viewW) / 2
		viewY = 0
	} else {
		// Window is taller than target aspect ratio
		viewW = int32(winW)
		viewH = int32(float32(winW) / targetAspect)
		viewX = 0
		viewY = (int32(winH) - viewH) / 2
	}

	gl.Viewport(viewX, viewY, viewW, viewH)
	// Scissor ensures that glClear only clears the viewport area if needed,
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor(viewX, viewY, viewW, viewH)
}
