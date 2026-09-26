package palette

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"
)

func TestVerticalGradientRetainsPhenomenaChannelTruncation(t *testing.T) {
	stops := []GradientStop{
		{Color: color.RGBA{R: 0x44, B: 0x44, A: 0xff}, Offset: 0},
		{Color: color.RGBA{R: 0xff, G: 0xdd, B: 0xff, A: 0xff}, Offset: .5},
		{Color: color.RGBA{R: 0x11, G: 0x11, B: 0x44, A: 0xff}, Offset: 1},
	}
	for _, height := range []int{9, 12, 33} {
		got, err := NewGradient(GradientConfig{Width: 4, Height: height, Stops: stops})
		if err != nil {
			t.Fatal(err)
		}
		want := image.NewRGBA(image.Rect(0, 0, 4, height))
		for y := 0; y < height; y++ {
			time := float64(y) / float64(height-1)
			var sample color.RGBA
			for i := 0; i < len(stops)-1; i++ {
				if time >= stops[i].Offset && time <= stops[i+1].Offset {
					fraction := (time - stops[i].Offset) / (stops[i+1].Offset - stops[i].Offset)
					a, b := stops[i].Color, stops[i+1].Color
					sample = color.RGBA{
						R: uint8(float64(a.R)*(1-fraction) + float64(b.R)*fraction),
						G: uint8(float64(a.G)*(1-fraction) + float64(b.G)*fraction),
						B: uint8(float64(a.B)*(1-fraction) + float64(b.B)*fraction),
						A: uint8(float64(a.A)*(1-fraction) + float64(b.A)*fraction),
					}
					break
				}
			}
			for x := 0; x < 4; x++ {
				want.SetRGBA(x, y, sample)
			}
		}
		if !bytes.Equal(got.Pix, want.Pix) {
			t.Fatalf("gradient height %d changed bytes", height)
		}
	}
}

func TestTintableSilhouetteAndInversionKeepSourceAlpha(t *testing.T) {
	source := image.NewNRGBA(image.Rect(5, 7, 7, 9))
	source.SetNRGBA(5, 7, color.NRGBA{R: 17, G: 64, B: 200, A: 0})
	source.SetNRGBA(6, 7, color.NRGBA{R: 20, G: 120, B: 230, A: 80})
	source.SetNRGBA(5, 8, color.NRGBA{R: 100, G: 200, B: 30, A: 255})
	source.SetNRGBA(6, 8, color.NRGBA{R: 255, G: 0, B: 1, A: 170})
	white, err := WhiteSilhouette(source)
	if err != nil {
		t.Fatal(err)
	}
	inverted, err := InvertRGB(source)
	if err != nil {
		t.Fatal(err)
	}
	if white.Bounds() != image.Rect(0, 0, 2, 2) || inverted.Bounds() != white.Bounds() {
		t.Fatal("material bounds retained the source offset")
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			original := source.NRGBAAt(x+5, y+7)
			if got := white.NRGBAAt(x, y); got != (color.NRGBA{R: 255, G: 255, B: 255, A: original.A}) {
				t.Fatalf("white alpha at %d,%d = %+v", x, y, got)
			}
			if got := inverted.NRGBAAt(x, y); got != (color.NRGBA{R: 255 - original.R, G: 255 - original.G, B: 255 - original.B, A: original.A}) {
				t.Fatalf("inverted pixel at %d,%d = %+v", x, y, got)
			}
		}
	}
}

func TestHSLColorCyclePrimaryColors(t *testing.T) {
	for _, sample := range []struct {
		hue float64
		rgb [3]float64
	}{{0, [3]float64{1, 0, 0}}, {1.0 / 3.0, [3]float64{0, 1, 0}}, {2.0 / 3.0, [3]float64{0, 0, 1}}} {
		r, g, b := HSLToRGB(sample.hue, 1, .5)
		for i, value := range [...]float64{r, g, b} {
			if math.Abs(value-sample.rgb[i]) > 1e-12 {
				t.Fatalf("hue %v channel %d = %v, want %v", sample.hue, i, value, sample.rgb[i])
			}
		}
	}
}
