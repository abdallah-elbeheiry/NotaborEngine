package main

import (
	"NotaborEngine/notacore"
	"NotaborEngine/notagl"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notassets"
	"log"
)

func main() {
	engine := &notacore.Engine{
		Settings: &notacore.Settings{Vsync: true},
	}

	if err := engine.InitPlatform(); err != nil {
		log.Fatal("Init failed:", err)
	}
	defer engine.Shutdown()

	renderLoop := &notacore.RenderLoop{MaxHz: 60}
	logicLoop := &notacore.FixedHzLoop{Hz: 500}

	cfg := notacore.WindowConfig{
		X:          50,
		Y:          50,
		W:          800,
		H:          600,
		Title:      "Entity Test",
		Type:       notacore.Windowed,
		Resizable:  true,
		RenderLoop: renderLoop,
		LogicLoops: []*notacore.FixedHzLoop{logicLoop},
	}

	win, err := engine.CreateWindow2D(cfg)
	if err != nil {
		panic(err)
	}

	texture, err := win.LoadTexture("test", "resources/hahaha.jpg")
	if err != nil {
		panic(err)
	}

	texturedShader := notashader.Shader{
		Name:           "textured",
		VertexString:   notashader.TexturedVertex2D,
		FragmentString: notashader.TexturedFragment2D,
	}

	if err := win.CreateShader(texturedShader); err != nil {
		log.Fatal("Failed to create shader:", err)
	}

	if err := win.UseShader("textured"); err != nil {
		log.Fatal("Failed to use shader:", err)
	}

	entity := notassets.NewEntity("quad", "Test Quad")

	sprite := &notassets.Sprite{
		Texture: texture,
		Name:    "quadSprite",
		Polygon: notagl.CreateTextureQuad(1, 1),
	}

	rect := notagl.CreateRectangle(notamath.Po2{
		X: 0,
		Y: 0,
	}, 1, 2)
	rect.SetColor(notashader.Yellow)
	entity.SetSprite(sprite)
	entity.SetPolygon(rect)

	// draw entity
	renderLoop.Runnables = []notacore.Runnable{
		func() error {
			texture.Bind(0)
			entity.Draw(win.RunTime.Renderer)
			return nil
		},
	}

	// rotate and move entity
	logicLoop.Runnables = []notacore.Runnable{
		func() error {
			entity.Rotate(0.01)
			return nil
		},
	}

	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
