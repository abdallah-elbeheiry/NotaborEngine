package notaentity

import (
	"NotaborEngine/notacollision"
	"NotaborEngine/notacolor"
	"NotaborEngine/notacore"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"NotaborEngine/notatomic"
	"errors"
	"time"
)

type Entity struct {
	ID   string
	Name string

	Transform notatomic.Pointer[notamath.Transform2D]
	Active    notatomic.Bool
	Visible   notatomic.Bool

	Sprite   notatomic.Pointer[notatexture.Sprite]
	Polygon  notatomic.Pointer[notageometry.Polygon]
	Collider notatomic.Pointer[notacollision.Collider]
	Shader   notatomic.Pointer[notashader.Shader]

	Color notatomic.Pointer[notacolor.Color]

	lastSubmittedFrame notatomic.UInt64
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
func (e *Entity) Draw(window *notacore.Window, loop *notacore.Loop) error {
	e.snapShot()
	if !e.Visible.Get() || !e.Active.Get() {
		return nil
	}

	renderer := window.RunTime.Renderer
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

	shader := e.Shader.Get()
	if shader == nil {
		return errors.New("Entity with ID " + e.ID + " has no shader")
	}

	alpha := loop.Alpha(time.Now())
	if alpha > 1 {
		alpha = 1
	}

	t := e.Transform.Get()
	model := t.InterpolatedMatrix(alpha)

	color := e.Color.Get()
	if color == nil {
		color = &notacolor.White
	}

	if sprite := e.Sprite.Get(); sprite != nil && sprite.Polygon != nil {
		renderer.SubmitPolygon(sprite.Polygon, model, *color, sprite.Texture, shader)
		return nil
	}

	if poly := e.Polygon.Get(); poly != nil {
		renderer.SubmitPolygon(poly, model, *color, nil, shader)
	}
	return nil
}

func (e *Entity) CollidesWith(other *Entity) bool {
	c1 := e.Collider.Get()
	c2 := other.Collider.Get()

	if c1 == nil || c2 == nil {
		return false
	}
	return notacollision.Intersects(*c1, *c2)
}

func (e *Entity) snapShot() {
	t := e.Transform.Get()
	t.Snapshot()
}
