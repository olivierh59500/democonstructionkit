package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// QuantizedLogoLayer selects a pre-rendered image using Bias+Depth*Gain.
// HideFirst suppresses frame zero, as in entrances that begin near invisibility.
type QuantizedLogoLayer struct {
	Image      *ebiten.Image
	FrameCount int
	Scales     []float64
	Bias, Gain float64
	HideFirst  bool
	Filter     ebiten.Filter
}

type CoupledLogoPairConfig struct {
	Motion             motion.CoupledLogoConfig
	Primary, Secondary QuantizedLogoLayer
	CenterX            float64
	Filter             ebiten.Filter
	Blend              ebiten.Blend
}

// CoupledLogoPair owns the scale banks and pair clock. Source images are
// borrowed; the cached frames are deallocated by Close.
type CoupledLogoPair struct {
	config CoupledLogoPairConfig
	motion *motion.CoupledLogoMotion
	frames [2][]*ebiten.Image
	poses  [2]motion.CoupledLogoPose
	order  [2]int
}

func NewCoupledLogoPair(c CoupledLogoPairConfig) (*CoupledLogoPair, error) {
	if math.IsNaN(c.CenterX) || math.IsInf(c.CenterX, 0) {
		return nil, fmt.Errorf("sprites: nonfinite logo pair center")
	}
	clock, err := motion.NewCoupledLogoMotion(c.Motion)
	if err != nil {
		return nil, err
	}
	pair := &CoupledLogoPair{config: c, motion: clock}
	for i, layer := range [...]QuantizedLogoLayer{c.Primary, c.Secondary} {
		if math.IsNaN(layer.Bias) || math.IsInf(layer.Bias, 0) || math.IsNaN(layer.Gain) || math.IsInf(layer.Gain, 0) {
			pair.Close()
			return nil, fmt.Errorf("sprites: nonfinite logo frame selection")
		}
		pair.frames[i], err = NewScaleFrames(ScaleFramesConfig{
			Source: layer.Image, Count: layer.FrameCount, Scales: layer.Scales, Filter: layer.Filter,
		})
		if err != nil {
			pair.Close()
			return nil, err
		}
	}
	pair.poses = clock.Poses()
	pair.updateOrder()
	return pair, nil
}

func (p *CoupledLogoPair) updateOrder() {
	if p.poses[1].Depth >= p.poses[0].Depth {
		p.order = [2]int{0, 1}
	} else {
		p.order = [2]int{1, 0}
	}
}

func (p *CoupledLogoPair) Update(kit.Frame) error {
	if err := p.motion.Step(); err != nil {
		return err
	}
	p.poses = p.motion.Poses()
	p.updateOrder()
	return nil
}

func (p *CoupledLogoPair) SetSpeedMultiplier(speed float64) error {
	return p.motion.SetSpeedMultiplier(speed)
}

func (p *CoupledLogoPair) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil {
		return
	}
	for _, index := range p.order {
		layer := p.config.Primary
		if index == 1 {
			layer = p.config.Secondary
		}
		frameIndex := p.FrameIndex(index)
		if frameIndex < 0 || layer.HideFirst && frameIndex == 0 {
			continue
		}
		frame := p.frames[index][frameIndex]
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = p.config.Filter, p.config.Blend
		op.GeoM.Translate(p.config.CenterX-float64(frame.Bounds().Dx()/2),
			p.poses[index].Y-float64(frame.Bounds().Dy())/2)
		dst.DrawImage(frame, &op)
	}
}

func (p *CoupledLogoPair) FrameIndex(index int) int {
	layer := p.config.Primary
	if index == 1 {
		layer = p.config.Secondary
	}
	raw := layer.Bias + p.poses[index].Depth*layer.Gain
	if math.IsNaN(raw) {
		return -1
	}
	if raw <= 0 {
		return 0
	}
	if raw >= float64(len(p.frames[index])-1) {
		return len(p.frames[index]) - 1
	}
	return int(raw)
}

func (p *CoupledLogoPair) Poses() [2]motion.CoupledLogoPose { return p.poses }
func (p *CoupledLogoPair) DrawOrder() [2]int                { return p.order }
func (p *CoupledLogoPair) Clock() *motion.CoupledLogoMotion { return p.motion }

func (p *CoupledLogoPair) Close() error {
	if p == nil {
		return nil
	}
	for i := range p.frames {
		for _, frame := range p.frames[i] {
			frame.Deallocate()
		}
		p.frames[i] = nil
	}
	return nil
}
