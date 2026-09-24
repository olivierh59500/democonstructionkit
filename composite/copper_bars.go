package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// CopperClock selects how two table phases advance. MaskedClock matches
// integer counters with a power-of-two wrap. SingleWrapClock keeps fractional
// phases and performs one authored wrap after each logical update.
type CopperClock uint8

const (
	MaskedClock CopperClock = iota
	SingleWrapClock
)

// CopperDraw selects the source-sampling primitive. CopperQuads preserves
// batched raster geometry; CopperImages preserves DrawImage strip sampling.
type CopperDraw uint8

const (
	CopperQuads CopperDraw = iota
	CopperImages
)

// CopperBarsConfig describes a complete bank of animated raster bars. Offsets
// is sampled by two clocks at independently spaced row indices. BaseX and
// XShift transform the summed offset into a destination position. A bar uses
// one source strip but may stretch down to Height, preserving authored overlap.
// Image and Offsets are borrowed at construction; the table is copied once.
type CopperBarsConfig struct {
	Image                  *ebiten.Image
	Offsets                []int
	Height, Count          int
	RowStep, SourceStep    int
	SourcePeriod, SourceY  int
	BaseX, XShift          int
	PhaseA, PhaseB         float64
	VelocityA, VelocityB   float64
	IndexStepA, IndexStepB int
	Clock                  CopperClock
	DrawMode               CopperDraw
	Filter                 ebiten.Filter
	Blend                  ebiten.Blend
}

// CopperBars owns the two phase clocks, source-strip cache and bounded painter.
// Draw does not advance time, so the same image can be used on several layers.
type CopperBars struct {
	config                CopperBarsConfig
	phaseA, phaseB, speed float64
	indexMask             int
	strips                []*ebiten.Image
	batch                 *QuadBatch
}

func NewCopperBars(c CopperBarsConfig) (*CopperBars, error) {
	if c.Image == nil || len(c.Offsets) < 1 || len(c.Offsets) > 1<<20 || c.Height < 1 || c.Height > 8192 ||
		c.Count < 1 || c.Count > 16383 || c.RowStep < 1 || c.SourceStep < 1 || c.SourcePeriod < 1 ||
		c.SourceY < 0 || c.SourceY > c.Image.Bounds().Dy()-c.SourcePeriod || c.SourcePeriod%c.SourceStep != 0 ||
		c.RowStep > 1<<20 || c.XShift < 0 || c.XShift > 31 || c.Clock > SingleWrapClock || c.DrawMode > CopperImages ||
		c.IndexStepA < -(1<<20) || c.IndexStepA > 1<<20 || c.IndexStepB < -(1<<20) || c.IndexStepB > 1<<20 ||
		c.BaseX < -(1<<29) || c.BaseX > 1<<29 {
		return nil, fmt.Errorf("composite: invalid copper bar image or geometry")
	}
	if c.Clock == MaskedClock && len(c.Offsets)&(len(c.Offsets)-1) != 0 {
		return nil, fmt.Errorf("composite: masked copper clock needs a power-of-two table")
	}
	for _, v := range [...]float64{c.PhaseA, c.PhaseB, c.VelocityA, c.VelocityB} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1<<20 {
			return nil, fmt.Errorf("composite: invalid copper clock")
		}
	}
	for _, value := range c.Offsets {
		if value < -(1<<29) || value > 1<<29 {
			return nil, fmt.Errorf("composite: copper offset exceeds coordinate range")
		}
	}
	c.Offsets = append([]int(nil), c.Offsets...)
	b := &CopperBars{config: c, phaseA: c.PhaseA, phaseB: c.PhaseB, speed: 1, indexMask: -1}
	if len(c.Offsets)&(len(c.Offsets)-1) == 0 {
		b.indexMask = len(c.Offsets) - 1
	}
	if c.DrawMode == CopperQuads {
		b.batch = NewQuadBatch(c.Count)
		b.batch.Options.Filter, b.batch.Options.Blend = c.Filter, c.Blend
	} else {
		for y := 0; y < c.SourcePeriod; y += c.SourceStep {
			src := image.Rect(c.Image.Bounds().Min.X, c.Image.Bounds().Min.Y+c.SourceY+y,
				c.Image.Bounds().Max.X, c.Image.Bounds().Min.Y+c.SourceY+y+c.SourceStep)
			b.strips = append(b.strips, c.Image.SubImage(src).(*ebiten.Image))
		}
	}
	return b, nil
}

