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

	Color notatomic.Pointer[notacolor.Color]

	lastSubmittedFrame notatomic.UInt64
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

func (e *Entity) Move(delta notamath.Vec2) {
	if !e.Active.Get() {
		return
	}
	e.manager.SubmitMove(e.index, delta)
}

func (e *Entity) Rotate(rad float32) {
	if !e.Active.Get() {
		return
	}
	e.manager.SubmitRotation(e.index, rad)
}

func (e *Entity) Scale(factor notamath.Vec2) {
	if !e.Active.Get() {
		return
	}
	e.manager.SubmitScale(e.index, factor)
}

func (e *Entity) Position() notamath.Vec2 {
	return e.manager.getPositionIndex(e.index)
}

func (e *Entity) Rotation() float32 {
	return e.manager.getRotationIndex(e.index)
}

func (e *Entity) ScaleValue() notamath.Vec2 {
	return e.manager.getScaleIndex(e.index)
}

func (e *Entity) Draw(renderer *notarender.Renderer, alpha float32) error {
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

	transform := notamath.Transform2D{
		Position: pos,
		Rotation: rot,
		Scale:    scale,
	}
	transform.Snapshot()
	model := transform.InterpolatedMatrix(alpha)

	color := e.Color.Get()
	if color == nil {
		color = &notacolor.White
	}

	if sprite := e.Sprite.Get(); sprite != nil && sprite.Polygon != nil {
		renderer.SubmitPolygon(sprite.Polygon, model, *color, sprite.Texture, e.Shader.Get())
		return nil
	}

	if poly := e.Polygon.Get(); poly != nil {
		renderer.SubmitPolygon(poly, model, *color, nil, e.Shader.Get())
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
