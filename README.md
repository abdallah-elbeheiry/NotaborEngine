# Notabor Engine

**Notabor Engine** is a lightweight game/physics engine written in Go.
It aims to simplify game development by replacing complex object-oriented update systems with a **function-driven loop architecture**.

Instead of requiring developers to implement large class hierarchies or manually manage frame delta calculations, the engine organizes execution around **Runnable functions scheduled inside deterministic loops**.

The result is a **minimal, predictable, and composable execution model**.

---

# Features

* Function-based architecture
* Deterministic fixed-tick logic loops
* Parallel execution using worker pools
* Signal-based input system
* Simple action binding
* GLFW window management
* OpenGL rendering backend
* Dynamic runtime task scheduling
* Multiple window support

---

# Core Concepts

## Runnable Functions

The fundamental unit of execution in Notabor Engine is the **Runnable**.

```go
type Runnable func() error
```

A Runnable represents any piece of game logic that can be executed by the engine.

Examples include:

* physics updates
* AI logic
* entity behavior
* animation updates
* rendering commands

When a Runnable returns an error, the engine automatically removes it from the loop that executed it. This allows tasks to self-terminate without additional state management.

---

# Loop-Based Execution

Notabor Engine replaces traditional game update loops with **engine-managed loops**.

Two primary loop types exist.

## FixedHzLoop

A **FixedHzLoop** executes Runnables at a fixed tick rate.

Example configuration:

```
Physics loop: 120 Hz
AI loop:      30 Hz
Input loop:   240 Hz
```

This guarantees deterministic execution timing and avoids the need for manual delta-time scaling.

The loop internally:

1. Executes all active Runnables
2. Removes Runnables that returned errors
3. Schedules newly added Runnables
4. Runs tasks in parallel using a worker pool
5. Maintains precise tick scheduling

Because the tick rate is fixed, developers do **not need to multiply values by delta time**.

---

## RenderLoop

The **RenderLoop** executes rendering tasks.

Unlike FixedHzLoop, it runs as fast as possible (or up to a configured maximum Hz) and is responsible for:

* clearing the frame
* executing render Runnables
* flushing draw orders to the renderer backend

---

# Runnable Scheduling

Runnables can be dynamically attached to loops:

```go
loop.Add(func() error {
    updatePlayer()
    return nil
})
```

This approach enables:

* modular systems
* dynamic task scheduling
* simple lifecycle management

If a Runnable returns an error:

```
return errors.New("done")
```

the engine automatically removes it from the loop.

---

# Parallel Task Execution

`FixedHzLoop` uses a **worker pool** internally.

Each tick:

1. Runnables are converted into jobs
2. Jobs are dispatched to worker goroutines
3. Results are collected
4. Failed tasks are removed

This allows large numbers of independent systems to run concurrently without developers needing to manually manage goroutines.

---

# Input System

The engine includes a **signal-based input system**.

Instead of polling keys directly, inputs are mapped to **InputSignals** that track state transitions:

* Pressed
* Released
* Held
* Idle
* Changed

Example:

```go
if jumpSignal.Pressed() {
    player.Jump()
}
```

Signals can be bound to **Actions**, which execute Runnables when specific conditions are met.

Supported behaviors include:

* `RunWhileHeld`
* `RunWhileToggled`
* `RunOnceWhenPressed`
* `RunOnceWhenReleased`
* `RunWhileIdle`

This system works with keyboard, mouse, and gamepads.

---

# Window and Rendering System

Notabor Engine uses **GLFW** for window management and **OpenGL** for rendering.

Each window contains:

* a renderer
* a texture manager
* a render loop
* one or more logic loops

Supported window modes:

* Windowed
* Borderless
* Fullscreen

The engine automatically handles:

* viewport scaling
* aspect ratio preservation
* OpenGL initialization
* rendering backend setup

---

# Multiple Window Support

Notabor Engine can manage **multiple windows simultaneously**.

Each window is created through the engine and registered inside the internal `windowManager`. Every window maintains its own:

* OpenGL context
* renderer
* texture manager
* render loop
* logic loops

