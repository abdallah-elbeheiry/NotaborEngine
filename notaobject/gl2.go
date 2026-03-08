package notaobject

import (
	"NotaborEngine/notamath"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Vertex2D struct {
	Pos      notamath.Po2
	Color    Color
	UV       notamath.Vec2
	LocalPos notamath.Po2
}

type DrawOrder struct {
	Vertices []Vertex2D
	Texture  *Texture
	Shader   *Shader
}
type Renderer struct {
	Orders []DrawOrder

	currentShader  *Shader
	currentTexture *Texture

	DefaultShader *Shader
}

func (r *Renderer) Submit(p *Polygon, model notamath.Mat3, tex *Texture, shader *Shader) {
	var temp []DrawOrder
	p.AddToOrders(model, &temp)

	for _, order := range temp {
		tris := Triangulate2D(order.Vertices)
		if len(tris) == 0 {
			continue
		}

		r.Orders = append(r.Orders, DrawOrder{
			Vertices: tris,
			Texture:  tex,
			Shader:   shader,
		})
	}
}

type format2D struct {
	dimension int32 // should be 2
	stride    int32
}

type GLBackend struct {
	vao    uint32
	vbo    uint32
	format format2D
}

func (b *GLBackend) Init() {
	b.format = format2D{
		dimension: 2,
		stride:    int32(unsafe.Sizeof(Vertex2D{})),
	}

	gl.CreateVertexArrays(1, &b.vao)
	gl.CreateBuffers(1, &b.vbo)

	// Position Attribute (Location 0)
	gl.VertexArrayVertexBuffer(b.vao, 0, b.vbo, 0, b.format.stride)
	gl.VertexArrayAttribFormat(b.vao, 0, 2, gl.FLOAT, false, 0)
	gl.VertexArrayAttribBinding(b.vao, 0, 0)
	gl.EnableVertexArrayAttrib(b.vao, 0)

	// Color Attribute (Location 1)
	colorOffset := uint32(unsafe.Sizeof(notamath.Po2{}))
	gl.VertexArrayAttribFormat(b.vao, 1, 4, gl.FLOAT, false, colorOffset)
	gl.VertexArrayAttribBinding(b.vao, 1, 0)
	gl.EnableVertexArrayAttrib(b.vao, 1)

	// UV Attribute (Location 2)
	uvOffset := uint32(unsafe.Sizeof(notamath.Po2{}) + unsafe.Sizeof(Color{}))
	gl.VertexArrayAttribFormat(b.vao, 2, 2, gl.FLOAT, false, uvOffset)
	gl.VertexArrayAttribBinding(b.vao, 2, 0)
	gl.EnableVertexArrayAttrib(b.vao, 2)

	// LocalPos Attribute (Location 3)
	localOffset := uint32(unsafe.Sizeof(notamath.Po2{}) + unsafe.Sizeof(Color{}) + unsafe.Sizeof(notamath.Vec2{}))
	gl.VertexArrayAttribFormat(b.vao, 3, 2, gl.FLOAT, false, localOffset)
	gl.VertexArrayAttribBinding(b.vao, 3, 0)
	gl.EnableVertexArrayAttrib(b.vao, 3)
}

func (b *GLBackend) BindVao() {
	gl.BindVertexArray(b.vao)
}

func (b *GLBackend) UploadData(vertices interface{}) {
	verts := vertices.([]Vertex2D)
	gl.NamedBufferData(b.vbo, len(verts)*int(b.format.stride), gl.Ptr(verts), gl.DYNAMIC_DRAW)
}

func (r *Renderer) Flush(backend *GLBackend) {
	if len(r.Orders) == 0 {
		return
	}

	backend.BindVao()

	for _, order := range r.Orders {
		shader := order.Shader
		if shader == nil {
			shader = r.DefaultShader
		}

		if shader != r.currentShader {
			shader.Bind()
			r.currentShader = shader
		}

		if order.Texture != nil && order.Texture != r.currentTexture {
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, order.Texture.ID)
			r.currentTexture = order.Texture
		}

		backend.UploadData(order.Vertices)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(order.Vertices)))
	}

	r.Orders = r.Orders[:0]
}

func Triangulate2D(polygon []Vertex2D) []Vertex2D {
	n := len(polygon)
	if n < 3 {
		return nil
	}

	verts := append([]Vertex2D{}, polygon...)

	// Enforce CCW winding (using Pos for math)
	if !isCCWVertices(verts) {
		for i, j := 0, len(verts)-1; i < j; i, j = i+1, j-1 {
			verts[i], verts[j] = verts[j], verts[i]
		}
	}

	var result []Vertex2D

	for len(verts) > 3 {
		earFound := false

		for i := 0; i < len(verts); i++ {
			prev := verts[(i-1+len(verts))%len(verts)]
			curr := verts[i]
			next := verts[(i+1)%len(verts)]

			if isEarVertex(prev, curr, next, verts) {
				result = append(result, prev, curr, next)

				verts = append(verts[:i], verts[i+1:]...)
				earFound = true
				break
			}
		}

		if !earFound {
			return nil
		}
	}

	result = append(result, verts[0], verts[1], verts[2])
	return result
}

// Helper functions to use Vertex2D for triangulation math
func isCCWVertices(poly []Vertex2D) bool {
	var area float32
	for i := 0; i < len(poly); i++ {
		a := poly[i].Pos
		b := poly[(i+1)%len(poly)].Pos
		area += (b.X - a.X) * (b.Y + a.Y)
	}
	return area < 0
}

func isEarVertex(prev, curr, next Vertex2D, poly []Vertex2D) bool {
	if notamath.Orient(prev.Pos, curr.Pos, next.Pos) <= 0 {
		return false
	}
	for _, p := range poly {
		if p.Pos == prev.Pos || p.Pos == curr.Pos || p.Pos == next.Pos {
			continue
		}
		if PointInTriangle(p.Pos, prev.Pos, curr.Pos, next.Pos) {
			return false
		}
	}
	return true
}
