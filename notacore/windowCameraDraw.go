package notacore

import "NotaborEngine/notaentity"

// Draw queues entities for rendering.
// It uses the default camera if no camera is specified.
func (w *Window) Draw(alpha float32, cam *Camera2D, entities ...*notaentity.Entity) error {
	if cam == nil {
		cam = w.DefaultCamera
	}
	view := cam.ViewMatrix()
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if err := entity.DrawWithView(w.RunTime.Renderer, view, alpha); err != nil {
			return err
		}
	}
	return nil
}