This allows applications to run multiple independent rendering contexts such as:

* multi-view editors
* debugging windows
* split simulation views
* tools and inspectors

Example creation flow:

```go
renderLoop := &notacore.RenderLoop{MaxHz: 60}
logicLoop := &notacore.FixedHzLoop{Hz: 120}

cfg := notacore.WindowConfig{
    X:          50,
    Y:          50,
    W:          800,
    H:          600,
    Title:      "Main Window",
    Type:       notacore.Windowed,
    Resizable:  true,
    RenderLoop: renderLoop,
    LogicLoops: []*notacore.FixedHzLoop{logicLoop},
}

win, err := engine.CreateWindow(cfg)
```

Additional windows can be created with the same API, each with their own render and logic loops.

---

# Engine Execution Model

The engine coordinates:

* window management
* input processing
* fixed logic loops
* render loops

Execution flow:

1. Initialize the engine
2. Create windows
3. Start the input loop
4. Start logic loops
5. Run render loops for all windows until they close

All systems are driven by **Runnable scheduling**, eliminating the need for user-written update loops.

---

# Example

Below is a minimal example demonstrating window creation and loop usage.

```go
engine := &notacore.Engine{
    Settings: &notacore.Settings{Vsync: true},
}

engine.InitPlatform()

renderLoop := &notacore.RenderLoop{MaxHz: 60}
logicLoop := &notacore.FixedHzLoop{Hz: 120}

cfg := notacore.WindowConfig{
    X:          50,
    Y:          50,
    W:          800,
    H:          600,
    Title:      "Example",
    Type:       notacore.Windowed,
    Resizable:  true,
    RenderLoop: renderLoop,
    LogicLoops: []*notacore.FixedHzLoop{logicLoop},
}

win, _ := engine.CreateWindow(cfg)

renderLoop.Add(func() error {
    // draw calls here
    return nil
})

logicLoop.Add(func() error {
    // update logic here
    return nil
})

engine.Run()
```

---

# Entity System

Notabor Engine provides a lightweight **entity abstraction** for organizing objects in the world. An entity represents a transformable object that may optionally contain rendering and collision components.

Each entity contains:

* a `Transform2D` (position, rotation, scale)
* optional rendering components
* optional collision components
* runtime flags controlling activity and visibility

Entities are intentionally **minimal and composable** rather than implementing a heavy ECS framework.

Example entity structure:

```go
type Entity struct {
    ID        string
    Name      string
    Transform notamath.Transform2D
    Active    bool
    Visible   bool

    Sprite   *Sprite
    Polygon  *Polygon
    Collider notacollision.Collider
    Shader   *Shader
}
```

Entities support a **builder-style configuration API** for attaching components:

```go
entity := notaobject.NewEntity("quad", "Player").
    WithSprite(sprite).
    WithCollider(notacollision.NewPolygonCollider(points))
```

### Transform Updates

Entity movement automatically updates the associated collider to keep the physics representation synchronized with the transform.

```go
entity.Move(delta)
entity.Rotate(angle)
```

Internally the collider receives the updated transform so collision checks remain correct.

### Rendering

Entities can submit draw commands directly to the renderer.

```go
entity.Draw(renderer)
```

Depending on the attached components the entity may render:

* a polygon
* a textured sprite
* both

The renderer converts submissions into **batched draw orders** which are flushed by the `RenderLoop`.

### Collision

Entities can perform collision tests using their colliders:

```go
if entity.CollidesWith(other) {
    // handle collision
}
```

The collision system is implemented in the `notacollision` package and supports polygon-based intersection checks.

---

# Design Goals

## Simplicity

Game logic is expressed as plain functions rather than complex class hierarchies.

## Deterministic Timing

`FixedHzLoop` ensures consistent simulation timing without manual delta calculations.

## Concurrency by Default

The engine automatically parallelizes tasks where appropriate.

## Minimal Boilerplate

Developers write gameplay logic directly rather than wiring complex engine structures.

## Composable Systems

Behavior can be dynamically added or removed at runtime through Runnable scheduling.

---

# Status

The engine is currently under active development.

---

# License

MIT License
