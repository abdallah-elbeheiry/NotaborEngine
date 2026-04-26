package main

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
	"NotaborEngine/notaentity"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notatask"
	"NotaborEngine/notatomic"
	"fmt"
	"log"
	"math"
	"runtime"
	"time"
)

func main() {
	// START OF ENGINE SETUP
	Settings := &notacore.Settings{Vsync: true, SoundLevel: 1, Muted: false}
	engine, err := notacore.CreateEngine(Settings)
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Shutdown()

	logicLoop := notatask.CreateLoop(10000)
	drawingLoop := notatask.CreateLoop(60)

	fmt.Println(runtime.NumCPU())

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
	win.Camera().SetZoom(1.0)
	// END OF WINDOW CREATION

	// START OF ENTITY CREATION
	circleRadius := float32(0.25)
	ballVisual, err := win.LoadVisual("quadSprite", "resources/images/hahaha.jpg", notacore.VisualOptions{
		Width:        circleRadius * 2,
		Height:       circleRadius * 2,
		Mask:         notacore.MaskCircle,
		CircleRadius: 0.5,
		CircleEdge:   0.01,
	})

	if err != nil {
		log.Fatal(err)
	}

	entity := em.CreateEntity("quad").
		WithVisual(ballVisual).
		WithCollision(notaentity.CircleCollision(circleRadius)).
		WithColor(notacolor.White)

	// Static walls
	rect1 := notageometry.CreateRectangle(0.5, 2)
	texture, err := win.GetTexture("quadSprite")
	if err != nil {
		log.Fatal(err)
	}
	wallVisual, err := win.CreateVisual("wallSprite", texture, notacore.SpriteOptions(0.5, 2))
	if err != nil {
		log.Fatal(err)
	}
	entity1 := em.CreateEntity("wall").
		WithPolygon(rect1).
		WithVisual(wallVisual).
		WithCollision(notaentity.PolygonCollision(rect1.Points)).
		WithColor(notacolor.Green)
	entity1.Move(notamath.Vec2{X: 1.15, Y: 0})

	rect2 := notageometry.CreateRectangle(0.5, 2)
	wallVisual2, err := win.CreateVisual("wallSprite2", texture, notacore.SpriteOptions(0.5, 2))
	if err != nil {
		log.Fatal(err)
	}
	entity2 := em.CreateEntity("wall2").
		WithPolygon(rect2).
		WithVisual(wallVisual2).
		WithCollision(notaentity.PolygonCollision(rect2.Points)).
		WithColor(notacolor.Red)
	entity2.Move(notamath.Vec2{X: -1.15, Y: 0})

	em.AddToCollisionGroup("group0", entity)
	em.AddToCollisionGroup("group0", entity1)
	em.AddToCollisionGroup("group1", entity)
	em.AddToCollisionGroup("group1", entity2)

	em.Flush()

	// Add draw calls
	drawingLoop.Do(func() {
		alpha := logicLoop.Alpha(time.Now())
		err := win.Draw(alpha, nil, entity, entity1, entity2)
		if err != nil {
			panic(err)
		}
	})
	// END OF ENTITY CREATION

	// START OF INPUT MAPPING
	sigW := &notacore.InputSignal{}
	sigA := &notacore.InputSignal{}
	sigS := &notacore.InputSignal{}
	sigD := &notacore.InputSignal{}
	sigLeft := &notacore.InputSignal{}
	sigZ := &notacore.InputSignal{}
	sigX := &notacore.InputSignal{}
	sigC := &notacore.InputSignal{}

	im.BindInput(notacore.KeyW, sigW)
	im.BindInput(notacore.KeyA, sigA)
	im.BindInput(notacore.KeyS, sigS)
	im.BindInput(notacore.KeyD, sigD)
	im.BindInput(notacore.MouseLeft, sigLeft)
	im.BindInput(notacore.KeyZ, sigZ)
	im.BindInput(notacore.KeyX, sigX)
	im.BindInput(notacore.KeyC, sigC)

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

	actW.AddTask(notatask.Once(func() { deltaMove.Y += speed }))
	actS.AddTask(notatask.Once(func() { deltaMove.Y -= speed }))
	actA.AddTask(notatask.Once(func() { deltaMove.X -= speed }))
	actD.AddTask(notatask.Once(func() { deltaMove.X += speed }))
	actMouseLeft.AddTask(notatask.Once(func() { fmt.Println("Mouse left clicked") }))

	logicLoop.Do(func() {
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
	})
	val := notatomic.Float32{}
	val.Set(1)

	logicLoop.Do(func() {
		entity.Move(notamath.Vec2{X: 0.001 * val.Get(), Y: 0})
		entity.Rotate(0.001)
		em.Flush()
		em.SolveGroupCollision("group0")
		em.SolveGroupCollision("group1")
		collision, mtv := em.CollidesMTV(entity, entity1)
		if collision {
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
	})

	lastTs := time.Now()
	lastTickCount := logicLoop.TickCount()
	scalePhase := float32(0)
	printLoopTask := notatask.Every(time.Second*2, func() {
		now := time.Now()
		delta := now.Sub(lastTs).Seconds()
		currentTickCount := logicLoop.TickCount()
		count := currentTickCount - lastTickCount
		freq := float64(count) / delta
		fmt.Printf("Logic loop frequency: %.2f Hz\n", freq)
		lastTickCount = currentTickCount
		lastTs = now
	})

	logicLoop.Do(func() {
		em.Flush()
	})
	logicLoop.Do(func() {
		scalePhase += 0.0025
		targetScale := 1.0 + 0.35*math32Sin(scalePhase)
		currentScale := entity.ScaleValue()
		if currentScale.X != 0 && currentScale.Y != 0 {
			entity.Scale(notamath.Vec2{
				X: targetScale / currentScale.X,
				Y: targetScale / currentScale.Y,
			})
		}
	})
	logicLoop.Add(printLoopTask)
	// END OF GAME LOGIC

	// RUN ENGINE
	if err := engine.Run(); err != nil {
		log.Fatal("Engine run failed:", err)
	}
}

func math32Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}
