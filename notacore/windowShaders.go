package notacore

import (
	"NotaborEngine/notaobject"
	"errors"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func (w *Window) DeleteShader(name string) uint32 {
	shader, ok := w.Shaders[name]
	if !ok {
		return 0
	}
	gl.DeleteProgram(shader.Program)
	delete(w.Shaders, name)
	return shader.Program
}

func (w *Window) UpdateShader(name string) error {
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

func (w *Window) GetShader(name string) (*notaobject.Shader, error) {
	shader, ok := w.Shaders[name]
	if !ok {
		return nil, errors.New("shader with name " + name + " not found")
	}
	return shader, nil
}

func (w *Window) CreateShader(name, vertexPath, fragmentPath string) (*notaobject.Shader, error) {
	if w.Shaders == nil {
		w.Shaders = make(map[string]*notaobject.Shader)
	}

	if _, found := w.Shaders[name]; found {
		return nil, errors.New("shader with name " + name + " already exists")
	}

	w.MakeContextCurrent()
	shader, err := notaobject.NewShader(name, vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}
	w.Shaders[name] = shader
	return shader, nil
}

func (w *Window) UseShader(name string) error {
	shader, err := w.GetShader(name)
	if err != nil {
		return err
	}
	gl.UseProgram(shader.Program)
	return nil
}
