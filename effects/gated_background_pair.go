package effects

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// GatedBackgroundPairConfig layers two borrowed pictures over the shared
// coupled motion. Each image has its own editable crop, period and draw budget.
type GatedBackgroundPairConfig struct {
	Images      [2]*ebiten.Image
	Motion      motion.GatedBackgroundPairConfig
	Backgrounds [2]composite.BackgroundConfig
}

// GatedBackgroundPair owns no GPU images. It updates both clocks together and
// draws the two repeated backgrounds in their configured source order.
type GatedBackgroundPair struct {
	config     GatedBackgroundPairConfig
	motion     *motion.GatedBackgroundPair
	background [2]*composite.Background
}

func NewGatedBackgroundPair(c GatedBackgroundPairConfig) (*GatedBackgroundPair, error) {
	if c.Images[0] == nil || c.Images[1] == nil {
		return nil, fmt.Errorf("effects: missing paired background image")
	}
	clock, err := motion.NewGatedBackgroundPair(c.Motion)
	if err != nil {
		return nil, err
	}
	pair := &GatedBackgroundPair{config: c, motion: clock}
	for i, config := range c.Backgrounds {
		pair.background[i], err = composite.NewBackground(config)
		if err != nil {
			return nil, err
		}
	}
	return pair, nil
}

func (p *GatedBackgroundPair) Update(kit.Frame) error { return p.motion.Step() }

func (p *GatedBackgroundPair) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil {
		return
	}
	poses := p.motion.Poses()
	for i, pose := range poses {
		p.background[i].DrawAt(dst, p.config.Images[i], pose.X, pose.Y)
	}
}

func (p *GatedBackgroundPair) Motion() *motion.GatedBackgroundPair { return p.motion }

// Err reports a copy-budget failure from either background draw.
func (p *GatedBackgroundPair) Err() error {
	return errors.Join(p.background[0].Err(), p.background[1].Err())
}

func (p *GatedBackgroundPair) Close() error { return nil }
