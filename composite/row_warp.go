package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// RowWarpMode chooses whether a wave moves each source crop or its destination.
// Source sampling preserves fractional atlas coordinates; destination movement
// preserves integer source rectangles and their exact bitmap edges.
type RowWarpMode uint8

const (
	RowWarpDestinationX RowWarpMode = iota
	RowWarpSourceX
)

// RowWarpConfig describes a reusable horizontal strip program. A zero-width
// SourceWidth selects the whole source image in destination-motion mode.
// Wave samples are indexed in strip order and WaveStep advances once per Step.
type RowWarpConfig struct {
	Mode                            RowWarpMode
	Thickness                       int
	SourceX, SourceWidth            float64
	Wave                            []float64
	WaveStep                        int
	VerticalBase, VerticalAmplitude float64
	VerticalDivisor, VerticalStep   float64
	Filter                          ebiten.Filter
}

// RowWarp owns strip and vertical phases, but borrows source and destination
// images. One instance can be drawn into several surfaces without advancing.
type RowWarp struct {
	config        RowWarpConfig
	wavePhase     int
	verticalPhase float64
}

func NewRowWarp(config RowWarpConfig) (*RowWarp, error) {
	finite := func(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
	if config.Mode != RowWarpDestinationX && config.Mode != RowWarpSourceX || config.Thickness < 1 || config.Thickness > 8192 || len(config.Wave) > 1<<20 ||
		!finite(config.SourceX) || !finite(config.SourceWidth) || !finite(config.VerticalBase) || !finite(config.VerticalAmplitude) || !finite(config.VerticalDivisor) || !finite(config.VerticalStep) ||
		config.VerticalDivisor < 0 || config.Mode == RowWarpSourceX && config.SourceWidth <= 0 {
		return nil, fmt.Errorf("composite: invalid row warp settings")
	}
	for _, offset := range config.Wave {
		if !finite(offset) {
			return nil, fmt.Errorf("composite: nonfinite row warp offset")
		}
	}
	config.Wave = append([]float64(nil), config.Wave...)
	return &RowWarp{config: config}, nil
}

// Vertical returns the current shared vertical displacement. Drawing remains
// read-only, so another layer may use the same value for its own placement.
func (warp *RowWarp) Vertical() float64 {
	vertical := warp.config.VerticalBase
	if warp.config.VerticalDivisor > 0 {
		vertical += warp.config.VerticalAmplitude * math.Cos(warp.verticalPhase/warp.config.VerticalDivisor)
	}
	return vertical
}

// Horizontal returns the current wave value for a strip index.
func (warp *RowWarp) Horizontal(row int) float64 {
	if len(warp.config.Wave) == 0 {
		return 0
	}
	index := (warp.wavePhase + row) % len(warp.config.Wave)
	if index < 0 {
		index += len(warp.config.Wave)
	}
	return warp.config.Wave[index]
}

// Step advances the two independent clocks without allocating or drawing.
func (warp *RowWarp) Step() error {
	warp.wavePhase += warp.config.WaveStep
	warp.verticalPhase += warp.config.VerticalStep
	if math.IsNaN(warp.verticalPhase) || math.IsInf(warp.verticalPhase, 0) {
		return fmt.Errorf("composite: row warp vertical clock overflows")
	}
	return nil
}

// DrawInto copies each source strip into an existing destination. Callers
// clear or retain that destination explicitly to control feedback and layering.
func (warp *RowWarp) DrawInto(dst, src *ebiten.Image) {
	if warp == nil || dst == nil || src == nil {
		return
	}
	bounds := src.Bounds()
	vertical := warp.Vertical()
	for top, row := 0, 0; top < bounds.Dy(); top, row = top+warp.config.Thickness, row+1 {
		height := min(warp.config.Thickness, bounds.Dy()-top)
		wave := warp.Horizontal(row)
		op := ebiten.DrawImageOptions{Filter: warp.config.Filter}
		op.GeoM.Translate(0, float64(top)+vertical)
		if warp.config.Mode == RowWarpSourceX {
			region := Region{X: warp.config.SourceX + wave, Y: float64(bounds.Min.Y + top), Width: warp.config.SourceWidth, Height: float64(height)}
			DrawRegion(dst, src, region, &op)
			continue
		}
		rect := image.Rect(bounds.Min.X, bounds.Min.Y+top, bounds.Max.X, bounds.Min.Y+top+height)
		op.GeoM.Translate(wave, 0)
		Instance{Image: src, Source: &rect, Options: op}.Draw(dst)
	}
}
