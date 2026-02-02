package notassets

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notagl"
	"NotaborEngine/notamath"
)

type Entity struct {
	ID   string
	Name string

	Transform notamath.Transform2D

	Sprite   *Sprite
	Polygon  *notagl.Polygon
	Collider notacollision.Collider

	Active  bool
	Visible bool
}

// NewEntity creates a basic empty entity
func NewEntity(id, name string) *Entity {
	return &Entity{
		ID:        id,
		Name:      name,
		Active:    true,
		Visible:   true,
		Transform: notamath.NewTransform2D(),
	}
}

func (e *Entity) SetSprite(s *Sprite) {
	e.Sprite = s
}

func (e *Entity) SetPolygon(p *notagl.Polygon)         { e.Polygon = p }
func (e *Entity) SetCollider(c notacollision.Collider) { e.Collider = c }

func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active {
		return
	}

	e.Transform.TranslateBy(delta)

	if e.Collider != nil {
		e.Collider.UpdateFromTransform(&e.Transform)
	}
}

func (e *Entity) Rotate(rad float32) {
	if !e.Active {
		return
	}

	e.Transform.RotateBy(rad)

	if e.Collider != nil {
		e.Collider.UpdateFromTransform(&e.Transform)
	}
}

func (e *Entity) Draw(renderer *notagl.Renderer2D) {
	if !e.Visible || !e.Active {
		return
	}

	model := e.Transform.Matrix()

	if e.Polygon != nil {
		renderer.Submit(e.Polygon, model, nil)
	}

	if e.Sprite != nil && e.Sprite.Polygon != nil {
		renderer.Submit(e.Sprite.Polygon, model, e.Sprite.Texture)
	}
}

// CollidesWith returns true if this entity collides with another
func (e *Entity) CollidesWith(other *Entity) bool {
	if e.Collider == nil || other.Collider == nil {
		return false
	}
	return notacollision.Intersects(e.Collider, other.Collider)
}
