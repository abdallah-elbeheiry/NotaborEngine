package notaentity

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notacolor"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notarender"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"NotaborEngine/notatomic"
)

type Entity struct {
	ID string

	index   int
	manager *EntityManager

	Active  notatomic.Bool
	Visible notatomic.Bool

	Sprite   notatomic.Pointer[notatexture.Sprite]
	Polygon  notatomic.Pointer[notageometry.Polygon]
	Collider notatomic.Pointer[notacollision.Collider]
	Shader   notatomic.Pointer[notashader.Shader]
	Material notatomic.Pointer[notashader.Material]

	Color notatomic.Pointer[notacolor.Color]

	lastSubmittedFrame notatomic.UInt64
}

type Visual struct {
	Sprite   *notatexture.Sprite
	Material *notashader.Material
}

// CollisionProfile describes how an entity should participate in collision queries.
type CollisionProfile struct {
	Collider notacollision.Collider
}

func newEntity(id string, index int, manager *EntityManager) *Entity {
	e := &Entity{
		ID:      id,
		index:   index,
		manager: manager,
	}

	e.Active.Set(true)
	e.Visible.Set(true)
	e.Color.Set(&notacolor.White)

	return e
}

// NewVisual creates a reusable visual bundle from a sprite and optional material.
func NewVisual(sprite *notatexture.Sprite, material *notashader.Material) *Visual {
	return &Visual{
		Sprite:   sprite,
		Material: material,
	}
}

// CircleCollision creates a circular collision profile centered on the entity origin.
func CircleCollision(radius float32) CollisionProfile {
	return CollisionProfile{
		Collider: notacollision.NewCircleCollider(notamath.Po2{}, radius),
	}
}

// PolygonCollision creates a polygon collision profile from local-space points.
func PolygonCollision(points []notamath.Po2) CollisionProfile {
	return CollisionProfile{
		Collider: notacollision.NewPolygonCollider(points),
	}
}

// CustomCollision wraps an already constructed collider as a collision profile.
func CustomCollision(collider notacollision.Collider) CollisionProfile {
	return CollisionProfile{
		Collider: collider,
	}
}

func (e *Entity) WithSprite(s *notatexture.Sprite) *Entity {
	e.Sprite.Set(s)
	return e
}

func (e *Entity) WithPolygon(p *notageometry.Polygon) *Entity {
	e.Polygon.Set(p)
	return e
}

func (e *Entity) WithCollider(c notacollision.Collider) *Entity {
	e.Collider.Set(&c)
	return e
}

func (e *Entity) WithShader(s *notashader.Shader) *Entity {
	e.Shader.Set(s)
	return e
}

func (e *Entity) WithMaterial(m *notashader.Material) *Entity {
	e.Material.Set(m)
	return e
}

// WithVisual assigns a sprite and optional material to the entity in one call.
func (e *Entity) WithVisual(v *Visual) *Entity {
	if v == nil {
		return e
	}
	if v.Sprite != nil {
		e.WithSprite(v.Sprite)
	}
	if v.Material != nil {
		e.WithMaterial(v.Material)
	}
	return e
}

// WithCollision assigns a collision profile to the entity in one call.
func (e *Entity) WithCollision(profile CollisionProfile) *Entity {
	if profile.Collider == nil {
		return e
	}
	return e.WithCollider(profile.Collider)
}

// WithCircle is a compatibility helper that assigns a circular collider directly.
func (e *Entity) WithCircle(radius float32) *Entity {
	return e.WithCollider(notacollision.NewCircleCollider(notamath.Po2{}, radius))
}

// WithCircleSprite is a compatibility helper that assigns a sprite, material, and inferred circle collider together.
func (e *Entity) WithCircleSprite(sprite *notatexture.Sprite, material *notashader.Material) *Entity {
	e.WithSprite(sprite)
	e.WithMaterial(material)

	if sprite != nil && sprite.Polygon != nil {
		e.WithCircle(circleRadiusFromPolygon(sprite.Polygon))
	}

	return e
}

func (e *Entity) WithColor(c notacolor.Color) *Entity {
	e.Color.Set(&c)
	return e
}

