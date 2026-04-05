package main

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
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

	logicLoop := notatask.CreateLoop(50000)
	drawingLoop := notatask.CreateLoop(60)

	engine.SetInputFrequency(2000)
	engine.SoundManager.SetSoundsFolder("resources/sounds")
	im := engine.InputManager
	em := engine.EntityManager
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

	entity := em.CreateEntity("quad").
		WithSprite(sprite).
		WithCollider(notacollision.NewPolygonCollider(rect.Points)).WithShader(shader).
		WithColor(notacolor.White)

	// Static walls
	rect1 := notageometry.CreateRectangle(0.5, 2)
	sprite1 := &notatexture.Sprite{Texture: texture, Name: "quadSprite", Polygon: rect1}
	entity1 := em.CreateEntity("wall").
		WithPolygon(rect1).
		WithCollider(notacollision.NewPolygonCollider(rect1.Points)).
		WithSprite(sprite1).
		WithColor(notacolor.Green).WithShader(shader)
	entity1.Move(notamath.Vec2{X: 1.15, Y: 0})

	rect2 := notageometry.CreateRectangle(0.5, 2)
	entity2 := em.CreateEntity("wall2").
		WithPolygon(rect2).
		WithCollider(notacollision.NewPolygonCollider(rect2.Points)).
		WithColor(notacolor.Red).WithShader(shader)
	entity2.Move(notamath.Vec2{X: -1.15, Y: 0})

	em.AddToCollisionGroup("group0", entity)
	em.AddToCollisionGroup("group0", entity1)
	em.AddToCollisionGroup("group1", entity)
	em.AddToCollisionGroup("group1", entity2)

	em.Flush()

	// Add draw calls
	drawingLoop.Add(notatask.CreateTask(func() error {
		err := entity.Draw(win.RunTime.Renderer, logicLoop.Alpha(time.Now()))
		err = entity1.Draw(win.RunTime.Renderer, logicLoop.Alpha(time.Now()))
		err = entity2.Draw(win.RunTime.Renderer, logicLoop.Alpha(time.Now()))
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
	sigLeft := &notacore.InputSignal{}

	im.BindInput(notacore.KeyW, sigW)
	im.BindInput(notacore.KeyA, sigA)
	im.BindInput(notacore.KeyS, sigS)
	im.BindInput(notacore.KeyD, sigD)
	im.BindInput(notacore.MouseLeft, sigLeft)

	// Actions
	actW := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actS := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actA := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actD := &notacore.Action{Behavior: notacore.RunWhileHeld}
	actMouseLeft := &notacore.Action{Behavior: notacore.RunOnceWhenPressed}

	// Bind Actions to Signals
	im.BindAction(sigW, actW)
	im.BindAction(sigA, actA)
	im.BindAction(sigS, actS)
	im.BindAction(sigD, actD)
	im.BindAction(sigLeft, actMouseLeft)
	// END OF INPUT MAPPING

	// START OF GAME LOGIC
	var deltaMove notamath.Vec2

	// Movement speed
	speed := float32(0.001)

	actW.AddTask(notatask.CreateTask(func() error { deltaMove.Y += speed; return nil }, notatask.RunOnce()))
	actS.AddTask(notatask.CreateTask(func() error { deltaMove.Y -= speed; return nil }, notatask.RunOnce()))
	actA.AddTask(notatask.CreateTask(func() error { deltaMove.X -= speed; return nil }, notatask.RunOnce()))
	actD.AddTask(notatask.CreateTask(func() error { deltaMove.X += speed; return nil }, notatask.RunOnce()))
	actMouseLeft.AddTask(notatask.CreateTask(func() error { fmt.Println("Mouse left clicked"); return nil }, notatask.RunOnce()))

	logicLoop.Add(notatask.CreateTask(func() error {
		entity.Move(deltaMove)
		collision, mtv := em.CollidesMTV(entity, entity1)
		if collision {
			entity.Move(mtv)
		}
		collision = false
		mtv = notamath.Vec2{}
		collision, mtv = em.CollidesMTV(entity, entity2)

		if collision {
			entity.Move(mtv)
		}
		deltaMove = notamath.Vec2{} // reset
		return nil
	}))
	val := notatomic.Float32{}
	val.Set(1)
	logicLoop.Add(notatask.CreateTask(func() error {
		entity.Move(notamath.Vec2{X: 0.001 * val.Get(), Y: 0})
		entity.Rotate(0.001)
		em.Flush()
		em.SolveGroupCollision("group0")
		em.SolveGroupCollision("group1")
		collision, mtv := em.CollidesMTV(entity, entity1)
		if collision {
			//err := engine.SoundManager.Play("ding.mp3", notasound.MP3, 1, false)
			val.Set(val.Get() * -1)
			entity.Move(mtv)
		}
		collision = false
		mtv = notamath.Vec2{}
		collision, mtv = em.CollidesMTV(entity, entity2)

		if collision {
			val.Set(val.Get() * -1)
			entity.Move(mtv)
		}
		return nil
	}))

	i := notatomic.Int64{}

	incrementCounter := func() error {
		i.Inc()
		return nil
	}

	lastTs := time.Now()

	incrementTask := notatask.CreateTask(incrementCounter)
	printLoopTask := notatask.CreateTask(func() error {
		now := time.Now()
		delta := now.Sub(lastTs).Seconds()
		count := i.Get()
		freq := float64(count) / delta
		fmt.Printf("Logic loop frequency: %.2f Hz\n", freq)
		i.Set(0)
		lastTs = now
		return nil
	}, notatask.RepeatEvery(time.Second*2))

	logicLoop.Add(notatask.CreateTask(func() error {
		em.Flush()
		return nil
	}))
	logicLoop.Add(incrementTask)
	logicLoop.Add(printLoopTask)
	// END OF GAME LOGIC

	// RUN ENGINE
	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}
