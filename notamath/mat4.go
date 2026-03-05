package notamath

import "math"

type Mat4 struct {
	M [16]float32
}

// Mat4Identity Matrix creation functions
func Mat4Identity() Mat4 {
	return Mat4{M: [16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}}
}

func Mat4Translation(t Vec3) Mat4 {
	return Mat4{M: [16]float32{
		1, 0, 0, t.X,
		0, 1, 0, t.Y,
		0, 0, 1, t.Z,
		0, 0, 0, 1,
	}}
}

func Mat4Scale(s Vec3) Mat4 {
	return Mat4{M: [16]float32{
		s.X, 0, 0, 0,
		0, s.Y, 0, 0,
		0, 0, s.Z, 0,
		0, 0, 0, 1,
	}}
}

func Mat4RotationAxisAngle(axis Vec3, angle float32) Mat4 {
	k := axis.Normalize()
	c := float32(math.Cos(float64(angle)))
	s := float32(math.Sin(float64(angle)))
	t := 1 - c

	// Compute rotation matrix components
	xx, xy, xz := k.X*k.X*t, k.X*k.Y*t, k.X*k.Z*t
	yx, yy, yz := k.Y*k.X*t, k.Y*k.Y*t, k.Y*k.Z*t
	zx, zy, zz := k.Z*k.X*t, k.Z*k.Y*t, k.Z*k.Z*t

	xs, ys, zs := k.X*s, k.Y*s, k.Z*s

	return Mat4{M: [16]float32{
		c + xx, xy - zs, xz + ys, 0,
		yx + zs, c + yy, yz - xs, 0,
		zx - ys, zy + xs, c + zz, 0,
		0, 0, 0, 1,
	}}
}

// Mul Matrix multiplication
func (m Mat4) Mul(b Mat4) Mat4 {
	var r Mat4

	// Unrolled loop for clarity and performance
	r.M[0] = m.M[0]*b.M[0] + m.M[1]*b.M[4] + m.M[2]*b.M[8] + m.M[3]*b.M[12]
	r.M[1] = m.M[0]*b.M[1] + m.M[1]*b.M[5] + m.M[2]*b.M[9] + m.M[3]*b.M[13]
	r.M[2] = m.M[0]*b.M[2] + m.M[1]*b.M[6] + m.M[2]*b.M[10] + m.M[3]*b.M[14]
	r.M[3] = m.M[0]*b.M[3] + m.M[1]*b.M[7] + m.M[2]*b.M[11] + m.M[3]*b.M[15]

	r.M[4] = m.M[4]*b.M[0] + m.M[5]*b.M[4] + m.M[6]*b.M[8] + m.M[7]*b.M[12]
	r.M[5] = m.M[4]*b.M[1] + m.M[5]*b.M[5] + m.M[6]*b.M[9] + m.M[7]*b.M[13]
	r.M[6] = m.M[4]*b.M[2] + m.M[5]*b.M[6] + m.M[6]*b.M[10] + m.M[7]*b.M[14]
	r.M[7] = m.M[4]*b.M[3] + m.M[5]*b.M[7] + m.M[6]*b.M[11] + m.M[7]*b.M[15]

	r.M[8] = m.M[8]*b.M[0] + m.M[9]*b.M[4] + m.M[10]*b.M[8] + m.M[11]*b.M[12]
	r.M[9] = m.M[8]*b.M[1] + m.M[9]*b.M[5] + m.M[10]*b.M[9] + m.M[11]*b.M[13]
	r.M[10] = m.M[8]*b.M[2] + m.M[9]*b.M[6] + m.M[10]*b.M[10] + m.M[11]*b.M[14]
	r.M[11] = m.M[8]*b.M[3] + m.M[9]*b.M[7] + m.M[10]*b.M[11] + m.M[11]*b.M[15]

	r.M[12] = m.M[12]*b.M[0] + m.M[13]*b.M[4] + m.M[14]*b.M[8] + m.M[15]*b.M[12]
	r.M[13] = m.M[12]*b.M[1] + m.M[13]*b.M[5] + m.M[14]*b.M[9] + m.M[15]*b.M[13]
	r.M[14] = m.M[12]*b.M[2] + m.M[13]*b.M[6] + m.M[14]*b.M[10] + m.M[15]*b.M[14]
	r.M[15] = m.M[12]*b.M[3] + m.M[13]*b.M[7] + m.M[14]*b.M[11] + m.M[15]*b.M[15]

	return r
}

// TransformPo3 Transforms a point according to matrix
func (m Mat4) TransformPo3(p Po3) Po3 {
	return Po3{
		X: m.M[0]*p.X + m.M[1]*p.Y + m.M[2]*p.Z + m.M[3],
		Y: m.M[4]*p.X + m.M[5]*p.Y + m.M[6]*p.Z + m.M[7],
		Z: m.M[8]*p.X + m.M[9]*p.Y + m.M[10]*p.Z + m.M[11],
	}
}

