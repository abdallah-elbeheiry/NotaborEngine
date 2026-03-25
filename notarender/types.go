package notarender

import (
	"NotaborEngine/notacolor"
	"NotaborEngine/notamath"
	"NotaborEngine/notashader"
	"NotaborEngine/notatexture"
)

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
}

type PolygonRenderData struct {
	Vertices []Vertex2D
	Color    notacolor.Color
}
