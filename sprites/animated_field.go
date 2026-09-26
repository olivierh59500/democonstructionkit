package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// AnimatedFieldConfig binds an independent frame clock for each sprite to a
// borrowed image sequence. Select maps fractional phases to atlas indices;
// nil uses nearest-frame rounding and ignores out-of-range indices.
type AnimatedFieldConfig struct {
	Motion           motion.FrameFieldConfig
	Frames           []*ebiten.Image
	OffsetX, OffsetY float64
	Select           func(phase float64) int
	Filter           ebiten.Filter
	Blend            ebiten.Blend
}

// AnimatedField owns the particle clocks and borrowed frame lookup. Drawing
// does not advance time; a screen can place it between any other effect layers.
type AnimatedField struct {
	motion *motion.FrameField
	config AnimatedFieldConfig
}

func NewAnimatedField(c AnimatedFieldConfig) (*AnimatedField, error) {
	if len(c.Frames) == 0 || len(c.Frames) > 1<<16 ||
		math.IsNaN(c.OffsetX) || math.IsInf(c.OffsetX, 0) ||
		math.IsNaN(c.OffsetY) || math.IsInf(c.OffsetY, 0) {
		return nil, fmt.Errorf("sprites: invalid animated field frames or placement")
	}
	field, err := motion.NewFrameField(c.Motion)
	if err != nil {
		return nil, err
	}
	c.Frames = append([]*ebiten.Image(nil), c.Frames...)
	if c.Select == nil {
		c.Select = func(phase float64) int { return int(math.Round(phase)) }
	}
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendSourceOver
	}
	return &AnimatedField{motion: field, config: c}, nil
}

func (f *AnimatedField) Update(kit.Frame) error { return f.motion.Step() }

func (f *AnimatedField) Draw(dst *ebiten.Image) {
	if f == nil || dst == nil {
		return
	}
	c := f.config
	for _, p := range f.motion.Samples() {
		index := c.Select(p.Phase)
		if index < 0 || index >= len(c.Frames) || c.Frames[index] == nil {
			continue
		}
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.GeoM.Translate(c.OffsetX+p.X, c.OffsetY+p.Y)
		dst.DrawImage(c.Frames[index], &op)
	}
}

// Motion exposes live speed and sample state for editor or music cues.
func (f *AnimatedField) Motion() *motion.FrameField { return f.motion }
