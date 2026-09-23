package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/composite"
)

// TeamG1LogoOffsets returns independent samples for five authored sections.
// Each entry remains editable before construction and is safe to reuse with any
// image or screen size.
func TeamG1LogoOffsets() []float64 {
	result := make([]float64, 0, 600)
	for i := 0; i < 200; i++ {
		result = append(result, 50*math.Sin(float64(i)*.05))
	}
	for i := 0; i < 100; i++ {
		result = append(result, 30*math.Sin(float64(i)*.1)+20*math.Cos(float64(i)*.07))
	}
	for i := 0; i < 150; i++ {
		result = append(result, 40*math.Sin(float64(i)*.03))
	}
	for i := 0; i < 100; i++ {
		result = append(result, 20*math.Sin(float64(i)*.08))
	}
	for i := 0; i < 50; i++ {
		result = append(result, 10*math.Sin(float64(i)*.1))
	}
	return result
}

// TeamG1LogoProfile reproduces the original row sampling and left/right wrap.
func TeamG1LogoProfile(canvasWidth float64) composite.ProfileImageConfig {
	return composite.ProfileImageConfig{
		Offsets: TeamG1LogoOffsets(), PhaseStep: 2, RowStep: 2,
		Gain: .15, BaseX: canvasWidth / 2, BaseY: 60,
		MotionRate: .01, MotionSpan: canvasWidth, WrapWidth: canvasWidth,
	}
}
