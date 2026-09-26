package motion

import (
	"fmt"
	"math"
)

// TableWarpRowsConfig samples horizontal placement from a compiled curve and
// combines it with one independent vertical bounce. Source rows may exceed
// the source image: the renderer's transparent crop keeps the authored timing.
type TableWarpRowsConfig struct {
	Curve              []float64
	Rows               int
	CounterStart, Step int
	XBase, HalfWidth   float64
	YBase, RowOffset   float64
	SourceHeight       float64
	BounceDivisor      float64
	Bounce             WaveClockConfig
}

// TableWarpRows implements SampledRowProgram and advances only when its host
// calls Advance. Every row of one draw reads the same bounce and curve cursor.
type TableWarpRows struct {
	config  TableWarpRowsConfig
	counter int
	bounce  *WaveClock
	value   float64
}

func NewTableWarpRows(c TableWarpRowsConfig) (*TableWarpRows, error) {
	if len(c.Curve) == 0 || len(c.Curve) > 1<<20 || c.Rows < 1 || c.Rows > 4096 ||
		c.SourceHeight <= 0 || c.BounceDivisor == 0 {
		return nil, fmt.Errorf("motion: invalid table-warped rows")
	}
	for _, v := range [...]float64{c.XBase, c.HalfWidth, c.YBase, c.RowOffset,
		c.SourceHeight, c.BounceDivisor} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("motion: nonfinite table-warped row setting")
		}
	}
	for _, value := range c.Curve {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite table-warped curve")
		}
	}
	bounce, err := NewWaveClock(c.Bounce)
	if err != nil {
		return nil, err
	}
	c.Curve = append([]float64(nil), c.Curve...)
	p := &TableWarpRows{config: c, counter: c.CounterStart, bounce: bounce}
	p.value = bounce.At(0)
	return p, nil
}

func (p *TableWarpRows) Sample(_ float64, _ int, row int) SampledRowPose {
	index := (p.counter + row) % len(p.config.Curve)
	if index < 0 {
		index += len(p.config.Curve)
	}
	return SampledRowPose{
		SourceY: float64(row), SourceHeight: p.config.SourceHeight,
		X:      p.config.XBase + p.config.Curve[index] - p.config.HalfWidth,
		Y:      p.config.YBase + p.value/p.config.BounceDivisor + float64(row) + p.config.RowOffset,
		ScaleX: 1, ScaleY: 1,
	}
}

func (p *TableWarpRows) Advance() {
	p.counter += p.config.Step
	p.bounce.Step()
	p.value = p.bounce.At(0)
}

func (p *TableWarpRows) Bounce() float64 { return p.value }
func (p *TableWarpRows) Counter() int    { return p.counter }
func (p *TableWarpRows) Reset() {
	p.counter = p.config.CounterStart
	p.bounce.Reset()
	p.value = p.bounce.At(0)
}
