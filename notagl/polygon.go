package notagl

import (
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
)

type Polygon struct {
	Vertices []Vertex2D
	Color    notashader.Color // Default uniform color
}

func (p *Polygon) SetColor(c notashader.Color) {
	p.Color = c

	// Clear vertex overrides
	for i := range p.Vertices {
		p.Vertices[i].Color = notashader.Color{}
	}
}

// Fixate Adjusts points according to center point
func (p *Polygon) Fixate() {
	if len(p.Vertices) == 0 {
		return
	}

	center := polygonCentroid(p.Vertices)

	for i := range p.Vertices {
		p.Vertices[i].Pos.X -= center.X
		p.Vertices[i].Pos.Y -= center.Y
	}
}

func (p *Polygon) AddToOrders(model notamath.Mat3, orders *[]DrawOrder) {
	if len(p.Vertices) < 3 {
		return
	}

	verts := make([]Vertex2D, len(p.Vertices))

	for i, v := range p.Vertices {
		verts[i] = v
		verts[i].Pos = model.TransformPo2(v.Pos)

		if v.Color == (notashader.Color{}) {
			verts[i].Color = p.Color
		}
	}

	*orders = append(*orders, DrawOrder{
		Vertices: verts,
	})
}

func (p *Polygon) SetVerticalGradient(top, bottom notashader.Color) {
	if len(p.Vertices) == 0 {
		return
	}

	minY := p.Vertices[0].Pos.Y
	maxY := p.Vertices[0].Pos.Y

	for _, v := range p.Vertices {
		if v.Pos.Y < minY {
			minY = v.Pos.Y
		}
		if v.Pos.Y > maxY {
			maxY = v.Pos.Y
		}
	}

	rangeY := maxY - minY
	if rangeY == 0 {
		return
	}

	for i, v := range p.Vertices {
		t := (v.Pos.Y - minY) / rangeY
		p.Vertices[i].Color = bottom.Lerp(top, t)
	}
}

func (p *Polygon) SetHorizontalGradient(left, right notashader.Color) {
	if len(p.Vertices) == 0 {
		return
	}

	minX := p.Vertices[0].Pos.X
	maxX := p.Vertices[0].Pos.X

	for _, v := range p.Vertices {
		if v.Pos.X < minX {
			minX = v.Pos.X
		}
		if v.Pos.X > maxX {
			maxX = v.Pos.X
		}
	}

	rangeX := maxX - minX
	if rangeX == 0 {
		return
	}

	for i, v := range p.Vertices {
		t := (v.Pos.X - minX) / rangeX
		p.Vertices[i].Color = left.Lerp(right, t)
	}
}

func CreateRectangle(w, h float32) *Polygon {
	hw := w / 2
	hh := h / 2

	p := Polygon{
		Vertices: []Vertex2D{
			{Pos: notamath.Po2{X: -hw, Y: -hh}},
			{Pos: notamath.Po2{X: +hw, Y: -hh}},
			{Pos: notamath.Po2{X: +hw, Y: +hh}},
			{Pos: notamath.Po2{X: -hw, Y: +hh}},
		},
		Color: notashader.Color{R: 1, G: 1, B: 1, A: 1},
	}

	return &p
}

// CreateCircle creates a uniform quad
// to actually make this into a circle you will use to use the default shader and set UseCircle to true and set radius/edge
func CreateCircle(radius float32) *Polygon {
	size := radius * 2
	return CreateRectangle(size, size)
}

func IsCCW(poly []notamath.Po2) bool {
	var area float32
	for i := 0; i < len(poly); i++ {
		a := poly[i]
		b := poly[(i+1)%len(poly)]
		area += (b.X - a.X) * (b.Y + a.Y)
	}
	return area < 0
}

func PointInTriangle(p, a, b, c notamath.Po2) bool {
	o1 := notamath.Orient(a, b, p)
	o2 := notamath.Orient(b, c, p)
	o3 := notamath.Orient(c, a, p)

	hasNeg := (o1 < 0) || (o2 < 0) || (o3 < 0)
	hasPos := (o1 > 0) || (o2 > 0) || (o3 > 0)

	return !(hasNeg && hasPos)
}

func IsEar(prev, curr, next notamath.Po2, poly []notamath.Po2) bool {
	// Must be convex (CCW polygon)
	if notamath.Orient(prev, curr, next) <= 0 {
		return false
	}

	for _, p := range poly {
		if p == prev || p == curr || p == next {
			continue
		}
		if PointInTriangle(p, prev, curr, next) {
			return false
		}
	}
	return true
}

func polygonCentroid(poly []Vertex2D) notamath.Po2 {
	var cx, cy, area float32

	n := len(poly)
	for i := 0; i < n; i++ {
		p0 := poly[i].Pos
		p1 := poly[(i+1)%n].Pos

		cross := p0.X*p1.Y - p1.X*p0.Y
		area += cross
		cx += (p0.X + p1.X) * cross
		cy += (p0.Y + p1.Y) * cross
	}

	area *= 0.5
	if area == 0 {
		return notamath.Po2{}
	}

	inv := 1.0 / (6.0 * area)
	return notamath.Po2{
		X: cx * inv,
		Y: cy * inv,
	}
}

func CreateTextureQuad(width, height float32) *Polygon {
	hw := width / 2
	hh := height / 2

	p := Polygon{
		Vertices: []Vertex2D{
			{
				Pos: notamath.Po2{X: -hw, Y: -hh},
				UV:  notamath.Vec2{X: 0, Y: 0},
			},
			{
				Pos: notamath.Po2{X: +hw, Y: -hh},
				UV:  notamath.Vec2{X: 1, Y: 0},
			},
			{
				Pos: notamath.Po2{X: +hw, Y: +hh},
				UV:  notamath.Vec2{X: 1, Y: 1},
			},
			{
				Pos: notamath.Po2{X: -hw, Y: +hh},
				UV:  notamath.Vec2{X: 0, Y: 1},
			},
		},
		Color: notashader.White,
	}
	return &p
}
