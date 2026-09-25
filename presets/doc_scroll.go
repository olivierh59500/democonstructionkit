package presets

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// DOCRowWave keeps the alternating four-section row-position program shared
// by the authored two-pass text image. Samples are independent for each caller.
func DOCRowWave() []float64 {
	first := TeamG1ScrollWave()[:389]
	result := append([]float64(nil), first...)
	step := 8.0 / 180.0 * math.Pi
	for i := 0; i < 68; i++ {
		result = append(result, 30*math.Sin(float64(i)*step))
	}
	result = append(result, first...)
	for i := 0; i < 189; i++ {
		result = append(result, 30*math.Sin(float64(i)*step))
	}
	return result
}

// CuddlyDOCRowWave retains the shorter three-section sampled-source program
// used by the paired inner and outer fonts of the Cuddly screen.
func CuddlyDOCRowWave() []float64 {
	wave := make([]float64, 389)
	for i := range wave {
		wave[i] = 20*math.Sin(float64(i)*(7.0/180*math.Pi)) + 30*math.Cos(float64(i)*(3.0/180*math.Pi))
	}
	for i := 0; i < 68; i++ {
		wave = append(wave, 30*math.Sin(float64(i)*(8.0/180*math.Pi)))
	}
	for i := 0; i < 189; i++ {
		wave = append(wave, 30*math.Sin(float64(i)*(8.0/180*math.Pi)))
	}
	return wave
}

// CuddlyDOCRowWarps returns two independently advancing strip programs over
// the same editable source-X lookup. The inner font also moves vertically.
func CuddlyDOCRowWarps() (outer, inner composite.RowWarpConfig) {
	outer = composite.RowWarpConfig{
		Mode: composite.RowWarpSourceX, Thickness: 2, SourceX: 64, SourceWidth: 640,
		Wave: CuddlyDOCRowWave(), WaveStep: 1, Filter: ebiten.FilterLinear,
	}
	inner = outer
	inner.VerticalBase, inner.VerticalAmplitude = 30, 30
	inner.VerticalDivisor, inner.VerticalStep = 20, 1.2
	return outer, inner
}

// DOCIntroRowBands keeps the scene cue in the message while DCK owns the
// circular fixed-width text transport and borrowed atlas glyph construction.
func DOCIntroRowBands(font *scrolling.Atlas, text string) scrolling.RowBandsConfig {
	return scrolling.RowBandsConfig{
		Font: font, Text: text, Advance: 62,
		FallbackRect: image.Rect(0, 0, 62, 50),
		WorkWidth:    768, Height: 50, TextSpeed: 5, Y: 62,
	}
}

// DOCMainRowBands renders two ordered horizontal displacement passes and a
// vertically moving second pass before the original visible crop.
func DOCMainRowBands(font *scrolling.Atlas, text string) scrolling.RowBandsConfig {
	wave := DOCRowWave()
	return scrolling.RowBandsConfig{
		Font: font, Text: text, Advance: 62,
		FallbackRect: image.Rect(0, 0, 62, 50),
		WorkWidth:    1024, Height: 50, TextSpeed: 3,
		Passes: []scrolling.RowBandPass{
			{Width: 1024, Height: 50, Thickness: 2, Wave: wave, WaveStep: 1},
			{Width: 1024, Height: 120, Thickness: 2, Wave: wave, WaveStep: 1,
				VerticalBase: 30, VerticalAmplitude: 30, VerticalDivisor: 20, VerticalStep: 1.2},
		},
		Crop: image.Rect(128, 0, 896, 120), Y: 62,
	}
}
