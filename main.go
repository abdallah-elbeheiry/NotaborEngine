package main

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
	"NotaborEngine/notaentity"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatask"
	"NotaborEngine/notatexture"
	"NotaborEngine/notatomic"
	"fmt"
	"log"
	"time"
)

func main() {
	// START OF ENGINE SETUP
	Settings := &notacore.Settings{Vsync: true, SoundLevel: 0.2, Muted: false}
	engine, err := notacore.CreateEngine(Settings)
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Shutdown()

	logicLoop := notatask.CreateLoop(1000)
	drawingLoop := notatask.CreateLoop(120)

	engine.SetInputFrequency(3000)
	engine.SoundManager.SetSoundsFolder("resources/sounds")
	im := engine.InputManager

	// END OF ENGINE SETUP

	// START OF WINDOW CREATION
	cfg := notacore.WindowConfig{
		X:         50,
		Y:         50,
		W:         800,
		H:         600,
		Title:     "Entity Test",
		Type:      notacore.Windowed,
		Resizable: true,
		TargetFPS: 60,
		Loops:     []*notatask.Loop{logicLoop, drawingLoop},
	}

	win, err := engine.CreateWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// END OF WINDOW CREATION

	// START OF ENTITY CREATION
	texture, err := win.LoadTexture("test", "resources/images/hahaha.jpg")
	if err != nil {
		log.Fatal(err)
	}
	rect := notageometry.CreateRectangle(0.5, 0.5)
	sprite := &notatexture.Sprite{
		Texture: texture,
		Name:    "quadSprite",
		Polygon: rect,
	}
	shader, err := notashader.NewShader("notashader/shaders/basic.vert", "notashader/shaders/basic.frag")

	if err != nil {
		fmt.Println(err)
	}

	entity := notaentity.NewEntity("quad", "Test Quad").
		WithSprite(sprite).
		WithCollider(notacollision.NewPolygonCollider(rect.Points)).WithShader(shader).
		WithColor(notacolor.White)

	// Static walls
	rect1 := notageometry.CreateRectangle(0.2, 2)
	sprite1 := &notatexture.Sprite{Texture: texture, Name: "quadSprite", Polygon: rect1}
	entity1 := notaentity.NewEntity("wall", "Test Wall").
		WithPolygon(rect1).
		WithCollider(notacollision.NewPolygonCollider(rect1.Points)).
		WithSprite(sprite1).
		WithColor(notacolor.Green).WithShader(shader)
	entity1.Move(notamath.Vec2{X: 1, Y: 0})

	rect2 := notageometry.CreateRectangle(0.2, 2)
	entity2 := notaentity.NewEntity("wall2", "Test Wall").
		WithPolygon(rect2).
		WithCollider(notacollision.NewPolygonCollider(rect2.Points)).
		WithColor(notacolor.Red).WithShader(shader)
	entity2.Move(notamath.Vec2{X: -1, Y: 0})

	// Add draw calls
	drawingLoop.Add(notatask.CreateTask(func() error {
		err := entity.Draw(win, logicLoop)
		err = entity1.Draw(win, logicLoop)
		err = entity2.Draw(win, logicLoop)
		if err != nil {
			return err
		}
		return nil
	}))
	// END OF ENTITY CREATION

	// START OF INPUT MAPPING
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
	// END OF INPUT MAPPING

	// START OF GAME LOGIC
	var deltaMove notamath.Vec2

	// Movement speed
	speed := float32(0.001)

	actW.AddTask(notatask.CreateTask(func() error { deltaMove.Y += speed; return nil }, notatask.RunOnce()))
	actS.AddTask(notatask.CreateTask(func() error { deltaMove.Y -= speed; return nil }, notatask.RunOnce()))
	actA.AddTask(notatask.CreateTask(func() error { deltaMove.X -= speed; return nil }, notatask.RunOnce()))
	actD.AddTask(notatask.CreateTask(func() error { deltaMove.X += speed; return nil }, notatask.RunOnce()))

	logicLoop.Add(notatask.CreateTask(func() error {
		entity.Move(deltaMove)
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			entity.Move(deltaMove.Neg()) // undo
		}
		deltaMove = notamath.Vec2{} // reset
		return nil
	}))
	val := 1.0
	logicLoop.Add(notatask.CreateTask(func() error {
		entity.Move(notamath.Vec2{X: float32(0.001 * val), Y: 0})
		if entity.CollidesWith(entity1) || entity.CollidesWith(entity2) {
			//err := engine.SoundManager.Play("ding.mp3", notasound.MP3, 1, false)
			if err != nil {
				return err
			}
			val *= -1
		}
		return nil
	}))

	i := notatomic.Int64{}

	incrementCounter := func() error {
		i.Inc()
		return nil
	}

	printLoopSpeed := func() error {
		fmt.Printf("Average Hz: %d\n", i.Get()/10)
		return nil
	}

	incrementTask := notatask.CreateTask(incrementCounter, notatask.FinishAfter(time.Second*10))
	printLoopTask := notatask.CreateTask(printLoopSpeed, notatask.WithDelay(time.Second*11), notatask.RunOnce())

	logicLoop.Add(incrementTask)

	logicLoop.Add(printLoopTask)
	// END OF GAME LOGIC

	// RUN ENGINE
	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
