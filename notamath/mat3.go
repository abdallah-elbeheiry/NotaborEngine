package notamath

import (
	"fmt"
	"math"

	"github.com/viterin/vek/vek32"
)

type Mat3 struct {
	M [9]float32
}

func Mat3Identity() Mat3 {
	return Mat3{M: [9]float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}}
}

func Mat3Translation(t Vec2) Mat3 {
	return Mat3{M: [9]float32{
		1, 0, t.X,
		0, 1, t.Y,
		0, 0, 1,
	}}
}

func Mat3Scale(s Vec2) Mat3 {
	return Mat3{M: [9]float32{
		s.X, 0, 0,
		0, s.Y, 0,
		0, 0, 1,
	}}
}

func Mat3Rotation(rad float32) Mat3 {
	c := float32(math.Cos(float64(rad)))
	s := float32(math.Sin(float64(rad)))

	return Mat3{M: [9]float32{
		c, -s, 0,
		s, c, 0,
		0, 0, 1,
	}}
}

func Mat3Shear(kx, ky float32) Mat3 {
	return Mat3{M: [9]float32{
		1, kx, 0,
		ky, 1, 0,
		0, 0, 1,
	}}
}

func Mat3TRS(pos Vec2, rot float32, scale Vec2) Mat3 {
	c := float32(math.Cos(float64(rot)))
	s := float32(math.Sin(float64(rot)))

	return Mat3{M: [9]float32{
		scale.X * c, -scale.X * s, pos.X,
		scale.Y * s, scale.Y * c, pos.Y,
		0, 0, 1,
	}}
}

func (m Mat3) Mul(b Mat3) Mat3 {
	var r Mat3
	vek32.MatMul_Into(r.M[:], m.M[:], b.M[:], 3)
	return r
}

func (m Mat3) TransformPo2(p Po2) Po2 {
	return Po2{
		X: m.M[0]*p.X + m.M[1]*p.Y + m.M[2],
		Y: m.M[3]*p.X + m.M[4]*p.Y + m.M[5],
	}
}

func (m Mat3) TransformVec2(v Vec2) Vec2 {
	return Vec2{
		X: m.M[0]*v.X + m.M[1]*v.Y,
		Y: m.M[3]*v.X + m.M[4]*v.Y,
	}
}

func (m Mat3) Transpose() Mat3 {
	return Mat3{M: [9]float32{
		m.M[0], m.M[3], m.M[6],
		m.M[1], m.M[4], m.M[7],
		m.M[2], m.M[5], m.M[8],
	}}
}

func (m Mat3) Det() float32 {
	return m.M[0]*m.M[4] - m.M[1]*m.M[3]
}

func (m Mat3) InverseAffine() Mat3 {
	a, b, c := m.M[0], m.M[1], m.M[2]
	d, e, f := m.M[3], m.M[4], m.M[5]

	det := m.Det()
	if det == 0 {
		return Mat3Identity()
	}

	invDet := 1 / det

	return Mat3{M: [9]float32{
		e * invDet, -b * invDet, (b*f - e*c) * invDet,
		-d * invDet, a * invDet, (d*c - a*f) * invDet,
		0, 0, 1,
	}}
}

func (m Mat3) String() string {
	return fmt.Sprintf(
		"[%f %f %f\n %f %f %f\n %f %f %f]",
		m.M[0], m.M[1], m.M[2],
		m.M[3], m.M[4], m.M[5],
		m.M[6], m.M[7], m.M[8],
	)
}
