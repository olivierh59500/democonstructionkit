package presets

import "github.com/olivierh59500/democonstructionkit/scrolling"

// VivaPseudo3D describes four editable text banks. Face uses atlas metrics;
// sliced may instead provide a borrowed scrolling.Atlas to preserve custom
// per-rune source images and fallback behavior.
func VivaPseudo3D(face scrolling.Face, sliced *scrolling.Atlas, texts [4]string, width, height float64) scrolling.Pseudo3DConfig {
	bases := [...]float64{500, 250, 375, 125}
	banks := make([]scrolling.Pseudo3DBank, 4)
	for index, text := range texts {
		banks[index] = scrolling.Pseudo3DBank{Text: text, BaseY: bases[index], InvertScale: index < 2}
	}
	return scrolling.Pseudo3DConfig{
		Banks: banks, Face: face, Atlas: sliced,
		Advance: 64, PixelsPerUpdate: 4, Visible: 8,
		TicksPerSecond: 60, TimeOffset: 19, Width: width, Height: height,
		Pose: scrolling.Pseudo3DPose{
			ZLag: .15, ZRate: 5, ZIndexStep: .75,
			XTimeRate: 7, XIndexRate: 18,
			YLag: .1, YRate: 7, YIndexStep: .7,
			ZBase: 1.5, ZAmplitude: .5, ScaleSum: 3,
			HorizontalRate: .25, VerticalRate: .5,
			XOrigin: 40, XAmplitude: 32, YAmplitude: 42, YDepth: 32,
			OutputScaleX: 2, CullMargin: 100, Alpha: .9,
		},
	}
}
