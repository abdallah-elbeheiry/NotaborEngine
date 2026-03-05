package notacore

import (
	"NotaborEngine/notashader"
	"errors"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func (w *GlfwWindow2D) DeleteShader(name string) uint32 {
	shader, ok := w.Shaders[name]
	if !ok {
		return 0
	}
	gl.DeleteProgram(shader.Program)
	delete(w.Shaders, name)
	return shader.Program
}

func (w *GlfwWindow2D) UpdateShader(name string) error {
	shader, ok := w.Shaders[name]
	if !ok {
		return errors.New("shader with name " + name + " not found")
	}

	// Reload shader from files
	if err := shader.Reload(); err != nil {
		return err
	}

	return nil
}

func (w *GlfwWindow2D) GetShader(name string) (*notashader.Shader, error) {
	shader, ok := w.Shaders[name]
	if !ok {
		return nil, errors.New("shader with name " + name + " not found")
	}
	return shader, nil
}

func (w *GlfwWindow2D) CreateShader(name, vertexPath, fragmentPath string) (*notashader.Shader, error) {
	if w.Shaders == nil {
		w.Shaders = make(map[string]*notashader.Shader)
	}

	if _, found := w.Shaders[name]; found {
		return nil, errors.New("shader with name " + name + " already exists")
	}

	w.MakeContextCurrent()
	shader := notashader.NewShader(name, vertexPath, fragmentPath)
	w.Shaders[name] = shader
	return shader, nil
}

func (w *GlfwWindow2D) UseShader(name string) error {
	shader, err := w.GetShader(name)
	if err != nil {
		return err
	}
	gl.UseProgram(shader.Program)
	return nil
}

func (w *GlfwWindow3D) DeleteShader(name string) uint32 {
	shader, ok := w.Shaders[name]
	if !ok {
		return 0
	}
	gl.DeleteProgram(shader.Program)
	delete(w.Shaders, name)
	return shader.Program
}

func (w *GlfwWindow3D) UpdateShader(name string) error {
	shader, ok := w.Shaders[name]
	if !ok {
		return errors.New("shader with name " + name + " not found")
	}

	if err := shader.Reload(); err != nil {
		return err
	}

	return nil
}

func (w *GlfwWindow3D) GetShader(name string) (*notashader.Shader, error) {
	shader, ok := w.Shaders[name]
	if !ok {
		return nil, errors.New("shader with name " + name + " not found")
	}
	return shader, nil
}

func (w *GlfwWindow3D) CreateShader(name, vertexPath, fragmentPath string) error {
	if w.Shaders == nil {
		w.Shaders = make(map[string]*notashader.Shader)
	}

	if _, found := w.Shaders[name]; found {
		return errors.New("shader with name " + name + " already exists")
	}

	w.MakeContextCurrent()
	shader := notashader.NewShader(name, vertexPath, fragmentPath)
	w.Shaders[name] = shader
	return nil
}

func (w *GlfwWindow3D) UseShader(name string) error {
	shader, err := w.GetShader(name)
	if err != nil {
		return err
	}
	gl.UseProgram(shader.Program)
	return nil
}
