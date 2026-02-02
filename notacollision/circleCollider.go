package notacollision

import "NotaborEngine/notamath"

type CircleCollider struct {
	LocalCenter notamath.Po2
	LocalRadius float32

	worldCenter notamath.Po2
	worldRadius float32
}

func NewCircleCollider(center notamath.Po2, radius float32) *CircleCollider {
	return &CircleCollider{
		LocalCenter: center,
		LocalRadius: radius,
	}
}

func (c *CircleCollider) UpdateFromTransform(t *notamath.Transform2D) {
	c.worldCenter = t.TransformPoint(c.LocalCenter)

	avgScale := (t.Scale.X + t.Scale.Y) * 0.5
	c.worldRadius = c.LocalRadius * avgScale
}

func (c *CircleCollider) AABB() AABBCollider {
	return AABBCollider{
		Min: notamath.Vec2{
			X: c.worldCenter.X - c.worldRadius,
			Y: c.worldCenter.Y - c.worldRadius,
		},
		Max: notamath.Vec2{
			X: c.worldCenter.X + c.worldRadius,
			Y: c.worldCenter.Y + c.worldRadius,
		},
	}
}

func (c *CircleCollider) WorldCenter() notamath.Po2 {
	return c.worldCenter
}

func (c *CircleCollider) WorldRadius() float32 {
	return c.worldRadius
}
