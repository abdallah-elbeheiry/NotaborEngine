package notashader

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/shadercross"
)

/*
	BIND GROUP MODEL:

	Group 0:
		Binding 0 -> MaterialUniforms (uniform buffer)

	Group 1:
		Binding 0 -> Texture (future extension)
*/

const (
	DefaultVertexShaderPath   = "resources/shaders/basic_shader.vert.hlsl"
	DefaultFragmentShaderPath = "resources/shaders/basic_shader.frag.hlsl"
)

const (
	MaterialBindGroup = 0
	MaterialBinding   = 0
)

const vertex2DPitch = 52

type MaterialUniforms struct {
	UseTexture   uint32
	UseCircle    uint32
	CircleRadius float32
	CircleEdge   float32
}

// ================= SHADER =================

type Shader struct {
	Device *sdl.GPUDevice

	VertexShader   *sdl.GPUShader
	FragmentShader *sdl.GPUShader
	Pipeline       *sdl.GPUGraphicsPipeline

	VertexPath   string
	FragmentPath string

	ColorTargetFormat sdl.GPUTextureFormat

	mu sync.RWMutex
}

func NewShader(device *sdl.GPUDevice, vertexPath, fragmentPath string, format sdl.GPUTextureFormat) (*Shader, error) {
	s := &Shader{
		Device:            device,
		VertexPath:        vertexPath,
		FragmentPath:      fragmentPath,
		ColorTargetFormat: format,
	}
	return s, s.Reload()
}

func (s *Shader) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vertSrc, err := os.ReadFile(s.VertexPath)
	if err != nil {
		return err
	}
	fragSrc, err := os.ReadFile(s.FragmentPath)
	if err != nil {
		return err
	}

	// Compile HLSL to SPIR-V using shadercross
	vertSpv, err := shadercross.CompileSPIRVFromHLSL(&shadercross.HLSLInfo{
		Source:      string(vertSrc),
		Entrypoint:  "main",
		ShaderStage: shadercross.SHADERSTAGE_VERTEX,
	})
	if err != nil {
		return fmt.Errorf("vertex compile failed: %w", err)
	}

	fragSpv, err := shadercross.CompileSPIRVFromHLSL(&shadercross.HLSLInfo{
		Source:      string(fragSrc),
		Entrypoint:  "main",
		ShaderStage: shadercross.SHADERSTAGE_FRAGMENT,
	})
	if err != nil {
		return fmt.Errorf("fragment compile failed: %w", err)
	}

	// Create GPU shaders from SPIR-V
	vertShader, err := s.Device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:        vertSpv,
		Format:      sdl.GPU_SHADERFORMAT_SPIRV,
		Stage:       sdl.GPU_SHADERSTAGE_VERTEX,
		Entrypoint:  "main",
		NumSamplers: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to create vertex shader: %w", err)
	}

	fragShader, err := s.Device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:        fragSpv,
		Format:      sdl.GPU_SHADERFORMAT_SPIRV,
		Stage:       sdl.GPU_SHADERSTAGE_FRAGMENT,
		Entrypoint:  "main",
		NumSamplers: 1,
	})
	if err != nil {
		s.Device.ReleaseShader(vertShader)
		return fmt.Errorf("failed to create fragment shader: %w", err)
	}

	// Create graphics pipeline with push constants only (no descriptor sets needed)
	pipeline, err := s.Device.CreateGraphicsPipeline(&sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader:   vertShader,
		FragmentShader: fragShader,

		VertexInputState: sdl.GPUVertexInputState{
			VertexAttributes: []sdl.GPUVertexAttribute{
				{Location: 0, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Offset: 0},
				{Location: 1, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT4, Offset: 8},
				{Location: 2, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Offset: 24},
				{Location: 3, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Offset: 32},
				{Location: 4, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Offset: 40},
				{Location: 5, BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT, Offset: 48},
			},
			VertexBufferDescriptions: []sdl.GPUVertexBufferDescription{
				{Slot: 0, Pitch: vertex2DPitch, InputRate: sdl.GPU_VERTEXINPUTRATE_VERTEX},
			},
		},

		PrimitiveType: sdl.GPU_PRIMITIVETYPE_TRIANGLELIST,

		RasterizerState: sdl.GPURasterizerState{
			FillMode:  sdl.GPU_FILLMODE_FILL,
			CullMode:  sdl.GPU_CULLMODE_NONE,
			FrontFace: sdl.GPU_FRONTFACE_COUNTER_CLOCKWISE,
		},

		MultisampleState: sdl.GPUMultisampleState{
			SampleCount: sdl.GPU_SAMPLECOUNT_1,
		},

		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: []sdl.GPUColorTargetDescription{
				{
					Format: s.ColorTargetFormat,
					BlendState: sdl.GPUColorTargetBlendState{
						SrcColorBlendfactor: sdl.GPU_BLENDFACTOR_SRC_ALPHA,
						DstColorBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
						ColorBlendOp:        sdl.GPU_BLENDOP_ADD,
						SrcAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE,
						DstAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
						AlphaBlendOp:        sdl.GPU_BLENDOP_ADD,
						EnableBlend:         true,
					},
				},
			},
		},
	})

	if err != nil {
		s.Device.ReleaseShader(vertShader)
		s.Device.ReleaseShader(fragShader)
		return fmt.Errorf("pipeline creation failed: %w", err)
	}

	// Preserve old resources so we can release them after the new pipeline is bound.
	oldVert := s.VertexShader
	oldFrag := s.FragmentShader
	oldPipe := s.Pipeline

	// Assign new resources
	s.VertexShader = vertShader
	s.FragmentShader = fragShader
	s.Pipeline = pipeline

	// Release previous resources (if any) to avoid leaking GPU objects.
	if oldVert != nil {
		s.Device.ReleaseShader(oldVert)
	}
	if oldFrag != nil {
		s.Device.ReleaseShader(oldFrag)
	}
	if oldPipe != nil {
		s.Device.ReleaseGraphicsPipeline(oldPipe)
	}

	return nil
}

