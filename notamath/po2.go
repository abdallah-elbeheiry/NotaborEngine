package notamath

import "fmt"

type Po2 struct {
	X, Y float32
}

func (p Po2) Add(v Vec2) Po2 {
	return Po2{p.X + v.X, p.Y + v.Y}
}

func (p Po2) Sub(q Po2) Vec2 {
	return Vec2{p.X - q.X, p.Y - q.Y}
}

func (p Po2) DistanceSquared(q Po2) float32 {
	return p.Sub(q).LenSquared()
}

func (p Po2) Distance(q Po2) float32 {
	return p.Sub(q).Len()
}

func (p Po2) Equals(q Po2, eps float32) bool {
	return p.Sub(q).LenSquared() <= eps*eps
}

func (p Po2) String() string {
	return fmt.Sprintf("Point2(%f, %f)", p.X, p.Y)
}

func (p Po2) ToVec2() Vec2 {
	return Vec2{p.X, p.Y}
}

func Orient(a, b, c Po2) float32 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}
