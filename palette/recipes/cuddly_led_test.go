package recipes

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/palette"
)

func TestCuddlyLEDGradientMatchesEverySourcePixel(t *testing.T) {
	got, err := palette.NewUniformGradient(CuddlyLEDGradient())
	if err != nil {
		t.Fatal(err)
	}
	stops := []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255},
		{0, 0, 255, 255}, {0, 255, 0, 255}, {255, 0, 0, 255},
		{0, 255, 0, 255}, {0, 0, 255, 255}, {255, 0, 0, 255},
		{0, 255, 0, 255}, {0, 0, 255, 255}, {0, 255, 0, 255}}
	want := image.NewRGBA(image.Rect(0, 0, 384, 2000))
	for y := 0; y < 2000; y++ {
		phase := (float64(y) + .5) / 2000 * 10
		index := min(9, int(phase))
		fraction := phase - float64(index)
		a, b := stops[index], stops[index+1]
		clr := color.RGBA{
			R: uint8(math.Round(float64(a.R)*(1-fraction) + float64(b.R)*fraction)),
			G: uint8(math.Round(float64(a.G)*(1-fraction) + float64(b.G)*fraction)),
			B: uint8(math.Round(float64(a.B)*(1-fraction) + float64(b.B)*fraction)), A: 255,
		}
		for x := 0; x < 384; x++ {
			want.SetRGBA(x, y, clr)
		}
	}
	if !bytes.Equal(got.Pix, want.Pix) {
		for i := range got.Pix {
			if got.Pix[i] != want.Pix[i] {
				t.Fatalf("pixel byte %d = %d, want %d", i, got.Pix[i], want.Pix[i])
			}
		}
	}
}

func BenchmarkCuddlyLEDGradient(b *testing.B) {
	config := CuddlyLEDGradient()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := palette.NewUniformGradient(config); err != nil {
			b.Fatal(err)
		}
	}
}
