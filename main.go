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
		log.Fatal(err)
	}

	engine.SetInputFrequency(1000) // 1000 Hz input polling
	inputManager := engine.InputManager

	// Create InputSignals for WASD
	keyW := &notacore.InputSignal{}
	keyA := &notacore.InputSignal{}
	keyS := &notacore.InputSignal{}
	keyD := &notacore.InputSignal{}

	inputManager.BindInput(notacore.KeyW, keyW)
	inputManager.BindInput(notacore.KeyA, keyA)
	inputManager.BindInput(notacore.KeyS, keyS)
	inputManager.BindInput(notacore.KeyD, keyD)

	// Load texture
	texture, err := win.LoadTexture("test", "resources/hahaha.jpg")
	if err != nil {
		log.Fatal(err)
	}

	rect := notaobject.CreateRectangle(0.5, 0.5)
	sprite := &notaobject.Sprite{
		Texture: texture,
		Name:    "quadSprite",
		Polygon: rect,
	}
	entity := notaobject.NewEntity("quad", "Test Quad").
		WithSprite(sprite).
		WithCollider(notacollision.NewPolygonCollider(rect.Points()))

	// Add static walls
	rect1 := notaobject.CreateRectangle(0.2, 2)
	sprite1 := &notaobject.Sprite{
		Texture: texture,
		Name:    "quadSprite",
		Polygon: rect1,
	}

	entity1 := notaobject.NewEntity("wall", "Test Wall").
		WithPolygon(rect1).
		WithCollider(notacollision.NewPolygonCollider(rect1.Points())).
		WithSprite(sprite1)
	rect1.SetColor(notaobject.Green)
	entity1.Move(notamath.Vec2{X: 1, Y: 0})

	rect2 := notaobject.CreateRectangle(0.2, 2)
	entity2 := notaobject.NewEntity("wall2", "Test Wall").
		WithPolygon(rect2).
		WithCollider(notacollision.NewPolygonCollider(rect2.Points()))
	rect2.SetColor(notaobject.Red)
	entity2.Move(notamath.Vec2{X: -1, Y: 0})

	// Render loop
	renderLoop.Add(func() error { entity.Draw(win.RunTime.Renderer); return nil })
	renderLoop.Add(func() error { entity1.Draw(win.RunTime.Renderer); return nil })
	renderLoop.Add(func() error { entity2.Draw(win.RunTime.Renderer); return nil })

	speed := float32(0.005)

	// Logic loop: movement with WASD
	logicLoop.Add(func() error {
		move := notamath.Vec2{X: 0, Y: 0}
		if keyW.Down() {
			move.Y += speed
		}
		if keyS.Down() {
			move.Y -= speed
		}
		if keyA.Down() {
			move.X -= speed
		}
		if keyD.Down() {
			move.X += speed
		}

		entity.Move(move)

		// Collision check
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			// Undo move if collision detected
			entity.Move(move.Mul(-1))
		}

		return nil
	})

	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
