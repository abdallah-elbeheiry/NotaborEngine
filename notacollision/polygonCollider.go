// notacollision/polygon.go

package notacollision

import (
	"NotaborEngine/notamath"
)

type PolygonCollider struct {
	LocalVertices []notamath.Po2
	WorldVertices []notamath.Po2
}

func NewPolygonCollider(points []notamath.Po2) *PolygonCollider {
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

// MTVPolygon computes the Minimum Translation Vector to separate two polygons
func MTVPolygon(a, b *PolygonCollider) notamath.Vec2 {
	var MaxMTVPerFrame = mTVTravelDistance

	if len(a.WorldVertices) == 0 || len(b.WorldVertices) == 0 {
		return notamath.Vec2{}
	}

	minOverlap := float32(1e30)
	var smallestAxis notamath.Vec2

	// Get all axes for SAT
	axes := append(getAxes(a.WorldVertices), getAxes(b.WorldVertices)...)

	for _, axis := range axes {
		minA, maxA := projectPolygon(axis, a.WorldVertices)
		minB, maxB := projectPolygon(axis, b.WorldVertices)

		overlap := getOverlap(minA, maxA, minB, maxB)
		if overlap <= 0 {
			// No collision
			return notamath.Vec2{}
		}
		if overlap < minOverlap {
			minOverlap = overlap
			smallestAxis = axis
		}
	}

	// Compute centroids for direction
	dir := a.GetCentroid().Sub(b.GetCentroid())
	if dot(dir, smallestAxis) < 0 {
		smallestAxis = smallestAxis.Neg()
	}

	// Compute MTV
	mtv := smallestAxis.Mul(minOverlap)

	// Clamp MTV to avoid extreme movement
	if mtv.Len() > MaxMTVPerFrame {
		mtv = mtv.Normalize().Mul(MaxMTVPerFrame)
	}

	return mtv
}

func (p *PolygonCollider) GetCentroid() notamath.Vec2 {
	var cx, cy float32
	n := len(p.WorldVertices)
	for _, v := range p.WorldVertices {
		cx += v.X
		cy += v.Y
	}
	return notamath.Vec2{X: cx / float32(n), Y: cy / float32(n)}
}

// getAxes returns the normalized perpendicular axes of a polygon's edges
func getAxes(vertices []notamath.Po2) []notamath.Vec2 {
	n := len(vertices)
	axes := make([]notamath.Vec2, n)
	for i := 0; i < n; i++ {
		curr := vertices[i]
		next := vertices[(i+1)%n]
		edge := notamath.Vec2{X: next.X - curr.X, Y: next.Y - curr.Y}
		axes[i] = notamath.Vec2{X: -edge.Y, Y: edge.X}.Normalize()
	}
	return axes
}

// projectPolygon projects all vertices onto an axis
func projectPolygon(axis notamath.Vec2, vertices []notamath.Po2) (min, max float32) {
	min = dot(axis, notamath.Vec2{X: vertices[0].X, Y: vertices[0].Y})
	max = min
	for _, v := range vertices[1:] {
		p := dot(axis, notamath.Vec2{X: v.X, Y: v.Y})
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
	}
	return
}

// getOverlap computes the scalar overlap between two projections
func getOverlap(minA, maxA, minB, maxB float32) float32 {
	return min(maxA, maxB) - max(minA, minB)
}

// dot product helper
func dot(a, b notamath.Vec2) float32 {
	return a.X*b.X + a.Y*b.Y
}
