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

	renderLoop := notacore.CreateRenderLoop(60)
	logicLoop := notacore.CreateFixedHzLoop(200000)
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

	engine.SetInputFrequency(3000)
	im := engine.InputManager

	// Load texture and create entity
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

	// Static walls
	rect1 := notaobject.CreateRectangle(0.2, 2)
	sprite1 := &notaobject.Sprite{Texture: texture, Name: "quadSprite", Polygon: rect1}
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

	// Movement speed
	speed := float32(0.001)

	// input
	sigW := &notacore.InputSignal{}
	sigA := &notacore.InputSignal{}
	sigS := &notacore.InputSignal{}
	sigD := &notacore.InputSignal{}

	im.BindInput(notacore.KeyW, sigW)
	im.BindInput(notacore.KeyA, sigA)
	im.BindInput(notacore.KeyS, sigS)
	im.BindInput(notacore.KeyD, sigD)

	// Actions
	actW := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actS := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actA := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actD := &notacore.Action{Behavior: notacore.RunWhileHeld}

	// Bind Actions to Signals
	im.BindAction(sigW, actW)
	im.BindAction(sigA, actA)
	im.BindAction(sigS, actS)
	im.BindAction(sigD, actD)

	var deltaMove notamath.Vec2

	actW.AddRunnable(func() error { deltaMove.Y += speed; return nil })
	actS.AddRunnable(func() error { deltaMove.Y -= speed; return nil })
	actA.AddRunnable(func() error { deltaMove.X -= speed; return nil })
	actD.AddRunnable(func() error { deltaMove.X += speed; return nil })

	logicLoop.Add(func() error {
		entity.Move(deltaMove)
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			entity.Move(deltaMove.Neg()) // undo
		}
		deltaMove = notamath.Vec2{} // reset
		return nil
	})
	val := 1.0
	logicLoop.Add(func() error {
		entity.Move(notamath.Vec2{X: float32(0.001 * val), Y: 0})
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			val *= -1
		}
		return nil
	})

	// Run engine
	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
