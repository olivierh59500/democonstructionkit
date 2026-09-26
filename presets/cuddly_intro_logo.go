package presets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyIntroMainLogoWarp keeps the row pass before the column pass. Both wave
// banks, working sizes, crop anchors and filters remain editable data.
func CuddlyIntroMainLogoWarp() []composite.WavePass {
	return []composite.WavePass{
		{Size: image.Pt(460, 120), X: 230, Y: 40, Wave: composite.WaveStrips{
			Axis: composite.Rows, Thickness: 1, CenterStrips: true, Filter: ebiten.FilterLinear,
			Waves: []composite.StripWave{{Amplitude: 7, Spatial: .03, Speed: -.035}, {Amplitude: 7, Spatial: .01, Speed: .05}},
		}},
		{Size: image.Pt(460, 80), X: 0, Y: 35, Wave: composite.WaveStrips{
			Axis: composite.Columns, Thickness: 1, CenterStrips: true, PixelSnap: true, Filter: ebiten.FilterNearest,
			Waves: []composite.StripWave{{Amplitude: 4, Spatial: .02, Speed: -.035}, {Amplitude: 4, Spatial: .005, Speed: .05}},
		}},
	}
}

// CuddlyIntroUnionLogoProfile binds a compiled row-offset table to the small
// logo; callers can replace the image without changing the deformation.
func CuddlyIntroUnionLogoProfile(offsets []float64) composite.ProfileStrips {
	return composite.ProfileStrips{Offsets: offsets, Speed: 2, Thickness: 1, Filter: ebiten.FilterNearest}
}

// CuddlyIntroSparkles uses borrowed sprite images and authored positions.
func CuddlyIntroSparkles(first, second *ebiten.Image, points []sprites.SparklePoint) sprites.SparkleConfig {
	return sprites.SparkleConfig{
		Images:    []sprites.SparkleImage{{Image: first, Spin: 10}, {Image: second, Angles: []float64{0, 45}}},
		Positions: points, StartScale: 1, EndScale: 0, ScaleStep: -.025,
		PauseTicks: 20, Filter: ebiten.FilterNearest,
	}
}
