package notacollision

import (
	"NotaborEngine/notamath"
	"math"
)

type AABBCollider struct {
	Min notamath.Vec2
	Max notamath.Vec2
}

var mTVTravelDistance float32 = 1.0

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

func IntersectsMTV(a, b Collider) (bool, notamath.Vec2) {
	if !BroadPhase(a, b) {
		return false, notamath.Vec2{}
	}

	switch a := a.(type) {
	case *CircleCollider:
		switch b := b.(type) {
		case *CircleCollider:
			return circleVsCircleMTV(a, b)
		case *PolygonCollider:
			return circleVsPolygonMTV(a, b)
		}
	case *PolygonCollider:
		switch b := b.(type) {
		case *CircleCollider:
			ok, mtv := circleVsPolygonMTV(b, a)
			return ok, mtv.Neg() // invert MTV for polygon->circle
		case *PolygonCollider:
			return polygonVsPolygonMTV(a, b)
		}
	}

	return false, notamath.Vec2{}
}

func polygonVsPolygonMTV(a, b *PolygonCollider) (bool, notamath.Vec2) {
	mtv := MTVPolygon(a, b)
	if mtv.X == 0 && mtv.Y == 0 {
		return false, notamath.Vec2{}
	}
	return true, mtv
}

func circleVsCircleMTV(a, b *CircleCollider) (bool, notamath.Vec2) {
	dir := b.WorldCenter().Sub(a.WorldCenter())
	distSq := dir.LenSquared()
	rSum := a.WorldRadius() + b.WorldRadius()
	if distSq >= rSum*rSum {
		return false, notamath.Vec2{}
	}

	dist := float32(1.0)
	if distSq > 0 {
		dist = float32(math.Sqrt(float64(distSq)))
	}

	overlap := rSum - dist
	mtv := dir.Mul(1 / dist).Mul(overlap) // normalize dir and scale by overlap
	return true, mtv
}

func circleVsPolygonMTV(c *CircleCollider, p *PolygonCollider) (bool, notamath.Vec2) {
	axes := getAxes(p.WorldVertices)
	center := c.WorldCenter()

	minOverlap := float32(1e30)
	var smallestAxis notamath.Vec2

	for _, axis := range axes {
		minP, maxP := projectPolygon(axis, p.WorldVertices)
		projC := dot(axis, notamath.Vec2{X: center.X, Y: center.Y})
		minC := projC - c.WorldRadius()
		maxC := projC + c.WorldRadius()

		overlap := getOverlap(minP, maxP, minC, maxC)
		if overlap <= 0 {
			return false, notamath.Vec2{}
		}
		if overlap < minOverlap {
			minOverlap = overlap
			smallestAxis = axis
		}
	}

	dir := p.WorldVertices[0].Sub(center)
	if dot(dir, smallestAxis) < 0 {
		smallestAxis = smallestAxis.Neg()
	}

	return true, smallestAxis.Mul(minOverlap)
}

// SetMaximumMTVTravelDistance sets the maximum travel distance for the MTV calculation per frame. the default is 1.0
// Use at your own risk, numbers too high or too low may cause instability
func SetMaximumMTVTravelDistance(amount float32) {
	mTVTravelDistance = amount
}
