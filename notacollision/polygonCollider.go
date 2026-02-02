// notacollision/polygon.go

package notacollision

import (
	"NotaborEngine/notagl"
	"NotaborEngine/notamath"
)

type PolygonCollider struct {
	// Static local vertices
	LocalVertices []notamath.Po2

	// Cached world vertices
	WorldVertices []notamath.Po2
}

func NewPolygonCollider(polygon *notagl.Polygon) *PolygonCollider {
	var points []notamath.Po2
	for _, vert := range polygon.Vertices {
		points = append(points, vert.Pos)
	}
	return &PolygonCollider{
		LocalVertices: points,
		WorldVertices: make([]notamath.Po2, len(points)),
	}
}

func (p *PolygonCollider) UpdateFromTransform(t *notamath.Transform2D) {
	if len(p.WorldVertices) != len(p.LocalVertices) {
		p.WorldVertices = make([]notamath.Po2, len(p.LocalVertices))
	}

	matrix := t.Matrix()

	for i, v := range p.LocalVertices {
		p.WorldVertices[i] = matrix.TransformPo2(v)
	}
}

func (p *PolygonCollider) AABB() AABBCollider {
	if len(p.WorldVertices) == 0 {
		return AABBCollider{}
	}

	minX := p.WorldVertices[0].X
	minY := p.WorldVertices[0].Y
	maxX := p.WorldVertices[0].X
	maxY := p.WorldVertices[0].Y

	for i := 1; i < len(p.WorldVertices); i++ {
		v := p.WorldVertices[i]
		if v.X < minX {
			minX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y > maxY {
			maxY = v.Y
		}
	}

	return AABBCollider{
		Min: notamath.Vec2{X: minX, Y: minY},
		Max: notamath.Vec2{X: maxX, Y: maxY},
	}
}

func (p *PolygonCollider) GetWorldVertices() []notamath.Po2 {
	return p.WorldVertices
}
