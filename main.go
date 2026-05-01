package main

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
	"NotaborEngine/notaentity"
	"NotaborEngine/notasdl"
	"NotaborEngine/notatask"
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

	drawingLoop.Do(func() {
		alpha := drawingLoop.Alpha(time.Now())
		err := win.Draw(alpha, nil, entity)
		if err != nil {
			panic(err)
		}
	})

	if err := engine.Run(); err != nil {
		log.Fatal(err)
	}

	_ = win
}
