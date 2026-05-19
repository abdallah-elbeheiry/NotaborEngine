package notarender

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notageometry"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
	"NotaborEngine/notatomic"
	"fmt"
	"sync"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Vertex2D defines a 2D vertex with position, color, UV, and local position
type Vertex2D struct {
	Pos      notamath.Po2
	Color    notacolor.Color
	UV       notamath.Vec2
	LocalPos notamath.Po2
}

type DrawOrder struct {
	Vertices []Vertex2D
	Texture  *notatexture.Texture
	Shader   *notashader.Shader
	Material *notashader.Material
}

type PolygonRenderData struct {
	Vertices []Vertex2D
	Color    notacolor.Color
}

// VulkanRenderer manages Vulkan rendering using SDL3 GPU API
type VulkanRenderer struct {
	mu     sync.Mutex
	Orders []DrawOrder

	FrameID notatomic.UInt64

	currentShader   *notashader.Shader
	currentTexture  *notatexture.Texture
	currentMaterial *notashader.Material

	DefaultShader *notashader.Shader
}

// Renderer is the primary rendering interface using SDL3 GPU (Vulkan/Metal/D3D12)
type Renderer = VulkanRenderer

// Backend provides low-level GPU buffer and device management
type Backend = VulkanBackend

// VulkanBackend handles Vulkan device, command buffers, and GPU memory management
type VulkanBackend struct {
	Device               *sdl.GPUDevice
	VertexBuffer         *sdl.GPUBuffer
	TransferBuffer       *sdl.GPUTransferBuffer
	DefaultSampler       *sdl.GPUSampler
	WhiteTexture         *notatexture.Texture
	MaxVertices          int
	CurrentVertexCount   int
	vertexData           []Vertex2D
	windowWidth          uint32
	windowHeight         uint32
	lastGraphicsPipeline *sdl.GPUGraphicsPipeline
}

// InitVulkan initializes the Vulkan backend with SDL3 GPU API
func (b *VulkanBackend) Init(windowWidth, windowHeight uint32) error {
	b.windowWidth = windowWidth
	b.windowHeight = windowHeight
	b.MaxVertices = 100_000

	// Set GPU driver hints before creating device
	err := sdl.SetHint(sdl.HINT_GPU_DRIVER, "vulkan")
	if err != nil {
		return fmt.Errorf("failed to set GPU driver hint: %w", err)
	}

	device, err := sdl.CreateGPUDevice(sdl.GPU_SHADERFORMAT_SPIRV, true, "NotaborEngine")
	if err != nil {
		return fmt.Errorf("failed to create GPU device: %w", err)
	}
	b.Device = device

	// Vertex buffer
	vertexBufferSize := uint32(b.MaxVertices * int(unsafe.Sizeof(Vertex2D{})))
	b.VertexBuffer, err = b.Device.CreateBuffer(&sdl.GPUBufferCreateInfo{
		Usage: sdl.GPU_BUFFERUSAGE_VERTEX,
		Size:  vertexBufferSize,
	})
	if err != nil {
		return fmt.Errorf("failed to create vertex buffer: %w", err)
	}

	// Transfer buffer
	b.TransferBuffer, err = b.Device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  vertexBufferSize,
	})
	if err != nil {
		return fmt.Errorf("failed to create transfer buffer: %w", err)
	}

	b.DefaultSampler, err = b.Device.CreateSampler(&sdl.GPUSamplerCreateInfo{
		MinFilter:    sdl.GPU_FILTER_LINEAR,
		MagFilter:    sdl.GPU_FILTER_LINEAR,
		MipmapMode:   sdl.GPU_SAMPLERMIPMAPMODE_LINEAR,
		AddressModeU: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
		AddressModeV: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
		AddressModeW: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
	})
	if err != nil {
		return fmt.Errorf("failed to create default sampler: %w", err)
	}

	whiteTexture := &notatexture.Texture{
		Width:     1,
		Height:    1,
		ImageData: []byte{255, 255, 255, 255},
	}
	if err := whiteTexture.CreateGPUTexture(b.Device); err != nil {
		return fmt.Errorf("failed to create default white texture: %w", err)
	}
	b.WhiteTexture = whiteTexture

	b.vertexData = make([]Vertex2D, 0, b.MaxVertices)
	return nil
}

// BeginFrame acquires a command buffer for the frame
func (b *VulkanBackend) BeginFrame() (*sdl.GPUCommandBuffer, error) {
	cmdBuf, err := b.Device.AcquireCommandBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire command buffer: %w", err)
	}
	return cmdBuf, nil
}

// UploadVertexData uploads vertex data to GPU using a copy pass
func (b *VulkanBackend) UploadVertexData(cmdBuf *sdl.GPUCommandBuffer, vertices []Vertex2D) error {
	if len(vertices) == 0 {
		return nil
	}

	b.CurrentVertexCount = len(vertices)
	vertexSize := uint32(len(vertices) * int(unsafe.Sizeof(Vertex2D{})))

	mapped, err := b.Device.MapTransferBuffer(b.TransferBuffer, false)
	if err != nil {
		return err
	}

	copy(unsafe.Slice((*Vertex2D)(unsafe.Pointer(mapped)), len(vertices)), vertices)
	b.Device.UnmapTransferBuffer(b.TransferBuffer)

	copyPass := cmdBuf.BeginCopyPass()
	copyPass.UploadToGPUBuffer(
		&sdl.GPUTransferBufferLocation{TransferBuffer: b.TransferBuffer, Offset: 0},
		&sdl.GPUBufferRegion{Buffer: b.VertexBuffer, Offset: 0, Size: vertexSize},
		false,
	)
	copyPass.End()

	return nil
}

