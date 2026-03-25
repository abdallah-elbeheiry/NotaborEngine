package notageometry

import (
	"NotaborEngine/notamath"
)

// Polygon is pure geometry—no rendering data
type Polygon struct {
	Points []notamath.Po2
}

func (p *Polygon) Fixate() {
	if len(p.Points) == 0 {
		return
	}
	center := PolygonCentroid(p.Points)
	for i := range p.Points {
		p.Points[i].X -= center.X
		p.Points[i].Y -= center.Y
	}
}

func CreateRectangle(w, h float32) *Polygon {
	hw := w / 2
	hh := h / 2
	return &Polygon{
		Points: []notamath.Po2{
			{X: -hw, Y: -hh},
			{X: +hw, Y: -hh},
			{X: +hw, Y: +hh},
			{X: -hw, Y: +hh},
		},
	}
}

func PointInTriangle(p, a, b, c notamath.Po2) bool {
	o1 := notamath.Orient(a, b, p)
	o2 := notamath.Orient(b, c, p)
	o3 := notamath.Orient(c, a, p)

	hasNeg := (o1 < 0) || (o2 < 0) || (o3 < 0)
	hasPos := (o1 > 0) || (o2 > 0) || (o3 > 0)

	return !(hasNeg && hasPos)
}

func PolygonCentroid(points []notamath.Po2) notamath.Po2 {
	var cx, cy, area float32

	n := len(points)
	for i := 0; i < n; i++ {
		p0 := points[i]
		p1 := points[(i+1)%n]

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
