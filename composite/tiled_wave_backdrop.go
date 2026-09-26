package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// TiledWaveBackdropConfig describes a cached tile field deformed by ordered
// row or column waves. Motion optionally scrolls the cached source before the
// wave; RetainSource preserves uncovered pixels between updates.
type TiledWaveBackdropConfig struct {
	Tile             *ebiten.Image
	TileCanvas       image.Point
	Tiles            BackgroundConfig
	Wave             WaveStrips
	OutputX, OutputY float64
	SourceX, SourceY float64
	MotionX, MotionY *motion.WrapBankConfig
	SourceFilter     ebiten.Filter
	SourceBlend      ebiten.Blend
	RetainSource     bool
}

// TiledWaveBackdrop owns one fixed tile surface and, when needed, one equally
// bounded scroll surface. It borrows its tile image and the draw destination.
// Draw samples phases; Step advances them once after the authored draw order.
type TiledWaveBackdrop struct {
	config TiledWaveBackdropConfig
	tiled  *ebiten.Image
	source *ebiten.Image
	x, y   *motion.WrapBank
	wave   WaveStrips
}

func NewTiledWaveBackdrop(config TiledWaveBackdropConfig) (*TiledWaveBackdrop, error) {
	if config.Tile == nil || config.TileCanvas.X < 1 || config.TileCanvas.Y < 1 ||
		config.TileCanvas.X > 8192 || config.TileCanvas.Y > 8192 || config.Wave.Axis > Columns ||
		len(config.Wave.Waves) > 64 {
		return nil, fmt.Errorf("composite: invalid tiled wave backdrop geometry")
	}
	for _, value := range [...]float64{config.OutputX, config.OutputY, config.SourceX, config.SourceY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite tiled wave placement")
		}
	}
	for _, wave := range config.Wave.Waves {
		for _, value := range [...]float64{wave.Phase, wave.Amplitude, wave.Spatial, wave.Speed} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("composite: nonfinite tiled wave parameter")
			}
		}
	}
	backdrop := &TiledWaveBackdrop{config: config, wave: config.Wave}
	backdrop.wave.Waves = append([]StripWave(nil), config.Wave.Waves...)
	for _, axis := range []struct {
		config *motion.WrapBankConfig
		clock  **motion.WrapBank
	}{{config.MotionX, &backdrop.x}, {config.MotionY, &backdrop.y}} {
		if axis.config == nil {
			continue
		}
		clock, err := motion.NewWrapBank(*axis.config)
		if err != nil {
			return nil, err
		}
		if clock.Len() != 1 {
			return nil, fmt.Errorf("composite: tiled wave motion needs one phase per axis")
		}
		*axis.clock = clock
	}
	background, err := NewBackground(config.Tiles)
	if err != nil {
		return nil, err
	}
	backdrop.tiled = render.NewSurface(config.TileCanvas.X, config.TileCanvas.Y)
	background.DrawAt(backdrop.tiled, config.Tile, 0, 0)
	if err := background.Err(); err != nil {
		backdrop.Close()
		return nil, err
	}
	if backdrop.x != nil || backdrop.y != nil || config.SourceX != 0 || config.SourceY != 0 || config.RetainSource {
		backdrop.source = render.NewSurface(config.TileCanvas.X, config.TileCanvas.Y)
	}
	return backdrop, nil
}

func (backdrop *TiledWaveBackdrop) Draw(dst *ebiten.Image) {
	if backdrop == nil || dst == nil || backdrop.tiled == nil {
		return
	}
	source := backdrop.tiled
	if backdrop.source != nil {
		if !backdrop.config.RetainSource {
			backdrop.source.Clear()
		}
		x, y := backdrop.config.SourceX, backdrop.config.SourceY
		if backdrop.x != nil {
			x += backdrop.x.At(0)
		}
		if backdrop.y != nil {
			y += backdrop.y.At(0)
		}
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = backdrop.config.SourceFilter, backdrop.config.SourceBlend
		op.GeoM.Translate(x, y)
		backdrop.source.DrawImage(backdrop.tiled, &op)
		source = backdrop.source
	}
	backdrop.wave.DrawAt(dst, source, backdrop.config.OutputX, backdrop.config.OutputY)
}

func (backdrop *TiledWaveBackdrop) Step() {
	if backdrop == nil {
		return
	}
	if backdrop.x != nil {
		backdrop.x.Step()
	}
	if backdrop.y != nil {
		backdrop.y.Step()
	}
	backdrop.wave.Advance()
}

// Wave and motion controllers allow live phase or speed changes for a logo,
// backdrop or scroller using the same compiled tile material.
func (backdrop *TiledWaveBackdrop) Wave() *WaveStrips         { return &backdrop.wave }
func (backdrop *TiledWaveBackdrop) MotionX() *motion.WrapBank { return backdrop.x }
func (backdrop *TiledWaveBackdrop) MotionY() *motion.WrapBank { return backdrop.y }

func (backdrop *TiledWaveBackdrop) Close() error {
	if backdrop == nil {
		return nil
	}
	if backdrop.source != nil {
		backdrop.source.Deallocate()
		backdrop.source = nil
	}
	if backdrop.tiled != nil {
		backdrop.tiled.Deallocate()
		backdrop.tiled = nil
	}
	return nil
}