// Move moves an entity by a vector amount, movement is additively applied
func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active.Get() {
		return
	}
	e.manager.submitMove(e.index, delta)
}

// Rotate rotates an entity by an amount (radians), rotating is additively applied
func (e *Entity) Rotate(rad float32) {
	if !e.Active.Get() {
		return
	}
	e.manager.submitRotation(e.index, rad)
}

// Scale scales an entity by an amount, scaling is multiplicatively applied
// scale vectors scales horizontally and vertically according to the vector's x and y values
func (e *Entity) Scale(factor notamath.Vec2) {
	if !e.Active.Get() {
		return
	}
	e.manager.submitScale(e.index, factor)
}

// Position gets the current position of the entity
func (e *Entity) Position() notamath.Vec2 {
	return e.manager.getPositionIndex(e.index)
}

// Rotation gets the current rotation degree of the entity (radians)
func (e *Entity) Rotation() float32 {
	return e.manager.getRotationIndex(e.index)
}

// ScaleValue gets the current relative scale of the entity
func (e *Entity) ScaleValue() notamath.Vec2 {
	return e.manager.getScaleIndex(e.index)
}

// Draw sends a draw request to the window's renderer and draws at the next tick
// alpha is recommended to be provided by the loop (loop.Alpha(time.now())) if the object is dynamic, 1 is the object is stationary
// the draw will happen on the window which the renderer belongs to
func (e *Entity) Draw(renderer *notarender.Renderer, alpha float32) error {
	return e.DrawWithView(renderer, notamath.Mat3Identity(), alpha)
}

// DrawWithView sends a draw request to the renderer using the provided view matrix.
func (e *Entity) DrawWithView(renderer *notarender.Renderer, view notamath.Mat3, alpha float32) error {
	if !e.Visible.Get() || !e.Active.Get() {
		return nil
	}

	frame := renderer.FrameID.Get()

	for {
		last := e.lastSubmittedFrame.Get()
		if last == frame {
			return nil
		}
		if e.lastSubmittedFrame.CompareAndSwap(last, frame) {
			break
		}
	}

	pos := e.manager.getPositionIndex(e.index)
	scale := e.manager.getScaleIndex(e.index)
	rot := e.manager.getRotationIndex(e.index)

	// For now, since Transform2D is created locally, Snapshot() would just set prev=current.
	// If we want actual interpolation, the Entity needs to store its Transform2D and Snapshot()
	// it at the beginning of the logic tick.
	// For now, we'll just use the current values.
	model := view.Mul(notamath.Mat3TRS(pos, rot, scale))

	color := e.Color.Get()
	if color == nil {
		color = &notacolor.White
	}

	if sprite := e.Sprite.Get(); sprite != nil && sprite.Polygon != nil {
		renderer.SubmitPolygon(sprite.Polygon, model, *color, sprite.Texture, e.Shader.Get(), e.Material.Get())
		return nil
	}

	if poly := e.Polygon.Get(); poly != nil {
		renderer.SubmitPolygon(poly, model, *color, nil, e.Shader.Get(), e.Material.Get())
	}

	return nil
}

// Collider update
func (e *Entity) updateCollider() {
	cPtr := e.Collider.Get()
	if cPtr == nil {
		return
	}

	pos := e.manager.getPositionIndex(e.index)
	rot := e.manager.getRotationIndex(e.index)
	scale := e.manager.getScaleIndex(e.index)

	t := notamath.Transform2D{}
	t.SetPosition(pos)
	t.SetRotation(rot)
	t.SetScale(scale)

	c := *cPtr
	c.UpdateFromTransform(&t)
}

func (e *Entity) GetId() string {
	return e.ID
}

func circleRadiusFromPolygon(poly *notageometry.Polygon) float32 {
	if poly == nil || len(poly.Points) == 0 {
		return 0
	}

	minX, maxX := poly.Points[0].X, poly.Points[0].X
	minY, maxY := poly.Points[0].Y, poly.Points[0].Y
	for _, p := range poly.Points[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	width := maxX - minX
	height := maxY - minY
	if width < height {
		return width * 0.5
	}
	return height * 0.5
}
