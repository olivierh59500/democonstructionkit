// Package palette builds reusable RGBA material images and color cycles without
// depending on a graphics device. Upload the resulting image once for Ebiten.
package palette

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type GradientAxis uint8

const (
	GradientVertical GradientAxis = iota
	GradientHorizontal
)

type GradientStop struct {
	Color  color.RGBA
	Offset float64
}

type GradientConfig struct {
	Width, Height int
	Axis          GradientAxis
	Stops         []GradientStop
}

// NewGradient samples adjacent stops with per-channel float interpolation and
// byte truncation. A coordinate outside the supplied stop intervals stays
// transparent, preserving authored palette gaps.
func NewGradient(c GradientConfig) (*image.RGBA, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 ||
		c.Axis > GradientHorizontal || len(c.Stops) < 2 {
		return nil, fmt.Errorf("palette: invalid gradient dimensions or stops")
	}
	for i, stop := range c.Stops {
		if math.IsNaN(stop.Offset) || math.IsInf(stop.Offset, 0) ||
			i > 0 && stop.Offset <= c.Stops[i-1].Offset {
			return nil, fmt.Errorf("palette: invalid gradient stop order")
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	sample := func(coordinate, length int) color.RGBA {
		t := 0.0
		if length > 1 {
			t = float64(coordinate) / float64(length-1)
		}
		for i := 0; i < len(c.Stops)-1; i++ {
			if t >= c.Stops[i].Offset && t <= c.Stops[i+1].Offset {
				local := (t - c.Stops[i].Offset) / (c.Stops[i+1].Offset - c.Stops[i].Offset)
				return MixRGBA(c.Stops[i].Color, c.Stops[i+1].Color, local)
			}
		}
		return color.RGBA{}
	}
	if c.Axis == GradientVertical {
		for y := 0; y < c.Height; y++ {
			clr := sample(y, c.Height)
			row := img.Pix[y*img.Stride : y*img.Stride+c.Width*4]
			for x := 0; x < len(row); x += 4 {
				row[x], row[x+1], row[x+2], row[x+3] = clr.R, clr.G, clr.B, clr.A
			}
		}
	} else {
		for x := 0; x < c.Width; x++ {
			clr := sample(x, c.Width)
			for y := 0; y < c.Height; y++ {
				offset := y*img.Stride + x*4
				img.Pix[offset], img.Pix[offset+1], img.Pix[offset+2], img.Pix[offset+3] = clr.R, clr.G, clr.B, clr.A
			}
		}
	}
	return img, nil
}

// MixRGBA retains the channel arithmetic of byte-truncated gradients.
func MixRGBA(first, second color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(first.R)*(1-t) + float64(second.R)*t),
		G: uint8(float64(first.G)*(1-t) + float64(second.G)*t),
		B: uint8(float64(first.B)*(1-t) + float64(second.B)*t),
		A: uint8(float64(first.A)*(1-t) + float64(second.A)*t),
	}
}

// WhiteSilhouette retains every source alpha while replacing RGB with white.
// Bounds are normalized to an origin at (0,0) for GPU upload.
func WhiteSilhouette(source image.Image) (*image.NRGBA, error) {
	if source == nil {
		return nil, fmt.Errorf("palette: nil silhouette source")
	}
	bounds := source.Bounds()
	mask := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := mask.Pix[(y-bounds.Min.Y)*mask.Stride:]
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := source.At(x, y).RGBA()
			offset := (x - bounds.Min.X) * 4
			row[offset], row[offset+1], row[offset+2], row[offset+3] = 0xff, 0xff, 0xff, uint8(alpha>>8)
		}
	}
	return mask, nil
}

// InvertRGB inverts unpremultiplied RGB while preserving alpha and source crop.
func InvertRGB(source image.Image) (*image.NRGBA, error) {
	if source == nil {
		return nil, fmt.Errorf("palette: nil inversion source")
	}
	bounds := source.Bounds()
	inverted := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := inverted.Pix[(y-bounds.Min.Y)*inverted.Stride:]
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(source.At(x, y)).(color.NRGBA)
			offset := (x - bounds.Min.X) * 4
			row[offset], row[offset+1], row[offset+2], row[offset+3] =
				0xff-pixel.R, 0xff-pixel.G, 0xff-pixel.B, pixel.A
		}
	}
	return inverted, nil
}

// HSLToRGB returns channel multipliers for a live hue-cycling material.
func HSLToRGB(h, s, l float64) (float64, float64, float64) {
	if s == 0 {
		return l, l, l
	}
	hueToRGB := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6.0 {
			return p + (q-p)*6*t
		}
		if t < 1.0/2.0 {
			return q
		}
		if t < 2.0/3.0 {
			return p + (q-p)*(2.0/3.0-t)*6
		}
		return p
	}
	q := 0.0
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	return hueToRGB(p, q, h+1.0/3.0), hueToRGB(p, q, h), hueToRGB(p, q, h-1.0/3.0)
}