func (s *Shader) Bind(rp *sdl.GPURenderPass) {
	if s == nil || s.Pipeline == nil {
		return
	}
	rp.BindGraphicsPipeline(s.Pipeline)
}

func (s *Shader) Delete() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.VertexShader != nil {
		s.Device.ReleaseShader(s.VertexShader)
		s.VertexShader = nil
	}

	if s.FragmentShader != nil {
		s.Device.ReleaseShader(s.FragmentShader)
		s.FragmentShader = nil
	}

	if s.Pipeline != nil {
		s.Device.ReleaseGraphicsPipeline(s.Pipeline)
		s.Pipeline = nil
	}
}

type Material struct {
	Shader *Shader

	UseCircle    bool
	CircleRadius float32
	CircleEdge   float32

	mu sync.RWMutex
}

func NewMaterial(shader *Shader) *Material {
	return &Material{Shader: shader}
}

func (m *Material) BuildUniformBuffer() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u := MaterialUniforms{
		UseTexture:   0, // Textures disabled in this simplified version
		UseCircle:    boolToUint(m.UseCircle),
		CircleRadius: m.CircleRadius,
		CircleEdge:   m.CircleEdge,
	}

	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:], u.UseTexture)
	binary.LittleEndian.PutUint32(data[4:], u.UseCircle)
	binary.LittleEndian.PutUint32(data[8:], math.Float32bits(u.CircleRadius))
	binary.LittleEndian.PutUint32(data[12:], math.Float32bits(u.CircleEdge))

	return data
}

func (m *Material) CircleParams() (bool, float32, float32) {
	if m == nil {
		return false, 0, 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.UseCircle, m.CircleRadius, m.CircleEdge
}

// Apply applies the material (placeholder for future use)
func (m *Material) Apply(cmd *sdl.GPUCommandBuffer) {
	if m == nil || m.Shader == nil {
		return
	}
	// For now, shaders are simple and don't require uniform data
	// This will be expanded when descriptor sets/push constants are properly set up
}

func boolToUint(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

// ...existing code...

func (m *Material) CircleMask(radius, edge float32) *Material {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.UseCircle = true
	m.CircleRadius = radius
	m.CircleEdge = edge

	return m
}

func (m *Material) NoCircleMask() *Material {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.UseCircle = false
	return m
}

func (m *Material) Clone() *Material {
	if m == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return &Material{
		Shader:       m.Shader,
		UseCircle:    m.UseCircle,
		CircleRadius: m.CircleRadius,
		CircleEdge:   m.CircleEdge,
	}
}
