package notacollision

import (
	"NotaborEngine/notamath"
)

type AABBCollider struct {
	Min notamath.Vec2
	Max notamath.Vec2
}

type Collider interface {
	AABB() AABBCollider
	UpdateFromTransform(t *notamath.Transform2D)
}

func BroadPhase(a, b Collider) bool {
	return AABBIntersects(a.AABB(), b.AABB())
}

func AABBIntersects(a, b AABBCollider) bool {
	return a.Min.X <= b.Max.X &&
		a.Max.X >= b.Min.X &&
		a.Min.Y <= b.Max.Y &&
		a.Max.Y >= b.Min.Y
}

func Intersects(a, b Collider) bool {
	if !BroadPhase(a, b) {
		return false
	}

	switch a := a.(type) {
	case *CircleCollider:
		switch b := b.(type) {
		case *CircleCollider:
			return circleVsCircle(a, b)
		case *PolygonCollider:
			return circleVsPolygon(a, b)
		}

	case *PolygonCollider:
		switch b := b.(type) {
		case *CircleCollider:
			return circleVsPolygon(b, a)
		case *PolygonCollider:
			return polygonVsPolygon(a, b)
		}
	}

	return false
}

func polygonVsPolygon(a, b *PolygonCollider) bool {
	aVertices := a.GetWorldVertices()
	bVertices := b.GetWorldVertices()

	nA := len(aVertices)
	nB := len(bVertices)

	for i := 0; i < nA; i++ {
		a1 := aVertices[i]
		a2 := aVertices[(i+1)%nA]

		for j := 0; j < nB; j++ {
			b1 := bVertices[j]
			b2 := bVertices[(j+1)%nB]

			if segmentsIntersect(a1, a2, b1, b2) {
				return true
			}
		}
	}

	if pointInPolygon(aVertices[0], bVertices) {
		return true
	}

	if pointInPolygon(bVertices[0], aVertices) {
		return true
	}

	return false
}

func circleVsCircle(a, b *CircleCollider) bool {
	aCenter := a.WorldCenter()
	bCenter := b.WorldCenter()

	dx := aCenter.X - bCenter.X
	dy := aCenter.Y - bCenter.Y

	r := a.WorldRadius() + b.WorldRadius()

	return dx*dx+dy*dy <= r*r
}

func circleVsPolygon(c *CircleCollider, p *PolygonCollider) bool {
	center := c.WorldCenter()
	radius := c.WorldRadius()
	r2 := radius * radius

	vertices := p.GetWorldVertices()
	n := len(vertices)

	for i := 0; i < n; i++ {
		a := vertices[i]
		b := vertices[(i+1)%n]

		closest := closestPointOnSegment(a, b, center)

		if center.DistanceSquared(closest) <= r2 {
			return true
		}
	}

	if pointInPolygon(center, vertices) {
		return true
	}

	return false
}

//HELPERS

const epsilon float32 = 1e-6

func segmentsIntersect(p1, p2, q1, q2 notamath.Po2) bool {
	o1 := notamath.Orient(p1, p2, q1)
	o2 := notamath.Orient(p1, p2, q2)
	o3 := notamath.Orient(q1, q2, p1)
	o4 := notamath.Orient(q1, q2, p2)

	// Proper intersection
	if o1*o2 < 0 && o3*o4 < 0 {
		return true
	}

	// Collinear cases
	if almostZero(o1) && onSegment(p1, p2, q1) {
		return true
	}
	if almostZero(o2) && onSegment(p1, p2, q2) {
		return true
	}
	if almostZero(o3) && onSegment(q1, q2, p1) {
		return true
	}
	if almostZero(o4) && onSegment(q1, q2, p2) {
		return true
	}

	return false
}

func almostZero(v float32) bool {
	if v < 0 {
		return -v < epsilon
	}
	return v < epsilon
}

func onSegment(a, b, p notamath.Po2) bool {
	return p.X >= min(a.X, b.X)-epsilon &&
		p.X <= max(a.X, b.X)+epsilon &&
		p.Y >= min(a.Y, b.Y)-epsilon &&
		p.Y <= max(a.Y, b.Y)+epsilon
}

func pointInPolygon(point notamath.Po2, poly []notamath.Po2) bool {
	inside := false
	n := len(poly)

	for i := 0; i < n; i++ {
		j := (i + n - 1) % n

		pi := poly[i]
		pj := poly[j]

		intersect := ((pi.Y > point.Y) != (pj.Y > point.Y)) &&
			(point.X < (pj.X-pi.X)*(point.Y-pi.Y)/(pj.Y-pi.Y)+pi.X)

		if intersect {
			inside = !inside
		}
	}

	return inside
}

func closestPointOnSegment(a, b notamath.Po2, p notamath.Po2) notamath.Po2 {
	ab := b.Sub(a)
	ap := p.Sub(a)

	t := ap.Dot(ab) / ab.LenSquared()

	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	return notamath.Po2{
		X: a.X + ab.X*t,
		Y: a.Y + ab.Y*t,
	}
}
