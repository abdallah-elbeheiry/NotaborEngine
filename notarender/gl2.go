package notarender

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"NotaborEngine/notatomic"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Renderer struct {
	Orders []DrawOrder

	FrameID notatomic.UInt64

	currentShader  *notashader.Shader
	currentTexture *notatexture.Texture

	DefaultShader *notashader.Shader
}

// BuildPolygonRenderData converts geometry polygon to renderable vertices
func BuildPolygonRenderData(poly *notageometry.Polygon, color notacolor.Color) *PolygonRenderData {
	if len(poly.Points) < 3 {
		return nil
	}

	verts := make([]Vertex2D, len(poly.Points))

	minX, minY := poly.Points[0].X, poly.Points[0].Y
	maxX, maxY := minX, minY
	for _, p := range poly.Points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	rangeX := maxX - minX
	rangeY := maxY - minY

	for i, p := range poly.Points {
		verts[i].Pos = p
		verts[i].Color = color

		if rangeX > 0 {
			verts[i].UV.X = (p.X - minX) / rangeX
			verts[i].LocalPos.X = verts[i].UV.X - 0.5
		} else {
			verts[i].UV.X = 0.5
			verts[i].LocalPos.X = 0
		}

		if rangeY > 0 {
			verts[i].UV.Y = (p.Y - minY) / rangeY
			verts[i].LocalPos.Y = verts[i].UV.Y - 0.5
		} else {
			verts[i].UV.Y = 0.5
			verts[i].LocalPos.Y = 0
		}
	}

	return &PolygonRenderData{
		Vertices: verts,
		Color:    color,
	}
}

// SubmitPolygon transforms and queues polygon for rendering
func (r *Renderer) SubmitPolygon(poly *notageometry.Polygon, model notamath.Mat3, color notacolor.Color, tex *notatexture.Texture, shader *notashader.Shader) {
	renderData := BuildPolygonRenderData(poly, color)
	if renderData == nil {
		return
	}

	for i := range renderData.Vertices {
		renderData.Vertices[i].Pos = model.TransformPo2(renderData.Vertices[i].Pos)
	}

	tris := Triangulate2D(renderData.Vertices)
	if len(tris) == 0 {
		return
	}

	r.Orders = append(r.Orders, DrawOrder{
		Vertices: tris,
		Texture:  tex,
		Shader:   shader,
	})
}

// GL backend for vertex batching
type GLBackend struct {
	vao         uint32
	vbo         uint32
	stride      int32
	maxVertices int
	batchBuffer []Vertex2D
}

func (b *GLBackend) Init() {
	b.stride = int32(unsafe.Sizeof(Vertex2D{}))
	b.maxVertices = 100_000
	b.batchBuffer = make([]Vertex2D, 0, b.maxVertices)

	gl.CreateVertexArrays(1, &b.vao)
	gl.CreateBuffers(1, &b.vbo)
	gl.NamedBufferData(b.vbo, b.maxVertices*int(b.stride), nil, gl.STREAM_DRAW)

	// Bind VAO
	gl.VertexArrayVertexBuffer(b.vao, 0, b.vbo, 0, b.stride)

	// Position (loc 0)
	gl.VertexArrayAttribFormat(b.vao, 0, 2, gl.FLOAT, false, 0)
	gl.VertexArrayAttribBinding(b.vao, 0, 0)
	gl.EnableVertexArrayAttrib(b.vao, 0)

	// Color (loc 1)
	offsetColor := uint32(unsafe.Sizeof(notamath.Po2{}))
	gl.VertexArrayAttribFormat(b.vao, 1, 4, gl.FLOAT, false, offsetColor)
	gl.VertexArrayAttribBinding(b.vao, 1, 0)
	gl.EnableVertexArrayAttrib(b.vao, 1)

	// UV (loc 2)
	offsetUV := offsetColor + uint32(unsafe.Sizeof(notacolor.Color{}))
	gl.VertexArrayAttribFormat(b.vao, 2, 2, gl.FLOAT, false, offsetUV)
	gl.VertexArrayAttribBinding(b.vao, 2, 0)
	gl.EnableVertexArrayAttrib(b.vao, 2)

	// LocalPos (loc 3)
	offsetLocal := offsetUV + uint32(unsafe.Sizeof(notamath.Vec2{}))
	gl.VertexArrayAttribFormat(b.vao, 3, 2, gl.FLOAT, false, offsetLocal)
	gl.VertexArrayAttribBinding(b.vao, 3, 0)
	gl.EnableVertexArrayAttrib(b.vao, 3)
}

func (b *GLBackend) BindVao() {
	gl.BindVertexArray(b.vao)
}

// Draw batch structure
type drawBatch struct {
	shader     *notashader.Shader
	texture    *notatexture.Texture
	startIndex int
	vertCount  int
}

// Flush submits all queued orders and clears them
func (r *Renderer) Flush(backend *GLBackend) {
	if len(r.Orders) == 0 {
		return
	}

	// Clear once per frame
	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	backend.BindVao()
	backend.batchBuffer = backend.batchBuffer[:0]

	// Group consecutive orders with same shader+texture
	var batches []drawBatch
	var current *drawBatch

	for _, order := range r.Orders {
		shader := order.Shader
		if shader == nil {
			shader = r.DefaultShader
		}

		canBatch := current != nil && current.shader == shader && current.texture == order.Texture
		if !canBatch {
			batches = append(batches, drawBatch{
				shader:     shader,
				texture:    order.Texture,
				startIndex: len(backend.batchBuffer),
				vertCount:  0,
			})
			current = &batches[len(batches)-1]
		}

		backend.batchBuffer = append(backend.batchBuffer, order.Vertices...)
		current.vertCount += len(order.Vertices)
	}

	// Upload vertex buffer
	if len(backend.batchBuffer) > 0 {
		gl.NamedBufferSubData(backend.vbo, 0, len(backend.batchBuffer)*int(backend.stride), gl.Ptr(backend.batchBuffer))
	}

	// Draw batches
	for _, batch := range batches {
		if batch.shader != r.currentShader {
			batch.shader.Bind()
			r.currentShader = batch.shader
		}

		if batch.texture != r.currentTexture {
			if batch.texture != nil {
				batch.shader.SetUniform(notashader.UseTexture, true)
				gl.ActiveTexture(gl.TEXTURE0)
				gl.BindTexture(gl.TEXTURE_2D, batch.texture.ID)
			} else {
				batch.shader.SetUniform(notashader.UseTexture, false)
			}
			r.currentTexture = batch.texture
		}

		gl.DrawArrays(gl.TRIANGLES, int32(batch.startIndex), int32(batch.vertCount))
	}

	// Clear orders after flush
	r.Orders = r.Orders[:0]
}

// Triangulate polygon into triangles
func Triangulate2D(poly []Vertex2D) []Vertex2D {
	n := len(poly)
	if n < 3 {
		return nil
	}

	verts := append([]Vertex2D{}, poly...)
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
		if notageometry.PointInTriangle(p.Pos, prev.Pos, curr.Pos, next.Pos) {
			return false
		}
	}
	return true
}
