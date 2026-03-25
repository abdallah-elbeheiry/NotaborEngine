package notarender

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"unsafe"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Renderer struct {
	Orders []DrawOrder

	currentShader  *notashader.Shader
	currentTexture *notatexture.Texture

	DefaultShader *notashader.Shader
}

// BuildPolygonRenderData converts a pure geometry polygon to renderable vertices
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

// SubmitPolygon transforms and submits polygon for rendering
func (r *Renderer) SubmitPolygon(poly *notageometry.Polygon, model notamath.Mat3, color notacolor.Color, tex *notatexture.Texture, shader *notashader.Shader) {
	renderData := BuildPolygonRenderData(poly, color)
	if renderData == nil {
		return
	}

	// Transform vertices
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

type format2D struct {
	dimension int32
	stride    int32
}

type GLBackend struct {
	vao         uint32
	vbo         uint32
	format      format2D
	maxVertices int
	batchBuffer []Vertex2D
}

func (b *GLBackend) Init() {
	b.format = format2D{
		dimension: 2,
		stride:    int32(unsafe.Sizeof(Vertex2D{})),
	}

	// Pre-allocate for 100k vertices
	b.maxVertices = 100000
	b.batchBuffer = make([]Vertex2D, 0, b.maxVertices)

	gl.CreateVertexArrays(1, &b.vao)
	gl.CreateBuffers(1, &b.vbo)

	// Pre-allocate persistent buffer
	gl.NamedBufferData(b.vbo, b.maxVertices*int(b.format.stride), nil, gl.STREAM_DRAW)

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
	uvOffset := uint32(unsafe.Sizeof(notamath.Po2{}) + unsafe.Sizeof(notacolor.Color{}))
	gl.VertexArrayAttribFormat(b.vao, 2, 2, gl.FLOAT, false, uvOffset)
	gl.VertexArrayAttribBinding(b.vao, 2, 0)
	gl.EnableVertexArrayAttrib(b.vao, 2)

	// LocalPos Attribute (Location 3)
	localOffset := uint32(unsafe.Sizeof(notamath.Po2{}) + unsafe.Sizeof(notacolor.Color{}) + unsafe.Sizeof(notamath.Vec2{}))
	gl.VertexArrayAttribFormat(b.vao, 3, 2, gl.FLOAT, false, localOffset)
	gl.VertexArrayAttribBinding(b.vao, 3, 0)
	gl.EnableVertexArrayAttrib(b.vao, 3)
}

func (b *GLBackend) BindVao() {
	gl.BindVertexArray(b.vao)
}

// Batch structure for grouping draw calls
type drawBatch struct {
	shader     *notashader.Shader
	texture    *notatexture.Texture
	startIndex int
	vertCount  int
}

func (r *Renderer) Flush(backend *GLBackend) {
	if len(r.Orders) == 0 {
		return
	}

	backend.BindVao()

	// Build batches by grouping consecutive orders with same shader+texture
	batches := make([]drawBatch, 0, 64)
	backend.batchBuffer = backend.batchBuffer[:0]

	var currentBatch *drawBatch

	for _, order := range r.Orders {
		shader := order.Shader
		if shader == nil {
			shader = r.DefaultShader
		}

		// Check if we can batch with previous
		canBatch := currentBatch != nil &&
			currentBatch.shader == shader &&
			currentBatch.texture == order.Texture

		if !canBatch {
			// Start new batch
			batches = append(batches, drawBatch{
				shader:     shader,
				texture:    order.Texture,
				startIndex: len(backend.batchBuffer),
				vertCount:  0,
			})
			currentBatch = &batches[len(batches)-1]
		}

		// Add vertices to batch buffer
		backend.batchBuffer = append(backend.batchBuffer, order.Vertices...)
		currentBatch.vertCount += len(order.Vertices)
	}

	// Upload all vertices in one call
	if len(backend.batchBuffer) > 0 {
		gl.NamedBufferSubData(backend.vbo, 0, len(backend.batchBuffer)*int(backend.format.stride), gl.Ptr(backend.batchBuffer))
	}

	// Draw all batches
	for _, batch := range batches {
		// Bind shader if changed
		if batch.shader != r.currentShader {
			batch.shader.Bind()
			r.currentShader = batch.shader
		}

		// Bind texture if changed
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

		// Single draw call for entire batch
		gl.DrawArrays(gl.TRIANGLES, int32(batch.startIndex), int32(batch.vertCount))
	}

	r.Orders = r.Orders[:0]
}

func Triangulate2D(polygon []Vertex2D) []Vertex2D {
	n := len(polygon)
	if n < 3 {
		return nil
	}

	verts := append([]Vertex2D{}, polygon...)

	// Enforce CCW winding
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
