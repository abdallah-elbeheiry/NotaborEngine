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
	"fmt"
)

type Entity struct {
	ID   string
	Name string

	// Use your atomic wrappers instead of primitives
	Transform notatomic.Pointer[notamath.Transform2D]
	Active    notatomic.Bool
	Visible   notatomic.Bool

	// Components wrapped in atomic pointers
	Sprite   notatomic.Pointer[notatexture.Sprite]
	Polygon  notatomic.Pointer[notageometry.Polygon]
	Collider notatomic.Pointer[notacollision.Collider]
	Shader   notatomic.Pointer[notashader.Shader]

	// Rendering properties
	Color notatomic.Pointer[notacolor.Color]
}

func NewEntity(id, name string) *Entity {
	e := &Entity{
		ID:   id,
		Name: name,
	}
	e.Active.Set(true)
	e.Visible.Set(true)
	trans := notamath.NewTransform2D()
	e.Transform.Set(&trans)

	// Default color
	defaultColor := notacolor.White
	e.Color.Set(&defaultColor)

	return e
}

// Builder methods
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

func (e *Entity) WithColor(c notacolor.Color) *Entity {
	e.Color.Set(&c)
	return e
}

// Movement methods
func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active.Get() {
		return
	}

	for {
		oldT := e.Transform.Get()
		newT := *oldT
		newT.TranslateBy(delta)

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

// Draw submits rendering commands to the renderer
func (e *Entity) Draw(renderer *notarender.Renderer) {
	if !e.Visible.Get() || !e.Active.Get() {
		return
	}

	shader := e.Shader.Get()
	if shader == nil {
		shader = renderer.DefaultShader
		if shader == nil {
			newShader, err := notashader.NewShader("notashader/shaders/basic.vert", "notashader/shaders/basic.frag")
			if err != nil {
				return
			}
			shader = newShader
			e.Shader.Set(shader)
		}
	}

	// Get transform
	t := e.Transform.Get()
	model := t.Matrix()

	// Get color
	color := e.Color.Get()
	if color == nil {
		color = &notacolor.White
	}

	// Render sprite if present
	if sprite := e.Sprite.Get(); sprite != nil && sprite.Polygon != nil {
		renderer.SubmitPolygon(sprite.Polygon, model, *color, sprite.Texture, shader)
		return
	}

	// Otherwise render polygon directly
	if poly := e.Polygon.Get(); poly != nil {
		renderer.SubmitPolygon(poly, model, *color, nil, shader)
	} else {
		fmt.Printf("Entity %s: has no polygon or sprite!\n", e.Name)
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
