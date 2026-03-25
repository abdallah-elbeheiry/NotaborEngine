package notatexture

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type Texture struct {
	ID        uint32
	Width     int32
	Height    int32
	ImageData []byte // Store raw pixel data
	Loaded    bool   // Whether OpenGL texture is created
}

// LoadImageData loads image data from file
// notagl/texture.go - Update LoadImageData
func LoadImageData(path string) (*Texture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture file: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("unsupported stride")
	}

	// Flip vertically while drawing
	bounds := rgba.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		srcY := bounds.Dy() - 1 - y // Flip Y coordinate
		draw.Draw(rgba,
			image.Rect(bounds.Min.X, y, bounds.Max.X, y+1),
			img,
			image.Point{X: bounds.Min.X, Y: srcY},
			draw.Src)
	}

	width := int32(rgba.Rect.Size().X)
	height := int32(rgba.Rect.Size().Y)

	return &Texture{
		ID:        0,
		Width:     width,
		Height:    height,
		ImageData: rgba.Pix,
		Loaded:    false,
	}, nil
}

// CreateGLTexture creates the actual OpenGL texture from stored image data
func (t *Texture) CreateGLTexture() error {
	if t.Loaded {
		return nil // Already created
	}

	if t.ImageData == nil {
		return fmt.Errorf("no image data to upload")
	}

	gl.GenTextures(1, &t.ID)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		t.Width,
		t.Height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(t.ImageData),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	t.Loaded = true
	return nil
}

func (t *Texture) Delete() {
	if t.Loaded {
		gl.DeleteTextures(1, &t.ID)
		t.Loaded = false
	}
}
