package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// ScanlineBackgroundConfig describes a horizontally tiled picture sampled one
// row at a time. Program provides the cumulative row displacement, WaveStep
// advances its source index per logical tick, and bounce moves source rows with
// a separate sine clock. All distances are source pixels. Sample may override
// the computed source coordinate when a production needs a different row map.
type ScanlineBackgroundConfig struct {
	Tile                  *ebiten.Image
	Program               *DisplacementProgram
	Width, Height         int
	SurfaceWidth          int
	SurfaceHeight         int
	BaseX, BaseY          int
	WaveDivisor, WaveStep int
	BounceAmplitude       float64
	BounceRate            float64 // Radians per logical tick.
	AlternateDiagonal     bool
	Filter                ebiten.Filter
	Blend                 ebiten.Blend
	Sample                func(row, displacement, bounce int) (x, y int)
}

// ScanlineBackground compiles a bounded tile surface and reuses one quad batch.
// It can be drawn below a logo or scroller without coupling their clocks.
type ScanlineBackground struct {
	config ScanlineBackgroundConfig
	tiled  *ebiten.Image
	batch  *QuadBatch
	rows   []int
	bounce int
	closed bool
}

func NewScanlineBackground(c ScanlineBackgroundConfig) (*ScanlineBackground, error) {
	if c.Tile == nil || c.Program == nil || c.Width <= 0 || c.Height <= 0 || c.Width > 4096 || c.Height > 4096 ||
		c.WaveStep < -(1<<20) || c.WaveStep > 1<<20 || math.IsNaN(c.BounceAmplitude) || math.IsInf(c.BounceAmplitude, 0) ||
		math.IsNaN(c.BounceRate) || math.IsInf(c.BounceRate, 0) || c.BounceAmplitude < 0 || c.BounceAmplitude > 1<<30 {
		return nil, fmt.Errorf("composite: invalid scanline background dimensions or motion")
	}
	if c.WaveDivisor == 0 {
		c.WaveDivisor = 1
	}
	if c.WaveDivisor < 0 || c.WaveDivisor > 1<<20 {
		return nil, fmt.Errorf("composite: invalid scanline background wave divisor")
	}
	if c.SurfaceWidth == 0 {
		c.SurfaceWidth = c.Width + c.Tile.Bounds().Dx()
	}
	if c.SurfaceHeight == 0 {
		c.SurfaceHeight = c.Tile.Bounds().Dy()
	}
	if c.Tile.Bounds().Dx() < 1 || c.Tile.Bounds().Dy() < 1 || c.SurfaceWidth < c.Width+c.Tile.Bounds().Dx()-1 ||
		c.SurfaceHeight < 1 || c.SurfaceHeight > c.Tile.Bounds().Dy() || c.SurfaceWidth > 8192 || c.SurfaceHeight > 8192 {
		return nil, fmt.Errorf("composite: invalid scanline background source surface")
	}
	tiled := ebiten.NewImage(c.SurfaceWidth, c.SurfaceHeight)
	for x := 0; x < c.SurfaceWidth; x += c.Tile.Bounds().Dx() {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(x), 0)
		tiled.DrawImage(c.Tile, &op)
	}
	batch := NewQuadBatch(c.Height)
	batch.AlternateDiagonal = c.AlternateDiagonal
	batch.Options.Filter, batch.Options.Blend = c.Filter, c.Blend
	background := &ScanlineBackground{config: c, tiled: tiled, batch: batch, rows: make([]int, c.Height)}
	if err := background.Update(kit.Frame{}); err != nil {
		tiled.Deallocate()
		return nil, err
	}
	return background, nil
}

func (b *ScanlineBackground) Update(frame kit.Frame) error {
	if b == nil || b.closed {
		return fmt.Errorf("composite: scanline background is closed")
	}
	step := b.config.WaveStep
	if step < 0 {
		step = -step
	}
	if frame.Tick > uint64(int(^uint(0)>>1))/uint64(max(1, step)) {
		return fmt.Errorf("composite: scanline background wave clock overflow")
	}
	b.config.Program.Fill(b.rows, int(frame.Tick)*b.config.WaveStep)
	b.bounce = int(b.config.BounceAmplitude * math.Abs(math.Sin(float64(frame.Tick)*b.config.BounceRate)))
	return nil
}

func (b *ScanlineBackground) Draw(dst *ebiten.Image) {
	if b == nil || b.closed || dst == nil {
		return
	}
	c := b.config
	b.batch.Begin(dst, b.tiled)
	for row, wave := range b.rows {
		x := wrapBackground(c.BaseX+wave/c.WaveDivisor, c.Tile.Bounds().Dx())
		y := wrapBackground(row+c.BaseY+b.bounce, c.SurfaceHeight)
		if c.Sample != nil {
			x, y = c.Sample(row, wave, b.bounce)
		}
		b.batch.Rect(image.Rect(x, y, x+c.Width, y+1), 0, float32(row), float32(c.Width), 1)
	}
	b.batch.Flush()
}

func (b *ScanlineBackground) Close() error {
	if b == nil || b.closed {
		return nil
	}
	b.closed = true
	b.tiled.Deallocate()
	b.tiled = nil
	return nil
}

func wrapBackground(value, period int) int {
	value %= period
	if value < 0 {
		value += period
	}
	return value
}