// SetSpeed changes phase advance without resetting either clock. Zero freezes.
func (b *CopperBars) SetSpeed(speed float64) error {
	if b == nil || math.IsNaN(speed) || math.IsInf(speed, 0) || speed < 0 || speed > 1<<20 {
		return fmt.Errorf("composite: invalid copper speed")
	}
	b.speed = speed
	return nil
}

// Phases exposes the current transport state for an editor or synchronization.
func (b *CopperBars) Phases() (float64, float64) { return b.phaseA, b.phaseB }

func (b *CopperBars) Update(kit.Frame) error {
	if b == nil {
		return fmt.Errorf("composite: nil copper bars")
	}
	c, period := b.config, float64(len(b.config.Offsets))
	if c.Clock == MaskedClock {
		mask := len(c.Offsets) - 1
		b.phaseA = float64((int(b.phaseA) + int(c.VelocityA*b.speed)) & mask)
		b.phaseB = float64((int(b.phaseB) + int(c.VelocityB*b.speed)) & mask)
		return nil
	}
	b.phaseA += c.VelocityA * b.speed
	b.phaseB += c.VelocityB * b.speed
	if b.phaseA >= period {
		b.phaseA -= period
		if b.phaseA >= period {
			b.phaseA = math.Mod(b.phaseA, period)
		}
	} else if b.phaseA < 0 {
		b.phaseA += period
		if b.phaseA < 0 {
			b.phaseA = math.Mod(b.phaseA, period) + period
		}
	}
	if b.phaseB >= period {
		b.phaseB -= period
		if b.phaseB >= period {
			b.phaseB = math.Mod(b.phaseB, period)
		}
	} else if b.phaseB < 0 {
		b.phaseB += period
		if b.phaseB < 0 {
			b.phaseB = math.Mod(b.phaseB, period) + period
		}
	}
	return nil
}

func (b *CopperBars) Draw(dst *ebiten.Image) {
	if b == nil || dst == nil {
		return
	}
	c := b.config
	if b.batch != nil {
		b.batch.Begin(dst, c.Image)
	}
	for i := 0; i < c.Count; i++ {
		y, height := i*c.RowStep, c.Height-i*c.RowStep
		if height <= 0 {
			break
		}
		first := b.index(int(b.phaseA) + i*c.IndexStepA)
		second := b.index(int(b.phaseB) + i*c.IndexStepB)
		x := (c.Offsets[first] + c.Offsets[second] + c.BaseX) >> c.XShift
		sourceY := c.SourceY + i*c.SourceStep%c.SourcePeriod
		if b.batch != nil {
			src := image.Rect(c.Image.Bounds().Min.X, c.Image.Bounds().Min.Y+sourceY,
				c.Image.Bounds().Max.X, c.Image.Bounds().Min.Y+sourceY+c.SourceStep)
			b.batch.Rect(src, float32(x), float32(y), float32(c.Image.Bounds().Dx()), float32(height))
			continue
		}
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.GeoM.Scale(1, float64(height)/float64(c.SourceStep))
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(b.strips[(i*c.SourceStep%c.SourcePeriod)/c.SourceStep], &op)
	}
	if b.batch != nil {
		b.batch.Flush()
	}
}

func (b *CopperBars) index(value int) int {
	if b.indexMask >= 0 {
		return value & b.indexMask
	}
	return copperIndex(value, len(b.config.Offsets))
}

func copperIndex(value, period int) int {
	value %= period
	if value < 0 {
		value += period
	}
	return value
}
