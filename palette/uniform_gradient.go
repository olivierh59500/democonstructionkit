package palette

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// UniformGradientConfig interpolates an evenly spaced color bank. CenterSamples
// reads each pixel at coordinate + 0.5, while RoundNearest selects math.Round
// instead of byte truncation. Axis can be vertical or horizontal.
type UniformGradientConfig struct {
	Width, Height int
	Axis          GradientAxis
	Colors        []color.RGBA
	CenterSamples bool
	RoundNearest  bool
}

// NewUniformGradient builds one CPU material image for later GPU upload. It
// fills identical rows/columns directly, avoiding a per-pixel Set call.
func NewUniformGradient(c UniformGradientConfig) (*image.RGBA, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 ||
		c.Axis > GradientHorizontal || len(c.Colors) < 2 || len(c.Colors) > 1<<16 {
		return nil, fmt.Errorf("palette: invalid uniform gradient")
	}
	length := c.Height
	if c.Axis == GradientHorizontal {
		length = c.Width
	}
	span := float64(length - 1)
	if c.CenterSamples {
		span = float64(length)
	}
	if span == 0 {
		span = 1
	}
	segments := len(c.Colors) - 1
	sample := func(index int) color.RGBA {
		position := float64(index)
		if c.CenterSamples {
			position += .5
		}
		phase := position / span * float64(segments)
		bank := min(segments-1, int(phase))
		fraction := phase - float64(bank)
		a, b := c.Colors[bank], c.Colors[bank+1]
		channel := func(first, second uint8) uint8 {
			value := float64(first)*(1-fraction) + float64(second)*fraction
			if c.RoundNearest {
				value = math.Round(value)
			}
			return uint8(value)
		}
		return color.RGBA{R: channel(a.R, b.R), G: channel(a.G, b.G),
			B: channel(a.B, b.B), A: channel(a.A, b.A)}
	}
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	if c.Axis == GradientVertical {
		for y := 0; y < c.Height; y++ {
			clr := sample(y)
			row := img.Pix[y*img.Stride : y*img.Stride+c.Width*4]
			for x := 0; x < len(row); x += 4 {
				row[x], row[x+1], row[x+2], row[x+3] = clr.R, clr.G, clr.B, clr.A
			}
		}
	} else {
		for x := 0; x < c.Width; x++ {
			clr := sample(x)
			for y := 0; y < c.Height; y++ {
				offset := y*img.Stride + x*4
				img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2], img.Pix[offset+3] = clr.R, clr.G, clr.B, clr.A
			}
		}
	}
	return img, nil
}
