package effects

import (
	"fmt"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// BallMotion is one sampled pose of a projected sprite train. AngleSpacing is
// in degrees between successive balls; SpinSpeed rotates the shared train.
type BallMotion struct {
	SpinSpeed, Height, AngleSpacing, Radius float64
}

// BallMotionFunc can compute a pose from time and slot, independent of the
// source image. A sequence may mix constant, sine and caller-defined motions.
type BallMotionFunc func(seconds float64, slot int) BallMotion

// BallSequence selects two motion programs and their interpolation weight.
// Named presets use fixed segments; a caller may supply any ordering or cue.
type BallSequence func(seconds float64, slot int) (from, to int, alpha float64)

// BallRotationStyle retains either the original two-step 3D rotation or the
// direct trigonometric form. The results can differ by one rounded RGB level.
type BallRotationStyle uint8

const (
	BallRotationSequential BallRotationStyle = iota
	BallRotationDirect
)

// BallDepthOrder controls how projected sprites are submitted to the painter.
type BallDepthOrder uint8

const (
	BallFarFirst BallDepthOrder = iota
	BallSlotOrder
)

// ProjectedBallTrainConfig describes the complete controller, projection,
// shadow palette and draw order. Images are borrowed. Program callbacks are
// Go extension points; presets supply named sequences for editor-friendly use.
// Update advances by TimeStep; AdvanceAt samples an external scene clock.
type ProjectedBallTrainConfig struct {
	Ball                                     *ebiten.Image
	Shadows                                  []*ebiten.Image
	Count                                    int
	Programs                                 []BallMotionFunc
	Sequence                                 BallSequence
	TimeStep, SpinStep                       float64
	Perspective, CenterX, CenterY            float64
	Scale, ShadowPlaneY                      float64
	BallAnchorX, BallAnchorY                 float64
	ShadowAnchorX, ShadowAnchorY             float64
	ShadowLift, ShadowLiftReference          float64
	ShadeReference, ShadeSteps, ShadeDivisor float64
	ShadowOrder, BallOrder                   BallDepthOrder
	Rotation                                 BallRotationStyle
	Filter                                   ebiten.Filter
}

// DefaultProjectedBallTrainConfig supplies the 3D DOC projection geometry.
// The caller provides images, movement programs and a sequence.
func DefaultProjectedBallTrainConfig() ProjectedBallTrainConfig {
	return ProjectedBallTrainConfig{
		Count: 4, TimeStep: 1.0 / 60, SpinStep: .15,
		Perspective: 400, CenterX: 384, CenterY: 310, Scale: .7,
		ShadowPlaneY: 60, BallAnchorX: 32, BallAnchorY: 32,
		ShadowAnchorX: 32, ShadowAnchorY: 8,
		ShadowLift: 26, ShadowLiftReference: 1,
		ShadeReference: .5, ShadeSteps: 10, ShadeDivisor: 2,
	}
}

func (c ProjectedBallTrainConfig) Validate() error {
	if c.Ball == nil || len(c.Shadows) == 0 || c.Count <= 0 || c.Count > 65535 ||
		len(c.Programs) == 0 || c.Sequence == nil || c.Perspective <= 0 ||
		c.Scale <= 0 || c.TimeStep < 0 || c.ShadeSteps <= 0 || c.ShadeDivisor <= 0 ||
		c.ShadowOrder > BallSlotOrder || c.BallOrder > BallSlotOrder ||
		c.Rotation > BallRotationDirect {
		return fmt.Errorf("effects: invalid projected ball train configuration")
	}
	for _, img := range c.Shadows {
		if img == nil {
			return fmt.Errorf("effects: missing projected ball shadow image")
		}
	}
	for _, v := range []float64{
		c.TimeStep, c.SpinStep, c.Perspective, c.CenterX, c.CenterY,
		c.Scale, c.ShadowPlaneY, c.BallAnchorX, c.BallAnchorY,
		c.ShadowAnchorX, c.ShadowAnchorY, c.ShadowLift,
		c.ShadowLiftReference, c.ShadeReference, c.ShadeSteps, c.ShadeDivisor,
	} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("effects: non-finite projected ball parameter")
		}
	}
	return nil
}