func (m Mat4) TransformVec3(v Vec3) Vec3 {
	return Vec3{
		X: m.M[0]*v.X + m.M[1]*v.Y + m.M[2]*v.Z,
		Y: m.M[4]*v.X + m.M[5]*v.Y + m.M[6]*v.Z,
		Z: m.M[8]*v.X + m.M[9]*v.Y + m.M[10]*v.Z,
	}
}

// Mat4TRS creates a Transform-Rotate-Scale matrix (applied in that order)
func Mat4TRS(pos Vec3, axis Vec3, angle float32, scale Vec3) Mat4 {
	return Mat4Translation(pos).
		Mul(Mat4RotationAxisAngle(axis, angle)).
		Mul(Mat4Scale(scale))
}

// Mat4Perspective Projection matrices
func Mat4Perspective(fovY, aspect, near, far float32) Mat4 {
	f := float32(1.0 / math.Tan(float64(fovY*0.5)))

	// Standard perspective projection matrix
	return Mat4{M: [16]float32{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) / (near - far), (2 * far * near) / (near - far),
		0, 0, -1, 0,
	}}
}

func Mat4LookAt(eye Vec3, center Vec3, up Vec3) Mat4 {
	forward := center.Sub(eye).Normalize()
	right := forward.Cross(up).Normalize()
	up = right.Cross(forward) // Recompute up to ensure orthogonality

	return Mat4{M: [16]float32{
		right.X, right.Y, right.Z, -right.Dot(eye),
		up.X, up.Y, up.Z, -up.Dot(eye),
		-forward.X, -forward.Y, -forward.Z, forward.Dot(eye),
		0, 0, 0, 1,
	}}
}

func Mat4Ortho(left, right, bottom, top, near, far float32) Mat4 {
	// Standard orthographic projection matrix
	w, h := right-left, top-bottom
	d := far - near

	return Mat4{M: [16]float32{
		2 / w, 0, 0, -(right + left) / w,
		0, 2 / h, 0, -(top + bottom) / h,
		0, 0, -2 / d, -(far + near) / d,
		0, 0, 0, 1,
	}}
}

func invertMat4Linear3x3(m Mat4) (r00, r01, r02, r10, r11, r12, r20, r21, r22 float32, ok bool) {
	// Extract 3x3 linear part
	a00, a01, a02 := m.M[0], m.M[1], m.M[2]
	a10, a11, a12 := m.M[4], m.M[5], m.M[6]
	a20, a21, a22 := m.M[8], m.M[9], m.M[10]

	// Compute determinant of 3x3
	det := a00*(a11*a22-a12*a21) -
		a01*(a10*a22-a12*a20) +
		a02*(a10*a21-a11*a20)

	if det == 0 {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, false
	}

	invDet := 1 / det

	// Inverse of 3x3
	r00 = (a11*a22 - a12*a21) * invDet
	r01 = (a02*a21 - a01*a22) * invDet
	r02 = (a01*a12 - a02*a11) * invDet

	r10 = (a12*a20 - a10*a22) * invDet
	r11 = (a00*a22 - a02*a20) * invDet
	r12 = (a02*a10 - a00*a12) * invDet

	r20 = (a10*a21 - a11*a20) * invDet
	r21 = (a01*a20 - a00*a21) * invDet
	r22 = (a00*a11 - a01*a10) * invDet

	return r00, r01, r02, r10, r11, r12, r20, r21, r22, true
}

// InverseAffine Matrix inversion
func (m Mat4) InverseAffine() Mat4 {
	r00, r01, r02, r10, r11, r12, r20, r21, r22, ok := invertMat4Linear3x3(m)
	if !ok {
		return Mat4Identity()
	}

	// Apply inverse to translation
	tx, ty, tz := m.M[3], m.M[7], m.M[11]
	itx := -(r00*tx + r01*ty + r02*tz)
	ity := -(r10*tx + r11*ty + r12*tz)
	itz := -(r20*tx + r21*ty + r22*tz)

	return Mat4{M: [16]float32{
		r00, r01, r02, itx,
		r10, r11, r12, ity,
		r20, r21, r22, itz,
		0, 0, 0, 1,
	}}
}

func (m Mat4) NormalMatrix() Mat3 {
	r00, r01, r02, r10, r11, r12, r20, r21, r22, ok := invertMat4Linear3x3(m)
	if !ok {
		return Mat3Identity()
	}

	// Transpose for normal matrix (inverse-transpose)
	return Mat3{M: [9]float32{
		r00, r10, r20,
		r01, r11, r21,
		r02, r12, r22,
	}}
}
