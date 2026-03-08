package main

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notacore"
	"NotaborEngine/notamath"
	"NotaborEngine/notaobject"
	"log"
	"time"
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
	logicLoop := &notacore.FixedHzLoop{Hz: 1000}

	logicLoop.EnableMonitor(time.Second)

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
	win, err := engine.CreateWindow(cfg)
	if err != nil {
		panic(err)
	}

	texture, err := win.LoadTexture("test", "resources/hahaha.jpg")
	if err != nil {
		panic(err)
	}

	shader, _ := notaobject.NewShader("notaobject/shaders/basic.vert", "notaobject/shaders/basic.frag")
	shader.SetUniform(notaobject.UseTexture, true)
	rect := notaobject.CreateTextureQuad(0.5, 0.5)
	sprite := &notaobject.Sprite{
		Texture: texture,
		Name:    "quadSprite",
		Polygon: rect,
	}

	circle := shader.Clone()
	circle.SetUniform(notaobject.UseCircle, true)
	circle.SetUniform(notaobject.UseTexture, true)
	circle.SetUniform(notaobject.CircleRadius, 0.1)
	entity := notaobject.NewEntity("quad", "Test Quad").
		WithSprite(sprite).
		WithCollider(notacollision.NewCircleCollider(notaobject.PolygonCentroid(rect.Vertices), 0.25)).WithShader(circle)

	rect1 := notaobject.CreateRectangle(0.2, 2)
	entity1 := notaobject.NewEntity("wall", "Test Wall").
		WithPolygon(rect1).
		WithCollider(notacollision.NewPolygonCollider(notaobject.PolygonPoints(rect1))).WithShader(shader)
	rect1.SetColor(notaobject.Green)
	entity1.Move(notamath.Vec2{X: 1, Y: 0})

	rect2 := notaobject.CreateRectangle(0.2, 2)
	entity2 := notaobject.NewEntity("wall2", "Test Wall").
		WithPolygon(rect2).
		WithCollider(notacollision.NewPolygonCollider(notaobject.PolygonPoints(rect2))).WithShader(shader)
	rect2.SetColor(notaobject.Red)
	entity2.Move(notamath.Vec2{X: -1, Y: 0})

	renderLoop.Add(func() error {
		entity.Draw(win.RunTime.Renderer)
		return nil
	})
	renderLoop.Add(func() error {
		entity1.Draw(win.RunTime.Renderer)
		return nil
	})
	renderLoop.Add(func() error {
		entity2.Draw(win.RunTime.Renderer)
		return nil
	})

	direction := 1.0

	logicLoop.Add(func() error {
		entity.Rotate(float32(direction * 0.01))
		entity.Move(notamath.Vec2{X: 0.001}.Mul(float32(direction)))
		return nil
	})
	logicLoop.Add(func() error {
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			direction *= -1.0
		}
		return nil
	})

	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
