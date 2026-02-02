package notagl

import (
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Vertex2D struct {
	Pos   notamath.Po2
	Color notashader.Color
	UV    notamath.Vec2
}
type DrawOrder2D struct {
	Vertices []Vertex2D
}

type Renderer2D struct {
	Orders         []DrawOrder2D
	CurrentTexture *Texture // Track current texture
}

func (r *Renderer2D) Submit(p *Polygon, model notamath.Mat3) {
	var temp []DrawOrder2D

	p.AddToOrders(model, &temp)

	for _, order := range temp {
		tris := Triangulate2D(order.Vertices)
		if len(tris) == 0 {
			continue
		}
		r.Orders = append(r.Orders, DrawOrder2D{
			Vertices: tris,
		})
	}
}

type vertexFormat2D struct {
	dimension int32 // should be 2
	stride    int32
}

type GLBackend2D struct {
	vao    uint32
	vbo    uint32
	format vertexFormat2D
}

func (b *GLBackend2D) Init() {
	b.format = vertexFormat2D{
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
	uvOffset := uint32(unsafe.Sizeof(notamath.Po2{}) + unsafe.Sizeof(notashader.Color{}))
	gl.VertexArrayAttribFormat(b.vao, 2, 2, gl.FLOAT, false, uvOffset)
	gl.VertexArrayAttribBinding(b.vao, 2, 0)
	gl.EnableVertexArrayAttrib(b.vao, 2)
}

func (b *GLBackend2D) BindVao() {
	gl.BindVertexArray(b.vao)
}

func (b *GLBackend2D) UploadData(vertices interface{}) {
	verts := vertices.([]Vertex2D)
	gl.NamedBufferData(b.vbo, len(verts)*int(b.format.stride), gl.Ptr(verts), gl.DYNAMIC_DRAW)
}

func (r *Renderer2D) Flush(backend *GLBackend2D) {
	if len(r.Orders) == 0 {
		return
	}

	var flat []Vertex2D
	for _, order := range r.Orders {
		flat = append(flat, order.Vertices...)
	}

	if len(flat) == 0 {
		return
	}

	backend.UploadData(flat)
	backend.BindVao()
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(flat)))
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
