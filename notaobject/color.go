package notaobject

import (
	"fmt"
	"strconv"
	"strings"
)

type Color struct {
	R, G, B, A float32
}

var (
	White       = Color{1, 1, 1, 1}
	Black       = Color{0, 0, 0, 1}
	Red         = Color{1, 0, 0, 1}
	Green       = Color{0, 1, 0, 1}
	Blue        = Color{0, 0, 1, 1}
	Magenta     = Color{1, 0, 1, 1}
	Yellow      = Color{1, 1, 0, 1}
	Cyan        = Color{0, 1, 1, 1}
	Gray        = Color{0.5, 0.5, 0.5, 1}
	Silver      = Color{0.75, 0.75, 0.75, 1}
	Maroon      = Color{0.5, 0, 0, 1}
	Olive       = Color{0.5, 0.5, 0, 1}
	Navy        = Color{0, 0, 0.5, 1}
	Purple      = Color{0.5, 0, 0.5, 1}
	Teal        = Color{0, 0.5, 0.5, 1}
	Orange      = Color{1, 0.5, 0, 1}
	Transparent = Color{0, 0, 0, 0}
)

func RGBA(r, g, b, a float32) Color {
	return Color{r, g, b, a}
}
func RGB(r, g, b float32) Color {
	return RGBA(r, g, b, 1)
}
func FromBytes(r, g, b, a uint8) Color {
	return RGBA(float32(r)/255, float32(g)/255, float32(b)/255, float32(a)/255)
}

func FromHex(hex string) (Color, error) {
	s := strings.TrimPrefix(hex, "#")

	if len(s) != 6 && len(s) != 8 {
		return Color{}, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return Color{}, err
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return Color{}, err
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return Color{}, err
	}

	a := uint64(255)
	if len(s) == 8 {
		a, err = strconv.ParseUint(s[6:8], 16, 8)
		if err != nil {
			return Color{}, err
		}
	}

	color := Color{
		R: float32(r) / 255.0,
		G: float32(g) / 255.0,
		B: float32(b) / 255.0,
		A: float32(a) / 255.0,
	}

	return color.Clamp(), nil
}

func (c Color) WithAlpha(a float32) Color {
	return Color{c.R, c.G, c.B, a}
}

func (c Color) Clamp() Color {
	if c.R < 0 {
		c.R = 0
	}
	if c.G < 0 {
		c.G = 0
	}
	if c.B < 0 {
		c.B = 0
	}
	if c.A < 0 {
		c.A = 0
	}
	if c.R > 1 {
		c.R = 1
	}
	if c.G > 1 {
		c.G = 1
	}
	if c.B > 1 {
		c.B = 1
	}
	if c.A > 1 {
		c.A = 1
	}
	return c
}

func (c Color) ToVec4() [4]float32 {
	return [4]float32{c.R, c.G, c.B, c.A}
}

func (c Color) Lerp(to Color, t float32) Color {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	return Color{
		R: c.R + (to.R-c.R)*t,
		G: c.G + (to.G-c.G)*t,
		B: c.B + (to.B-c.B)*t,
		A: c.A + (to.A-c.A)*t,
	}
}
