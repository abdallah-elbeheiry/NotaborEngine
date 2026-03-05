package notamath

type Transform3D struct {
	Position     Vec3
	RotationAxis Vec3 // Reference vector for rotation
	Rotation     float32
	Scale        Vec3
	Dirty        bool // true if matrix needs to be recomputed
	matrix       Mat4 // cached TRS matrix
	prevPosition Vec3
	prevRotation float32
	prevScale    Vec3
}

func NewTransform3D() Transform3D {
	return Transform3D{
		Scale:  Vec3{1, 1, 1},
		Dirty:  true,
		matrix: Mat4Identity(),
	}
}

func (t *Transform3D) SetPosition(p Vec3) {
	t.Position = p
	t.Dirty = true
}

func (t *Transform3D) SetRotationAxis(r Vec3) {
	t.RotationAxis = r.Normalize()
	t.Dirty = true
}

func (t *Transform3D) SetScale(s Vec3) {
	t.Scale = s
	t.Dirty = true
}

func (t *Transform3D) Matrix() Mat4 {
	if !t.Dirty {
		return t.matrix
	}

	tr := Mat4Translation(t.Position)

	rot := Mat4RotationAxisAngle(t.RotationAxis, t.Rotation)

	sc := Mat4Scale(t.Scale)

	t.matrix = tr.Mul(rot).Mul(sc)
	t.Dirty = false
	return t.matrix
}

func (t *Transform3D) TransformPo3(p Po3) Po3 {
	return t.Matrix().TransformPo3(p)
}

func (t *Transform3D) TransformVec3(v Vec3) Vec3 {
	return t.Matrix().TransformVec3(v)
}

func (t *Transform3D) TranslateBy(delta Vec3) {
	t.Position = t.Position.Add(delta)
	t.Dirty = true
}

func (t *Transform3D) RotateBy(delta float32) {
	t.Rotation += delta
	t.Dirty = true
}

func (t *Transform3D) ScaleBy(factor Vec3) {
	t.Scale = Vec3{t.Scale.X * factor.X, t.Scale.Y * factor.Y, t.Scale.Z * factor.Z}
	t.Dirty = true
}

func (t *Transform3D) Snapshot() {
	t.prevPosition = t.Position
	t.prevScale = t.Scale
	t.prevRotation = t.Rotation
}

func (t *Transform3D) InterpolatedMatrix(alpha float32) Mat4 {
	pos := t.prevPosition.Lerp(t.Position, alpha)
	scale := t.prevScale.Lerp(t.Scale, alpha)
	angle := lerpAngle(t.prevRotation, t.Rotation, alpha)

	return Mat4TRS(pos, t.RotationAxis, angle, scale)
}

type AxisMask uint8

const (
	AxisNone AxisMask = 0
	AxisX    AxisMask = 1 << iota // 001
	AxisY                         // 010
	AxisZ                         // 100

	AxisXY  = AxisX | AxisY
	AxisXZ  = AxisX | AxisZ
	AxisYZ  = AxisY | AxisZ
	AxisXYZ = AxisX | AxisY | AxisZ
)
