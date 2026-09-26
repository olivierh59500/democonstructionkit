package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyDNAOuterLogoRowMotion keeps the two bouncing logo copies' independent
// phase offsets and sampled source-row height.
func CuddlyDNAOuterLogoRowMotion() motion.OuterSineRowsConfig {
	return motion.OuterSineRowsConfig{
		Bases: []float64{60, 268}, PhaseOffsets: []float64{0, 10},
		BaseY: 100, HeightBase: .5, HeightAmplitude: .75,
		HeightTimeDivisor: 15, BounceAmplitude: 10, BounceTimeDivisor: 15,
		HorizontalAmplitude: 4, HorizontalDivisor: 10, SourceHeight: 1,
	}
}

// CuddlyDNACenterLogoRowMotion keeps the center logo's reverse source-row
// sampling and per-row zoom independently editable.
func CuddlyDNACenterLogoRowMotion() motion.ZoomSineRowsConfig {
	return motion.ZoomSineRowsConfig{
		BaseX: 212, BaseY: 120, ReverseSourceStart: 55, SourceFactor: .5, SourceHeight: 2,
		ZoomBase: .75, ZoomAmplitude: .25, ZoomFrequency: 4,
		ZoomTimeDivisor: 67, ZoomIndexDivisor: 131, XShift: 48,
		YAmplitude: 20, YFrequency: 5, YTimeDivisor: 61, YIndexDivisor: 127,
	}
}

func CuddlyDNAOuterLogoRows(image *ebiten.Image) (composite.SampledRowsConfig, error) {
	program, err := motion.NewOuterSineRows(CuddlyDNAOuterLogoRowMotion())
	if err != nil {
		return composite.SampledRowsConfig{}, err
	}
	return composite.SampledRowsConfig{
		Image: image, Program: program, Rows: 56, Copies: 2,
		SourceWidth: 96, UseTime: true, Filter: ebiten.FilterNearest,
	}, nil
}

func CuddlyDNACenterLogoRows(image *ebiten.Image) (composite.SampledRowsConfig, error) {
	program, err := motion.NewZoomSineRows(CuddlyDNACenterLogoRowMotion())
	if err != nil {
		return composite.SampledRowsConfig{}, err
	}
	return composite.SampledRowsConfig{
		Image: image, Program: program, Rows: 56, Copies: 1,
		SourceWidth: 96, UseTime: true, Filter: ebiten.FilterNearest,
	}, nil
}
