package notacollision

import (
	"NotaborEngine/notamath"
	"math"
)

// aabbCollider is a minimal bounding box collider used for broad phase checks
type aabbCollider struct {
	Min notamath.Vec2
	Max notamath.Vec2
}

var mTVTravelDistance float32 = 1.0

// Collider is an interface implemented by polygon collider and circular collider
type Collider interface {
	aabb() aabbCollider
	UpdateFromTransform(t *notamath.Transform2D)
}

func broadPhase(a, b Collider) bool {
	return aabbIntersects(a.aabb(), b.aabb())
}

func aabbIntersects(a, b aabbCollider) bool {
	return a.Min.X <= b.Max.X &&
		a.Max.X >= b.Min.X &&
		a.Min.Y <= b.Max.Y &&
		a.Max.Y >= b.Min.Y
}

// Intersects finds whether two colliders are intersecting or not, colliders can be polygon shaped or cirular
// Return whether colliders intersect or not and minimum translation vector (MTV)
func Intersects(a, b Collider) (bool, notamath.Vec2) {
	if !broadPhase(a, b) {
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
	mtv := mtvPolygon(a, b)
	if mtv.X == 0 && mtv.Y == 0 {
		return false, notamath.Vec2{}
	}
	return true, mtv
}

func circleVsCircleMTV(a, b *CircleCollider) (bool, notamath.Vec2) {
	dir := a.WorldCenter().Sub(b.WorldCenter())
	distSq := dir.LenSquared()
	rSum := a.WorldRadius() + b.WorldRadius()
	if distSq >= rSum*rSum {
		return false, notamath.Vec2{}
	}

	if distSq == 0 {
		return true, clampMTV(notamath.Vec2{X: rSum, Y: 0})
	}

	dist := float32(math.Sqrt(float64(distSq)))
	overlap := rSum - dist
	mtv := dir.Mul(1 / dist).Mul(overlap) // normalize dir and scale by overlap
	return true, clampMTV(mtv)
}

func circleVsPolygonMTV(c *CircleCollider, p *PolygonCollider) (bool, notamath.Vec2) {
	axes := getAxes(p.WorldVertices)
	center := c.WorldCenter()
	closest := closestVertex(center, p.WorldVertices)
	closestAxis := center.Sub(closest).Normalize()
	if closestAxis.LenSquared() > 0 {
		axes = append(axes, closestAxis)
	}

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

	dir := center.Sub(polygonCentroidPoint(p.WorldVertices))
	if dot(dir, smallestAxis) < 0 {
		smallestAxis = smallestAxis.Neg()
	}

	return true, clampMTV(smallestAxis.Mul(minOverlap))
}

func closestVertex(center notamath.Po2, vertices []notamath.Po2) notamath.Po2 {
	if len(vertices) == 0 {
		return notamath.Po2{}
	}

	closest := vertices[0]
	bestDist := center.DistanceSquared(closest)
	for _, v := range vertices[1:] {
		dist := center.DistanceSquared(v)
		if dist < bestDist {
			bestDist = dist
			closest = v
		}
	}
	return closest
}

func polygonCentroidPoint(vertices []notamath.Po2) notamath.Po2 {
	if len(vertices) == 0 {
		return notamath.Po2{}
	}

	var x, y float32
	for _, v := range vertices {
		x += v.X
		y += v.Y
	}

	inv := 1 / float32(len(vertices))
	return notamath.Po2{X: x * inv, Y: y * inv}
}

func clampMTV(mtv notamath.Vec2) notamath.Vec2 {
	if mtv.Len() > mTVTravelDistance {
		return mtv.Normalize().Mul(mTVTravelDistance)
	}
	return mtv
}

// SetMaximumMTVTravelDistance sets the maximum travel distance for the MTV calculation per frame. the default is 1.0
// Use at your own risk, numbers too high or too low may cause instability
func SetMaximumMTVTravelDistance(amount float32) {
	mTVTravelDistance = amount
}
