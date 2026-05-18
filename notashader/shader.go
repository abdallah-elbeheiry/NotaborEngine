package notashader

import (
	"fmt"
	"os"
	"sync"

	"github.com/Zyko0/go-sdl3/sdl"
)

type ShaderStage uint32

const (
	ShaderStageVertex   ShaderStage = ShaderStage(sdl.GPU_SHADERSTAGE_VERTEX)
	ShaderStageFragment ShaderStage = ShaderStage(sdl.GPU_SHADERSTAGE_FRAGMENT)

	// Uniform constants
	UseTexture   = "uUseTexture"
	TextureBind  = "uTextureBind"
	UseCircle    = "uUseCircle"
	CircleRadius = "uCircleRadius"
	CircleEdge   = "uCircleEdge"

	// Default shader paths
	DefaultVertexShaderPath   = "notashader/shaders/basic.vert"
	DefaultFragmentShaderPath = "notashader/shaders/basic.frag"
)

type Shader struct {
	Device         *sdl.GPUDevice
	VertexShader   *sdl.GPUShader
	FragmentShader *sdl.GPUShader
	Pipeline       *sdl.GPUGraphicsPipeline

	VertexPath   string
	FragmentPath string

	mu       sync.RWMutex
	Uniforms map[string]interface{}
}

// NewShader creates a shader from pre-compiled SPIR-V files
func NewShader(vertexPath, fragmentPath string) (*Shader, error) {
	sh := &Shader{
		VertexPath:   vertexPath,
		FragmentPath: fragmentPath,
		Uniforms:     make(map[string]interface{}),
	}

	if err := sh.Reload(); err != nil {
		return nil, err
	}
	return sh, nil
}

// Reload loads SPIR-V shaders and creates pipeline
func (s *Shader) Reload() error {
	vertData, err := os.ReadFile(s.VertexPath)
	if err != nil {
		return fmt.Errorf("failed to read vertex shader %s: %w", s.VertexPath, err)
	}

	fragData, err := os.ReadFile(s.FragmentPath)
	if err != nil {
		return fmt.Errorf("failed to read fragment shader %s: %w", s.FragmentPath, err)
	}

	// Create shaders
	vertShader, err := s.Device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:       vertData,
		Stage:      sdl.GPU_SHADERSTAGE_VERTEX,
		Format:     sdl.GPU_SHADERFORMAT_SPIRV,
		Entrypoint: "main",
	})
	if err != nil {
		s.Device.ReleaseShader(vertShader)
		return fmt.Errorf("failed to create vertex shader: %v", err)
	}

	fragShader, err := s.Device.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:       fragData,
		Stage:      sdl.GPU_SHADERSTAGE_FRAGMENT,
		Format:     sdl.GPU_SHADERFORMAT_SPIRV,
		Entrypoint: "main",
	})
	if err != nil {
		s.Device.ReleaseShader(vertShader)
		return fmt.Errorf("failed to create fragment shader: %v", err)
	}

	// Cleanup old resources
	s.Delete()

	s.VertexShader = vertShader
	s.FragmentShader = fragShader

	s.mu.Lock()
	s.Uniforms = make(map[string]interface{})
	s.mu.Unlock()

	return nil
}

// BindVulkan binds the pipeline to the current render pass
func (s *Shader) BindVulkan(renderPass *sdl.GPURenderPass) {
	if s.Pipeline != nil && renderPass != nil {
		renderPass.BindGraphicsPipeline(s.Pipeline)
	}
}

// SetUniform stores uniform value (for later push constants / descriptor sets)
func (s *Shader) SetUniform(name string, value interface{}) {
	s.mu.Lock()
	s.Uniforms[name] = value
	s.mu.Unlock()
}

// Bind is an alias for BindVulkan for compatibility
func (s *Shader) Bind() {
	// Placeholder for OpenGL-style binding if needed
}

// SetUniformVulkan sets uniforms for Vulkan
func (s *Shader) SetUniformVulkan(name string, value interface{}) {
	s.SetUniform(name, value)
}

func (s *Shader) Delete() {
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

	mu       sync.RWMutex
	uniforms map[string]interface{}
}

func NewMaterial(shader *Shader) *Material {
	return &Material{
		Shader:   shader,
		uniforms: make(map[string]interface{}),
	}
}

// Clone creates a shallow copy of the material that shares the shader but owns an independent uniform set.
func (m *Material) Clone() *Material {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &Material{
		Shader:   m.Shader,
		uniforms: make(map[string]interface{}, len(m.uniforms)),
	}
	for key, value := range m.uniforms {
		clone.uniforms[key] = value
	}
	return clone
}

func (m *Material) Set(name string, value interface{}) *Material {
	m.mu.Lock()
	m.uniforms[name] = value
	m.mu.Unlock()
	return m
}

// Bool Convenience methods
func (m *Material) Bool(name string, value bool) *Material     { return m.Set(name, value) }
func (m *Material) Int(name string, value int32) *Material     { return m.Set(name, value) }
func (m *Material) Float(name string, value float32) *Material { return m.Set(name, value) }

func (m *Material) CircleMask(radius, edge float32) *Material {
	return m.Bool(UseCircle, true).
		Float(CircleRadius, radius).
		Float(CircleEdge, edge)
}

// NoCircleMask disables the built-in circular mask on the material.
func (m *Material) NoCircleMask() *Material {
	return m.Bool(UseCircle, false)
}

// ApplyVulkan applies material uniforms to a shader for Vulkan rendering
func (m *Material) ApplyVulkan(textureEnabled bool, renderPass *sdl.GPURenderPass, shader *Shader) {
	if m == nil || m.Shader == nil {
		return
	}

	if shader != nil {
		shader.BindVulkan(renderPass)
		shader.SetUniform(UseTexture, textureEnabled)
		m.mu.RLock()
		defer m.mu.RUnlock()
		for name, value := range m.uniforms {
			shader.SetUniform(name, value)
		}
	}
}

// Apply applies material uniforms in a backend-agnostic way
func (m *Material) Apply(textureEnabled bool) {
	if m == nil || m.Shader == nil {
		return
	}

	m.Shader.Bind()
	m.Shader.SetUniform(UseTexture, textureEnabled)
	m.Shader.SetUniform(TextureBind, int32(0))
	m.Shader.SetUniform(UseCircle, false)

	m.mu.RLock()
	defer m.mu.RUnlock()
	for name, value := range m.uniforms {
		m.Shader.SetUniform(name, value)
	}
}
