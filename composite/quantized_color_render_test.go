//go:build dck_gpu_rendercheck

package composite

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func TestQuantizedColorGPUIntegerGrids(t *testing.T) {
	testutil.RequireGPU(t)
	pixels := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for i := 0; i < 256; i++ {
		pixels.SetNRGBA(i%16, i/16, color.NRGBA{R: uint8(i), G: uint8(255 - i), B: uint8(i * 13), A: 255})
	}
	src, dst := ebiten.NewImageFromImage(pixels), ebiten.NewImage(16, 16)
	defer src.Deallocate()
	defer dst.Deallocate()
	for _, levels := range [][3]uint16{{3, 5, 15}, {31, 63, 31}, {255, 255, 255}} {
		q, err := NewQuantizedColor(QuantizedColorConfig{Levels: levels})
		if err != nil {
			t.Fatal(err)
		}
		for _, state := range []QuantizedColorState{
			{}, {Mode: QuantizedKeep},
			{Mode: QuantizedReplace, Target: [3]uint16{1000, 1, 0}},
			{Mode: QuantizedScale, Numerator: 1, Denominator: 2},
			{Mode: QuantizedScale, Numerator: 65535, Denominator: 65536},
			{Mode: QuantizedFromTarget, Target: levels, Numerator: 199, Denominator: 257},
			{Mode: QuantizedFromTarget, Target: [3]uint16{0, 1, 2}, Numerator: 511, Denominator: 1024},
		} {
			dst.Clear()
			if err := q.Draw(dst, src, state); err != nil {
				t.Fatal(err)
			}
			got := make([]byte, 16*16*4)
			dst.ReadPixels(got)
			for i := 0; i < 256; i++ {
				for channel := 0; channel < 3; channel++ {
					value := int(pixels.Pix[i*4+channel])
					want := value
					if state.Mode != QuantizedPassthrough {
						maxValue := int(levels[channel])
						value = (value*maxValue + 127) / 255
						switch state.Mode {
						case QuantizedReplace:
							value = int(state.Target[channel])
						case QuantizedScale:
							value = value * int(state.Numerator) / int(state.Denominator)
						case QuantizedFromTarget:
							target := int(state.Target[channel])
							value = target - int(math.Floor(float64((target-value)*int(state.Numerator))/float64(state.Denominator)))
						}
						value = min(maxValue, max(0, value))
						want = int(math.Round(float64(value) * 255 / float64(maxValue)))
					}
					tolerance := 0
					if state.Mode != QuantizedPassthrough && 255%int(levels[channel]) != 0 {
						tolerance = 1 // UNORM8 rounding for grids such as RGB565.
					}
					if delta := int(got[i*4+channel]) - want; delta < -tolerance || delta > tolerance {
						t.Fatalf("grid %v state %+v color %d channel %d: %d != %d", levels, state, i, channel, got[i*4+channel], want)
					}
				}
				if got[i*4+3] != 255 {
					t.Fatal("opaque grid operation changed alpha")
				}
			}
		}
		q.Close()
	}
}

func TestQuantizedColorGPUCropCutoffAndAlpha(t *testing.T) {
	testutil.RequireGPU(t)
	art := image.NewNRGBA(image.Rect(0, 0, 10, 4))
	for i, alpha := range []uint8{0, 64, 128} {
		art.SetNRGBA(5+i, 2, color.NRGBA{R: 153, G: 51, B: 204, A: alpha})
	}
	atlas := ebiten.NewImageFromImage(art)
	defer atlas.Deallocate()
	src := atlas.SubImage(image.Rect(5, 2, 8, 3)).(*ebiten.Image)
	dst := ebiten.NewImage(3, 1)
	defer dst.Deallocate()
	q, err := NewQuantizedColor(QuantizedColorConfig{AlphaThreshold: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	for _, state := range []QuantizedColorState{{Mode: QuantizedKeep}, {Mode: QuantizedReplace, Target: [3]uint16{15, 0, 0}}} {
		dst.Clear()
		if err := q.Draw(dst, src, state); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, 12)
		dst.ReadPixels(got)
		for _, value := range got[:8] {
			if value != 0 {
				t.Fatal("alpha cutoff did not discard the first two pixels", got)
			}
		}
		want := []int{77, 26, 102, 128}
		if state.Mode == QuantizedReplace {
			want = []int{128, 0, 0, 128}
		}
		for i, value := range want {
			if delta := int(got[8+i]) - value; delta < -1 || delta > 1 {
				t.Fatalf("cropped translucent state %+v: %v != %v", state, got[8:], want)
			}
		}
	}
	dst.Clear()
	if err := q.Draw(dst, src, QuantizedColorState{}); err != nil {
		t.Fatal(err)
	}
	got, original := make([]byte, 12), make([]byte, 12)
	dst.ReadPixels(got)
	src.ReadPixels(original)
	for i := range got {
		if got[i] != original[i] {
			t.Fatal("passthrough changed cropped source bytes", got, original)
		}
	}
}
