package notaobject

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notamath"
	"NotaborEngine/notatomic"
)

type Entity struct {
	ID   string
	Name string

	// Use your atomic wrappers instead of primitives
	Transform notatomic.Pointer[notamath.Transform2D]
	Active    notatomic.Bool
	Visible   notatomic.Bool

	// Components wrapped in atomic pointers
	Sprite   notatomic.Pointer[Sprite]
	Polygon  notatomic.Pointer[Polygon]
	Collider notatomic.Pointer[notacollision.Collider]
	Shader   notatomic.Pointer[Shader]
}

func NewEntity(id, name string) *Entity {
	e := &Entity{
		ID:   id,
		Name: name,
	}
	// Initialize atomic states
	e.Active.Set(true)
	e.Visible.Set(true)
	trans := notamath.NewTransform2D()
	e.Transform.Set(&trans)
	return e
}

// Builders (Thread-Safe)

func (e *Entity) WithSprite(s *Sprite) *Entity {
	e.Sprite.Set(s)
	return e
}

func (e *Entity) WithPolygon(p *Polygon) *Entity {
	e.Polygon.Set(p)
	return e
}

func (e *Entity) WithCollider(c notacollision.Collider) *Entity {
	e.Collider.Set(&c)
	return e
}

func (e *Entity) WithShader(s *Shader) *Entity {
	e.Shader.Set(s)
	return e
}

// Logic

func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active.Get() {
		return
	}

	for {
		oldT := e.Transform.Get()
		// Create a local copy to modify
		newT := *oldT
		newT.TranslateBy(delta)

		// Only swap if no other thread changed the transform in the meantime
		if e.Transform.CompareAndSwap(oldT, &newT) {
			e.updateCollider(&newT)
			break
		}
	}
}

func (e *Entity) Rotate(rad float32) {
	if !e.Active.Get() {
		return
	}

	for {
		oldT := e.Transform.Get()
		newT := *oldT
		newT.RotateBy(rad)

		if e.Transform.CompareAndSwap(oldT, &newT) {
			e.updateCollider(&newT)
			break
		}
	}
}

func (e *Entity) updateCollider(t *notamath.Transform2D) {
	if cPtr := e.Collider.Get(); cPtr != nil {
		c := *cPtr
		c.UpdateFromTransform(t)
	}
}

func (e *Entity) Draw(renderer *Renderer) {
	// Cache active/visible status once
	active := e.Active.Get()
	visible := e.Visible.Get()

	if !visible || !active {
		return
	}

	// Capture the current pointers
	shader := e.Shader.Get()
	if shader == nil {
		newShader, _ := NewShader("notaobject/shaders/basic.vert", "notaobject/shaders/basic.frag")
		e.Shader.CompareAndSwap(nil, newShader)
		shader = e.Shader.Get()
	}

	// Capture the transform snapshot
	t := e.Transform.Get()
	model := t.Matrix()

	if poly := e.Polygon.Get(); poly != nil {
		renderer.Submit(poly, model, nil, shader)
	}

	if sprite := e.Sprite.Get(); sprite != nil {
		if sprite.Polygon != nil {
			renderer.Submit(sprite.Polygon, model, sprite.Texture, shader)
		}
	}
}

func (e *Entity) CollidesWith(other *Entity) bool {
	c1 := e.Collider.Get()
	c2 := other.Collider.Get()

	if c1 == nil || c2 == nil {
		return false
	}
	return notacollision.Intersects(*c1, *c2)
}
