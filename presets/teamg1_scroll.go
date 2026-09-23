package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// TeamG1ScrollWave retains three independently sampled profile sections.
func TeamG1ScrollWave() []float64 {
	result := make([]float64, 0, 577)
	stp1 := 7.0 / 180.0 * math.Pi
	stp2 := 3.0 / 180.0 * math.Pi
	for i := 0; i < 389; i++ {
		result = append(result, 20*math.Sin(float64(i)*stp1)+30*math.Cos(float64(i)*stp2))
	}
	stp1 = 72.0 / 180.0 * math.Pi
	for i := 0; i < 120; i++ {
		result = append(result, 4*math.Sin(float64(i)*stp1))
	}
	stp1 = 8.0 / 180.0 * math.Pi
	for i := 0; i < 68; i++ {
		result = append(result, 40*math.Sin(float64(i)*stp1))
	}
	return result
}

// TeamG1ProfiledScroll can use any atlas/message while preserving the authored
// surface dimensions, two-pixel rows, clipping and independent clock speeds.
func TeamG1ProfiledScroll(font *scrolling.Atlas, text string) scrolling.ProfiledConfig {
	return scrolling.ProfiledConfig{
		Font: font, Text: text, GlyphScale: 1.5, MissingAdvance: 32,
		SurfaceWidth: 640 + 512, SurfaceHeight: 54,
		OutputWidth: 640, OutputY: 400 - 100, StripHeight: 2,
		SourceXOrigin: 64 + 512/2, CullMargin: 200,
		TextSpeed: 2, ProfileSpeed: .5, Profile: TeamG1ScrollWave(),
	}
}