// ProjectedBall is the current image-space position and depth of one slot.
type ProjectedBall struct{ X, Y, Z, Scale float64 }

// ProjectedBallTrain retains a bounded amount of state per ball. AdvanceAt
// updates angles and projected positions; Draw only submits cached images.
type ProjectedBallTrain struct {
	config  ProjectedBallTrainConfig
	balls   []ProjectedBall
	shadows []ProjectedBall
	angles  []float64
	order   []int
	phase   float64
	time    float64
}

func NewProjectedBallTrain(c ProjectedBallTrainConfig) (*ProjectedBallTrain, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.Programs = append([]BallMotionFunc(nil), c.Programs...)
	c.Shadows = append([]*ebiten.Image(nil), c.Shadows...)
	p := &ProjectedBallTrain{
		config: c, balls: make([]ProjectedBall, c.Count),
		shadows: make([]ProjectedBall, c.Count), angles: make([]float64, c.Count),
		order: make([]int, c.Count),
	}
	if err := p.PoseAt(0); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *ProjectedBallTrain) Update(kit.Frame) error { return p.AdvanceAt(p.time + p.config.TimeStep) }

// AdvanceAt advances the shared rotation once and samples every ball at the
// supplied absolute time. It is suitable for a scene that includes intro time.
func (p *ProjectedBallTrain) AdvanceAt(seconds float64) error {
	return p.sample(seconds, true)
}

// PoseAt changes the sampled time without stepping rotation, for example on
// the first visible frame immediately after an intro-to-main transition.
func (p *ProjectedBallTrain) PoseAt(seconds float64) error {
	return p.sample(seconds, false)
}

func (p *ProjectedBallTrain) sample(seconds float64, rotate bool) error {
	if p == nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return fmt.Errorf("effects: invalid projected ball clock")
	}
	c := p.config
	for slot := range p.balls {
		from, to, alpha := c.Sequence(seconds, slot)
		if from < 0 || from >= len(c.Programs) || to < 0 || to >= len(c.Programs) ||
			math.IsNaN(alpha) || math.IsInf(alpha, 0) {
			return fmt.Errorf("effects: projected ball sequence selected an invalid program")
		}
		a, b := c.Programs[from](seconds, slot), c.Programs[to](seconds, slot)
		motion := BallMotion{
			SpinSpeed:    a.SpinSpeed*(1-alpha) + b.SpinSpeed*alpha,
			Height:       a.Height*(1-alpha) + b.Height*alpha,
			AngleSpacing: a.AngleSpacing*(1-alpha) + b.AngleSpacing*alpha,
			Radius:       a.Radius*(1-alpha) + b.Radius*alpha,
		}
		for _, value := range []float64{motion.SpinSpeed, motion.Height, motion.AngleSpacing, motion.Radius} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("effects: non-finite projected ball motion")
			}
		}
		if rotate {
			p.phase += (math.Pi * 2 / 360) * motion.SpinSpeed * c.SpinStep
			p.phase = math.Mod(p.phase, math.Pi*2)
			p.angles[slot] = p.phase
		}
		if !p.project(slot, motion) {
			return fmt.Errorf("effects: projected ball crossed the camera plane")
		}
	}
	p.time = seconds
	p.sort()
	return nil
}