// AcquireSwapchainTexture gets the current backbuffer texture for rendering
func (b *VulkanBackend) AcquireSwapchainTexture(cmdBuf *sdl.GPUCommandBuffer, window *sdl.Window) (*sdl.GPUTexture, error) {
	swapchainTex, err := cmdBuf.WaitAndAcquireGPUSwapchainTexture(window)
	if err != nil {
		return nil, err
	}
	if swapchainTex == nil {
		return nil, fmt.Errorf("swapchain texture is nil")
	}
	return swapchainTex.Texture, nil
}

// Shutdown cleans up Vulkan resources
func (b *VulkanBackend) Shutdown() {
	if b.VertexBuffer != nil {
		b.Device.ReleaseBuffer(b.VertexBuffer)
	}
	if b.TransferBuffer != nil {
		b.Device.ReleaseTransferBuffer(b.TransferBuffer)
	}
	if b.DefaultSampler != nil {
		b.Device.ReleaseSampler(b.DefaultSampler)
	}
	if b.WhiteTexture != nil {
		b.WhiteTexture.Delete()
	}
	if b.Device != nil {
		b.Device.Destroy()
	}
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
func (r *VulkanRenderer) SubmitPolygon(poly *notageometry.Polygon, model notamath.Mat3, color notacolor.Color, tex *notatexture.Texture, shader *notashader.Shader, material *notashader.Material) {
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

	r.mu.Lock()
	defer r.mu.Unlock()
	r.Orders = append(r.Orders, DrawOrder{
		Vertices: tris,
		Texture:  tex,
		Shader:   shader,
		Material: material,
	})
}

// draw batch structure
type drawBatch struct {
	shader     *notashader.Shader
	material   *notashader.Material
	texture    *notatexture.Texture
	startIndex int
	vertCount  int
}

// Flush submits all queued orders and renders them
func (r *VulkanRenderer) Flush(backend *VulkanBackend, cmdBuf *sdl.GPUCommandBuffer, window *sdl.Window) error {
	r.mu.Lock()
	orders := make([]DrawOrder, len(r.Orders))
	copy(orders, r.Orders)
	r.Orders = r.Orders[:0]
	r.mu.Unlock()

	// Render-pass state does not persist across frames, so cached bindings
	// must be invalidated before we start issuing draw calls for a new pass.
	r.currentShader = nil
	r.currentMaterial = nil
	r.currentTexture = nil

	if len(orders) == 0 {
		return cmdBuf.Cancel()
	}

	backend.vertexData = backend.vertexData[:0]
	// Group consecutive orders with same shader+texture
	var batches []drawBatch
	var current *drawBatch

	for _, order := range orders {
		shader := order.Shader
		material := order.Material
		if material != nil && material.Shader != nil {
			shader = material.Shader
		}
		if shader == nil {
			shader = r.DefaultShader
		}

		canBatch := current != nil &&
			current.shader == shader &&
			current.material == material &&
			current.texture == order.Texture
		if !canBatch {
			batches = append(batches, drawBatch{
				shader:     shader,
				material:   material,
				texture:    order.Texture,
				startIndex: len(backend.vertexData),
				vertCount:  0,
			})
			current = &batches[len(batches)-1]
		}

		backend.vertexData = append(backend.vertexData, order.Vertices...)
		current.vertCount += len(order.Vertices)
	}

	// Upload vertex data in a copy pass
	if err := backend.UploadVertexData(cmdBuf, backend.vertexData); err != nil {
		_ = cmdBuf.Cancel()
		return err
	}

	swapchainTexture, err := backend.AcquireSwapchainTexture(cmdBuf, window)
	if err != nil {
		_ = cmdBuf.Cancel()
		return err
	}

	colorTarget := sdl.GPUColorTargetInfo{
		Texture:    swapchainTexture,
		ClearColor: sdl.FColor{R: 0.0, G: 0.0, B: 0.0, A: 1.0},
		LoadOp:     sdl.GPU_LOADOP_CLEAR,
		StoreOp:    sdl.GPU_STOREOP_STORE,
	}

	renderPass := cmdBuf.BeginRenderPass([]sdl.GPUColorTargetInfo{colorTarget}, nil)
	if renderPass == nil {
		_ = cmdBuf.Cancel()
		return fmt.Errorf("failed to begin render pass")
	}

	// Draw batches
	for _, batch := range batches {
		// Bind shader pipeline if needed
		if batch.shader != r.currentShader {
			batch.shader.Bind(renderPass)
			r.currentShader = batch.shader
		}

		// Apply material (uniforms, etc.) if it has changed
		if batch.material != r.currentMaterial || batch.texture != r.currentTexture {
			if batch.material != nil {
				batch.material.Apply(cmdBuf)
			}
			r.currentMaterial = batch.material
		}

		// Bind texture if needed
		boundTexture := batch.texture
		if boundTexture == nil || boundTexture.GPUTexture == nil {
			boundTexture = backend.WhiteTexture
		}
		if boundTexture != r.currentTexture {
			if boundTexture != nil {
				boundTexture.Bind(renderPass, backend.DefaultSampler)
			}
			r.currentTexture = boundTexture
		}

		// Bind and draw
		renderPass.BindVertexBuffers([]sdl.GPUBufferBinding{
			{Buffer: backend.VertexBuffer, Offset: 0},
		})

		renderPass.DrawPrimitives(uint32(batch.vertCount), 1, uint32(batch.startIndex), 0)
	}

	// End render pass
	renderPass.End()

	return cmdBuf.Submit()
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
