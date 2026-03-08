package notassets

import (
	"NotaborEngine/notagl"
	"NotaborEngine/notashader"
)

type Material struct {
	Shader  *notashader.Shader
	Texture *notagl.Texture
}
