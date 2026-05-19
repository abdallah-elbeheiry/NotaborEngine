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
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/go-gl/gl/v4.6-core/gl"
)

type Texture struct {
	ID        uint32
	Width     int32
	Height    int32
	ImageData []byte // Store raw pixel data
	Loaded    bool   // Whether OpenGL texture is created

	Device     *sdl.GPUDevice
	GPUTexture *sdl.GPUTexture
	GPULoaded  bool
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

// CreateGPUTexture creates an SDL GPU texture from the stored image data.
func (t *Texture) CreateGPUTexture(device *sdl.GPUDevice) error {
	if t == nil {
		return fmt.Errorf("texture is nil")
	}
	if device == nil {
		return fmt.Errorf("GPU device is nil")
	}
	if t.GPULoaded && t.GPUTexture != nil {
		return nil
	}
	if t.ImageData == nil {
		return fmt.Errorf("no image data to upload")
	}

	gpuTex, err := device.CreateTexture(&sdl.GPUTextureCreateInfo{
		Type:              sdl.GPU_TEXTURETYPE_2D,
		Format:            sdl.GPU_TEXTUREFORMAT_R8G8B8A8_UNORM,
		Usage:             sdl.GPU_TEXTUREUSAGE_SAMPLER,
		Width:             uint32(t.Width),
		Height:            uint32(t.Height),
		LayerCountOrDepth: 1,
		NumLevels:         1,
		SampleCount:       sdl.GPU_SAMPLECOUNT_1,
	})
	if err != nil {
		return fmt.Errorf("failed to create GPU texture: %w", err)
	}

	uploadSize := uint32(len(t.ImageData))
	transferBuffer, err := device.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  uploadSize,
	})
	if err != nil {
		device.ReleaseTexture(gpuTex)
		return fmt.Errorf("failed to create transfer buffer: %w", err)
	}

	ptr, err := device.MapTransferBuffer(transferBuffer, false)
	if err != nil {
		device.ReleaseTransferBuffer(transferBuffer)
		device.ReleaseTexture(gpuTex)
		return fmt.Errorf("failed to map transfer buffer: %w", err)
	}

	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(t.ImageData)),
		t.ImageData,
	)
	device.UnmapTransferBuffer(transferBuffer)

	cmdBuf, err := device.AcquireCommandBuffer()
	if err != nil {
		device.ReleaseTransferBuffer(transferBuffer)
		device.ReleaseTexture(gpuTex)
		return fmt.Errorf("failed to acquire command buffer: %w", err)
	}

	copyPass := cmdBuf.BeginCopyPass()
	copyPass.UploadToGPUTexture(
		&sdl.GPUTextureTransferInfo{
			TransferBuffer: transferBuffer,
			Offset:         0,
		},
		&sdl.GPUTextureRegion{
			Texture: gpuTex,
			W:       uint32(t.Width),
			H:       uint32(t.Height),
			D:       1,
		},
		false,
	)
	copyPass.End()

	if err := cmdBuf.Submit(); err != nil {
		device.ReleaseTransferBuffer(transferBuffer)
		device.ReleaseTexture(gpuTex)
		return fmt.Errorf("failed to submit texture upload: %w", err)
	}

	device.ReleaseTransferBuffer(transferBuffer)

	t.Device = device
	t.GPUTexture = gpuTex
	t.GPULoaded = true
	return nil
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
	if t.GPUTexture != nil && t.Device != nil {
		t.Device.ReleaseTexture(t.GPUTexture)
		t.GPUTexture = nil
		t.GPULoaded = false
	}
	if t.Loaded {
		gl.DeleteTextures(1, &t.ID)
		t.Loaded = false
	}
}

// Bind binds the texture for SDL GPU fragment sampling.
func (t *Texture) Bind(renderPass *sdl.GPURenderPass, sampler *sdl.GPUSampler) {
	if t == nil || renderPass == nil || sampler == nil || t.GPUTexture == nil {
		return
	}

	renderPass.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{
		{
			Texture: t.GPUTexture,
			Sampler: sampler,
		},
	})
}
