package notassets

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notagl"
	"NotaborEngine/notamath"
)

type Entity struct {
	ID        string
	Name      string
	Transform notamath.Transform2D
	Active    bool
	Visible   bool

	// Components - can be nil
	Sprite   *Sprite
	Polygon  *notagl.Polygon
	Collider notacollision.Collider
}

func NewEntity(id, name string) *Entity {
	return &Entity{
		ID:        id,
		Name:      name,
		Active:    true,
		Visible:   true,
		Transform: notamath.NewTransform2D(),
	}
}

// Builders

func (e *Entity) WithSprite(s *Sprite) *Entity {
	e.Sprite = s
	return e
}

func (e *Entity) WithPolygon(p *notagl.Polygon) *Entity {
	e.Polygon = p
	return e
}

func (e *Entity) WithCollider(c notacollision.Collider) *Entity {
	e.Collider = c
	return e
}

// Move and Rotate automatically update collider
func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active {
		return
	}
	e.Transform.TranslateBy(delta)
	e.updateCollider()
}

func (e *Entity) Rotate(rad float32) {
	if !e.Active {
		return
	}
	e.Transform.RotateBy(rad)
	e.updateCollider()
}

func (e *Entity) updateCollider() {
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

func (e *Entity) CollidesWith(other *Entity) bool {
	if e.Collider == nil || other.Collider == nil {
		return false
	}
	return notacollision.Intersects(e.Collider, other.Collider)
}
