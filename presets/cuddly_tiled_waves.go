package presets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CuddlyMegaScrollerBackdrop is a static 8-pixel tile field with two
// independent row waves; source artwork and wave terms remain editable.
func CuddlyMegaScrollerBackdrop(tile *ebiten.Image) composite.TiledWaveBackdropConfig {
	return composite.TiledWaveBackdropConfig{
		Tile: tile, TileCanvas: image.Pt(500, 240),
		Tiles: composite.BackgroundConfig{PeriodX: 8, PeriodY: 8, Filter: ebiten.FilterLinear},
		Wave: composite.WaveStrips{Axis: composite.Rows, Thickness: 1, Filter: ebiten.FilterNearest,
			Waves: []composite.StripWave{{Amplitude: 30, Spatial: .03, Speed: -.05}, {Amplitude: 30, Spatial: .01, Speed: .08}}},
		OutputX: -110, OutputY: -9,
	}
}

// CuddlyLEDBackdrop retains uncovered source pixels and can be drawn once
// before the first visible frame to preserve the screen's preloaded wave phase.
func CuddlyLEDBackdrop(tile *ebiten.Image) composite.TiledWaveBackdropConfig {
	motionConfig := CuddlyLEDTileOffset()
	return composite.TiledWaveBackdropConfig{
		Tile: tile, TileCanvas: image.Pt(384, 233),
		Tiles: composite.BackgroundConfig{PeriodX: 32, PeriodY: 33, Filter: ebiten.FilterLinear},
		Wave: composite.WaveStrips{Axis: composite.Rows, Thickness: 1, Filter: ebiten.FilterLinear,
			Waves: []composite.StripWave{{Amplitude: 6, Spatial: .08, Speed: .2}}},
		OutputX: -32, RetainSource: true,
		SourceFilter: ebiten.FilterLinear, SourceBlend: ebiten.BlendSourceOver,
		MotionY: &motionConfig,
	}
}