func (p *ProjectedBallTrain) project(slot int, m BallMotion) bool {
	c := p.config
	var angle, x, z float64
	if c.Rotation == BallRotationDirect {
		angle = (math.Pi * 2 / 360) * (m.AngleSpacing * float64(slot))
		x, z = m.Radius*math.Cos(angle), -m.Radius*math.Sin(angle)
		x, z = x*math.Cos(p.angles[slot])+z*math.Sin(p.angles[slot]), z*math.Cos(p.angles[slot])-x*math.Sin(p.angles[slot])
	} else {
		angle = (math.Pi * 2 / 360) * m.AngleSpacing * float64(slot)
		z = 0*math.Cos(angle) - m.Radius*math.Sin(angle)
		x = 0*math.Sin(angle) + m.Radius*math.Cos(angle)
		z, x = z*math.Cos(p.angles[slot])-x*math.Sin(p.angles[slot]), z*math.Sin(p.angles[slot])+x*math.Cos(p.angles[slot])
	}
	if c.Perspective+z == 0 {
		return false
	}
	scale := c.Perspective / (c.Perspective + z)
	p.balls[slot] = ProjectedBall{X: x*scale + c.CenterX, Y: m.Height*scale + c.CenterY, Z: z, Scale: scale * c.Scale}
	p.shadows[slot] = ProjectedBall{X: x*scale + c.CenterX, Y: c.ShadowPlaneY*scale + c.CenterY, Z: z, Scale: scale * c.Scale}
	return true
}

func (p *ProjectedBallTrain) sort() {
	for i := range p.order {
		p.order[i] = i
	}
	if len(p.order) > 32 {
		slices.SortStableFunc(p.order, func(a, b int) int {
			za, zb := p.balls[a].Z, p.balls[b].Z
			if za > zb {
				return -1
			}
			if za < zb {
				return 1
			}
			return 0
		})
		return
	}
	// This pairwise swap matches the original depth-order behavior, including
	// equal-depth ties. Shadow slot order remains independently selectable.
	for i := 0; i < len(p.order)-1; i++ {
		for j := i + 1; j < len(p.order); j++ {
			if p.balls[p.order[i]].Z < p.balls[p.order[j]].Z {
				p.order[i], p.order[j] = p.order[j], p.order[i]
			}
		}
	}
}

// Phase and Angle expose the current rotation for cue synchronization.
func (p *ProjectedBallTrain) Phase() float64 { return p.phase }
func (p *ProjectedBallTrain) Angle(slot int) float64 {
	if p == nil || slot < 0 || slot >= len(p.angles) {
		return 0
	}
	return p.angles[slot]
}

// Ball returns a projected slot without exposing the component's buffers.
func (p *ProjectedBallTrain) Ball(slot int) ProjectedBall {
	if p == nil || slot < 0 || slot >= len(p.balls) {
		return ProjectedBall{}
	}
	return p.balls[slot]
}

func (p *ProjectedBallTrain) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil {
		return
	}
	c := p.config
	for i := range p.shadows {
		idx := i
		if c.ShadowOrder == BallFarFirst {
			idx = p.order[i]
		}
		shadow := p.shadows[idx]
		shade := int(((shadow.Scale - c.ShadeReference) * c.ShadeSteps) / c.ShadeDivisor)
		shade = len(c.Shadows) - 1 - max(0, min(len(c.Shadows)-1, shade))
		lift := math.Min(1, math.Max(0, c.ShadowLiftReference-shadow.Scale)) * c.ShadowLift
		op := ebiten.DrawImageOptions{Filter: c.Filter}
		op.GeoM.Scale(shadow.Scale, shadow.Scale)
		op.GeoM.Translate(shadow.X-c.ShadowAnchorX, shadow.Y-c.ShadowAnchorY-lift)
		dst.DrawImage(c.Shadows[shade], &op)
	}
	for i := range p.balls {
		idx := i
		if c.BallOrder == BallFarFirst {
			idx = p.order[i]
		}
		ball := p.balls[idx]
		op := ebiten.DrawImageOptions{Filter: c.Filter}
		op.GeoM.Scale(ball.Scale, ball.Scale)
		op.GeoM.Translate(ball.X-c.BallAnchorX, ball.Y-c.BallAnchorY)
		dst.DrawImage(c.Ball, &op)
	}
}

// Close releases only per-instance slices; source images remain caller-owned.
func (p *ProjectedBallTrain) Close() error {
	if p != nil {
		p.balls, p.shadows, p.angles, p.order = nil, nil, nil, nil
	}
	return nil
}
