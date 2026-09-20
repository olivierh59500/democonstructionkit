// Package fidelity compares rendered output with an original production reference.
package fidelity

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Result measures all RGBA channels, including transparent pixels.
type Result struct {
	Pixels, DifferentPixels int
	MaxChannelError         uint8
	MeanAbsoluteError, RMSE float64
}

// Compare returns an amplified difference image as well as numerical evidence.
func Compare(reference, candidate image.Image) (Result, *image.RGBA, error) {
	if reference.Bounds().Size() != candidate.Bounds().Size() {
		return Result{}, nil, fmt.Errorf("fidelity: image dimensions differ")
	}
	size := reference.Bounds().Size()
	diff := image.NewRGBA(image.Rectangle{Max: size})
	result := Result{Pixels: size.X * size.Y}
	sum, squares := 0.0, 0.0
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			a := color.NRGBAModel.Convert(reference.At(x+reference.Bounds().Min.X, y+reference.Bounds().Min.Y)).(color.NRGBA)
			b := color.NRGBAModel.Convert(candidate.At(x+candidate.Bounds().Min.X, y+candidate.Bounds().Min.Y)).(color.NRGBA)
			d := [4]uint8{}
			changed := false
			for i, pair := range [4][2]uint8{{a.R, b.R}, {a.G, b.G}, {a.B, b.B}, {a.A, b.A}} {
				v := int(pair[0]) - int(pair[1])
				if v < 0 {
					v = -v
				}
				d[i] = uint8(min(255, v*8))
				changed = changed || v != 0
				result.MaxChannelError = max(result.MaxChannelError, uint8(v))
				sum += float64(v)
				squares += float64(v * v)
			}
			if changed {
				result.DifferentPixels++
			}
			diff.SetRGBA(x, y, color.RGBA{R: max(d[0], d[3]), G: d[1], B: d[2], A: 255})
		}
	}
	if result.Pixels > 0 {
		result.MeanAbsoluteError = sum / float64(result.Pixels*4)
		result.RMSE = math.Sqrt(squares / float64(result.Pixels*4))
	}
	return result, diff, nil
}
