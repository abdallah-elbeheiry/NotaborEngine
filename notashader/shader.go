package notashader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Shader struct {
	Name         string
	VertexPath   string
	FragmentPath string
	Program      uint32
	Uniforms     map[string]int32
}

const (
	UseTexture   = "uUseTexture"
	UseCircle    = "uCircleMask"
	CircleRadius = "uCircleRadius"
	CircleEdge   = "uCircleEdge"
	Texture      = "uTexture"
)

// Load shader source with #include support
func loadShaderSource(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	src := string(data)
	dir := filepath.Dir(path)
	return preprocessIncludes(src, dir)
}

// Recursively handle #include "file.glsl"
func preprocessIncludes(src, baseDir string) (string, error) {
	lines := strings.Split(src, "\n")
	var out []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#include") {
			file := strings.Trim(line[len("#include"):], ` "`)
			content, err := loadShaderSource(filepath.Join(baseDir, file))
			if err != nil {
				return "", err
			}
			out = append(out, content)
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n"), nil
}

// Create new shader from file paths
func NewShader(name, vertexPath, fragmentPath string) *Shader {
	sh := &Shader{
		Name:         name,
		VertexPath:   vertexPath,
		FragmentPath: fragmentPath,
		Uniforms:     make(map[string]int32),
	}
	err := sh.Reload()
	if err != nil {
		return nil
	}
	return sh
}

// Set uniform dynamically
func (s *Shader) SetUniform(name string, value interface{}) {
	loc, ok := s.Uniforms[name]
	if !ok {
		loc = gl.GetUniformLocation(s.Program, gl.Str(name+"\x00"))
		s.Uniforms[name] = loc
	}

	switch v := value.(type) {
	case float32:
		gl.Uniform1f(loc, v)
	case int32:
		gl.Uniform1i(loc, v)
	case bool:
		if v {
			gl.Uniform1i(loc, 1)
		} else {
			gl.Uniform1i(loc, 0)
		}
	case [4]float32:
		gl.Uniform4f(loc, v[0], v[1], v[2], v[3])
	case [16]float32: // optional: 4x4 matrix
		gl.UniformMatrix4fv(loc, 1, false, &v[0])
	}
}

// Compile & reload shader program
func (s *Shader) Reload() error {
	vertSrc, err := loadShaderSource(s.VertexPath)
	if err != nil {
		return err
	}
	fragSrc, err := loadShaderSource(s.FragmentPath)
	if err != nil {
		return err
	}

	newProg := CreateProgram(vertSrc, fragSrc)
	if s.Program != 0 {
		gl.DeleteProgram(s.Program) // free old GPU program
	}
	s.Program = newProg
	s.Uniforms = make(map[string]int32)
	return nil
}

// Compile shader source
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	sources, free := gl.Strs(source + "\x00")
	defer free()
	gl.ShaderSource(shader, 1, sources, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])
		return 0, fmt.Errorf("failed to compile shader: %s", log)
	}
	return shader, nil
}

// Link program
func CreateProgram(vertexSrc, fragmentSrc string) uint32 {
	vert, err := compileShader(vertexSrc, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	frag, err := compileShader(fragmentSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vert)
	gl.AttachShader(prog, frag)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(prog, logLength, nil, &log[0])
		panic(fmt.Sprintf("failed to link program: %s", log))
	}

	gl.DeleteShader(vert)
	gl.DeleteShader(frag)

	return prog
}
