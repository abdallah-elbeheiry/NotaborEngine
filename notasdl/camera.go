package notasdl

import (
	"NotaborEngine/notamath"
	"sync"
	"time"
)

// Camera2D controls how a window views world-space content.
type Camera2D struct {
	mu       sync.RWMutex
	position notamath.Vec2
	rotation float32
	zoom     notamath.Vec2

	// Smooth transitions
	targetPos    notamath.Vec2
	startPos     notamath.Vec2
	moveDuration float32
	moveElapsed  float32
	isMoving     bool

	targetZoom   notamath.Vec2
	startZoom    notamath.Vec2
	zoomDuration float32
	zoomElapsed  float32
	isZooming    bool
}

// NewCamera2D creates a 2D camera with identity zoom and no offset.
func NewCamera2D() *Camera2D {
	return &Camera2D{
		zoom:       notamath.Vec2{X: 1, Y: 1},
		targetZoom: notamath.Vec2{X: 1, Y: 1},
	}
}

// Position returns the camera's world-space center.
func (c *Camera2D) Position() notamath.Vec2 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.position
}

// SetPosition moves the camera's world-space center.
func (c *Camera2D) SetPosition(position notamath.Vec2) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.position = position
	c.isMoving = false
}

// Move translates the camera by a world-space delta.
func (c *Camera2D) Move(delta notamath.Vec2) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.position = c.position.Add(delta)
	c.isMoving = false
}

// SmoothMove initiates a linear movement to a target position over a duration.
func (c *Camera2D) SmoothMove(target notamath.Vec2, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startPos = c.position
	c.targetPos = target
	c.moveDuration = float32(duration.Seconds())
	c.moveElapsed = 0
	c.isMoving = true
}

// Rotation returns the camera rotation in radians.
func (c *Camera2D) Rotation() float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rotation
}

// SetRotation rotates the camera viewport in radians.
func (c *Camera2D) SetRotation(rotation float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotation = rotation
}

// Rotate adds a rotation delta in radians to the camera.
func (c *Camera2D) Rotate(delta float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rotation += delta
}

// Zoom returns the camera zoom factor on each axis.
func (c *Camera2D) Zoom() notamath.Vec2 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.zoom
}

// SetZoom sets the same zoom factor for both axes.
func (c *Camera2D) SetZoom(zoom float32) {
	c.SetZoomXY(zoom, zoom)
}

// SetZoomXY sets the camera zoom factor for each axis.
func (c *Camera2D) SetZoomXY(x, y float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.zoom = sanitizeZoom(notamath.Vec2{X: x, Y: y})
	c.isZooming = false
}

// SmoothZoom initiates a linear zoom to a target factor over a duration.
func (c *Camera2D) SmoothZoom(target float32, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startZoom = c.zoom
	c.targetZoom = notamath.Vec2{X: target, Y: target}
	c.zoomDuration = float32(duration.Seconds())
	c.zoomElapsed = 0
	c.isZooming = true
}

// Update processes smooth transitions. dt is in seconds.
func (c *Camera2D) Update(dt float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isMoving {
		if c.moveDuration <= 0 {
			c.position = c.targetPos
			c.isMoving = false
		} else {
			c.moveElapsed += dt
			if c.moveElapsed >= c.moveDuration {
				c.position = c.targetPos
				c.isMoving = false
			} else {
				t := c.moveElapsed / c.moveDuration
				c.position = c.startPos.Lerp(c.targetPos, t)
			}
		}
	}

	if c.isZooming {
		if c.zoomDuration <= 0 {
			c.zoom = c.targetZoom
			c.isZooming = false
		} else {
			c.zoomElapsed += dt
			if c.zoomElapsed >= c.zoomDuration {
				c.zoom = c.targetZoom
				c.isZooming = false
			} else {
				t := c.zoomElapsed / c.zoomDuration
				c.zoom = c.startZoom.Lerp(c.targetZoom, t)
			}
		}
	}
}

// ViewMatrix returns the affine matrix that transforms world space into this camera's view.
func (c *Camera2D) ViewMatrix() notamath.Mat3 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// To zoom IN (make things larger), we need to scale the WORLD by the zoom factor.
	// Since the camera's world transform is inverted to get the view matrix,
	// we should use 1/zoom in the camera's world transform.
	z := sanitizeZoom(c.zoom)
	invZoom := notamath.Vec2{X: 1.0 / z.X, Y: 1.0 / z.Y}
	return notamath.Mat3TRS(c.position, c.rotation, invZoom).InverseAffine()
}

func sanitizeZoom(zoom notamath.Vec2) notamath.Vec2 {
	if zoom.X == 0 {
		zoom.X = 1
	}
	if zoom.Y == 0 {
		zoom.Y = 1
	}
	return zoom
}
