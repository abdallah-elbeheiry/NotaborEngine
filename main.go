package main

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
	"NotaborEngine/notaentity"
	"NotaborEngine/notamath"
	"NotaborEngine/notasdl"
	"NotaborEngine/notatask"
	"fmt"
	"log"
	"time"
)

func main() {
	// Engine setup
	settings := &notacore.Settings{
		Vsync:      true,
		SoundLevel: 1,
		Muted:      false,
	}

	drawingLoop := notatask.CreateLoop(60)

	engine, err := notacore.CreateEngine(settings)
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Shutdown()

	cfg := &notasdl.WindowConfig{
		X:         50,
		Y:         50,
		W:         800,
		H:         600,
		Title:     "Entity Test",
		Type:      notasdl.Windowed,
		Resizable: true,
		TargetFPS: 60,
		Loops:     []*notatask.Loop{drawingLoop},
	}

	win, err := engine.CreateWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	em := engine.EntityManager
	circleRadius := float32(0.25)

	ballVisual, err := win.LoadVisual("quadSprite", "resources/images/hahaha.jpg", notasdl.VisualOptions{
		Width:        circleRadius * 2,
		Height:       circleRadius * 2,
		Mask:         notasdl.MaskCircle,
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

	moveStep := float32(0.05)
	inputCtx := engine.Input.GetContext()

	moveLeft := notacore.Input("moveLeft", notacore.KeyA, inputCtx)
	moveRight := notacore.Input("moveRight", notacore.KeyD, inputCtx)
	moveUp := notacore.Input("moveUp", notacore.KeyW, inputCtx)
	moveDown := notacore.Input("moveDown", notacore.KeyS, inputCtx)

	combo := notacore.InputCombo("combo", inputCtx, notacore.KeyE, notacore.KeyQ)

	leftClickSignal := notacore.Input("leftClick", notacore.MouseLeft, inputCtx)

	drawingLoop.Do(func() {
		engine.Input.BeginFrame()

		var moveX, moveY float32

		if moveLeft.Held() {
			moveX -= 1
		}
		if moveRight.Held() {
			moveX += 1
		}
		if moveUp.Held() {
			moveY += 1
		}
		if moveDown.Held() {
			moveY -= 1
		}

		if moveX != 0 || moveY != 0 {
			movement := notamath.Vec2{X: moveX, Y: moveY}.Mul(moveStep)
			entity.Move(movement)
		}

		// Standard signals for comparison
		if leftClickSignal.Pressed() {
			fmt.Println("left click")
		}

		if combo.Pressed() {
			fmt.Println("combo pressed")
		}

		em.Flush()
		alpha := drawingLoop.Alpha(time.Now())
		err := win.Draw(alpha, nil, entity)
		if err != nil {
			panic(err)
		}
	})

	if err := engine.Run(); err != nil {
		log.Fatal(err)
	}
}
