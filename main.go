package main

import (
	"NotaborEngine/notasdl"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
)

func main() {
	defer binsdl.Load().Unload()
	defer sdl.Quit()
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 6)
	_ = sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	_ = sdl.GL_SetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	defer sdl.Quit()

	cfg := notasdl.WindowConfig{
		X:         50,
		Y:         50,
		W:         800,
		H:         600,
		Title:     "Entity Test",
		Type:      notasdl.Borderless,
		Resizable: true,
		TargetFPS: 60,
	}

	var wm notasdl.WindowManager

	win, err := wm.Create(&cfg)
	if err != nil {
		panic(err)
	}
	win1, err := wm.Create(&cfg)
	if err != nil {
		panic(err)
	}
	win2, err := wm.Create(&cfg)
	if err != nil {
		panic(err)
	}

	for !win.ShouldClose && !win1.ShouldClose && !win2.ShouldClose {

		var ev sdl.Event

		for sdl.PollEvent(&ev) {
			switch ev.Type {
			case sdl.EVENT_WINDOW_CLOSE_REQUESTED:
				win.ShouldClose = true

			case sdl.EVENT_QUIT:
				win.ShouldClose = true
			}
		}

		win.MakeCurrent()

		gl.ClearColor(0, 1, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		win.RenderFrame()

		win1.MakeCurrent()

		gl.ClearColor(1, 0, 0, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		win1.RenderFrame()

		win2.MakeCurrent()

		gl.ClearColor(0, 0, 1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		win2.RenderFrame()
	}

	_ = win

}
