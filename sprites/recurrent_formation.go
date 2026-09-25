package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// RecurrentFormationConfig places repeated sprites or logos with an X sine
// multiplied by a cosine envelope and two additive Y harmonics. IndexStep fields
// control phase spacing between instances; tick divisors control animation.
// OutputScale maps native positions into a parent viewport independently of
// each image's own Scale.
type RecurrentFormationConfig struct {
	Image                       *ebiten.Image
	Count                       int
	Origin                      motion.Point
	XAmplitude, YAmplitude      float64
	YSecondaryAmplitude         float64
	XDivisor, XSecondaryDivisor float64
	YDivisor, YSecondaryDivisor float64
	XIndexStep, XSecondaryStep  float64
	YIndexStep, YSecondaryStep  float64
	YSecondaryCos               bool // Select cosine for the second Y harmonic.
	ScaleX, ScaleY              float64
	OutputScaleX, OutputScaleY  float64
	Opacity                     float32 // Zero defaults to one.
	Filter                      ebiten.Filter
	Blend                       ebiten.Blend
}

// RecurrentFormation computes four trigonometric seeds per Update, then visits
// all poses using a sine/cosine recurrence. Draw reuses prepared poses and
// borrowed artwork without allocating an intermediate surface.
type RecurrentFormation struct {
	config RecurrentFormationConfig
	poses  []motion.Point
	steps  [4][2]float64
	ready  bool
}

func NewRecurrentFormation(config RecurrentFormationConfig) (*RecurrentFormation, error) {
	formation := &RecurrentFormation{}
	if err := formation.SetConfig(config); err != nil {
		return nil, err
	}
	return formation, nil
}

// SetConfig changes count, artwork or wave parameters at a cue. It retains
// pose capacity when possible and requires a subsequent Update before Draw.
func (formation *RecurrentFormation) SetConfig(config RecurrentFormationConfig) error {
	if config.Image == nil || config.Count < 0 || config.Count > 1_000_000 ||
		config.XDivisor == 0 || config.XSecondaryDivisor == 0 ||
		config.YDivisor == 0 || config.YSecondaryDivisor == 0 ||
		config.Opacity < 0 || config.Opacity > 1 {
		return fmt.Errorf("sprites: invalid recurrent formation dimensions")
	}
	for _, value := range [...]float64{
		config.Origin.X, config.Origin.Y, config.XAmplitude, config.YAmplitude, config.YSecondaryAmplitude,
		config.XDivisor, config.XSecondaryDivisor, config.YDivisor, config.YSecondaryDivisor,
		config.XIndexStep, config.XSecondaryStep, config.YIndexStep, config.YSecondaryStep,
		config.ScaleX, config.ScaleY, config.OutputScaleX, config.OutputScaleY, float64(config.Opacity),
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("sprites: nonfinite recurrent formation setting")
		}
	}
	if config.ScaleX == 0 {
		config.ScaleX = 1
	}
	if config.ScaleY == 0 {
		config.ScaleY = 1
	}
	if config.OutputScaleX == 0 {
		config.OutputScaleX = 1
	}
	if config.OutputScaleY == 0 {
		config.OutputScaleY = 1
	}
	if config.Opacity == 0 {
		config.Opacity = 1
	}
	if cap(formation.poses) < config.Count {
		formation.poses = make([]motion.Point, config.Count)
	} else {
		formation.poses = formation.poses[:config.Count]
	}
	formation.config = config
	for index, step := range [...]float64{config.XIndexStep, config.XSecondaryStep, config.YIndexStep, config.YSecondaryStep} {
		formation.steps[index][0], formation.steps[index][1] = math.Sincos(step)
	}
	formation.ready = false
	return nil
}

func (formation *RecurrentFormation) Update(tick float64) error {
	if math.IsNaN(tick) || math.IsInf(tick, 0) {
		return fmt.Errorf("sprites: nonfinite recurrent formation tick")
	}
	c := formation.config
	xSin, xCos := math.Sincos(tick / c.XDivisor)
	xSecondarySin, xSecondaryCos := math.Sincos(tick / c.XSecondaryDivisor)
	ySin, yCos := math.Sincos(tick / c.YDivisor)
	ySecondarySin, ySecondaryCos := math.Sincos(tick / c.YSecondaryDivisor)
	for index := range formation.poses {
		secondaryY := ySecondarySin
		if c.YSecondaryCos {
			secondaryY = ySecondaryCos
		}
		formation.poses[index] = motion.Point{
			X: c.Origin.X + c.XAmplitude*xSin*xSecondaryCos,
			Y: c.Origin.Y + c.YAmplitude*ySin + c.YSecondaryAmplitude*secondaryY,
		}
		xSin, xCos = recurrentSinCosForward(xSin, xCos, formation.steps[0][0], formation.steps[0][1])
		xSecondarySin, xSecondaryCos = recurrentSinCosForward(xSecondarySin, xSecondaryCos, formation.steps[1][0], formation.steps[1][1])
		ySin, yCos = recurrentSinCosForward(ySin, yCos, formation.steps[2][0], formation.steps[2][1])
		ySecondarySin, ySecondaryCos = recurrentSinCosForward(ySecondarySin, ySecondaryCos, formation.steps[3][0], formation.steps[3][1])
	}
	formation.ready = true
	return nil
}

func recurrentSinCosForward(sinValue, cosValue, sinStep, cosStep float64) (float64, float64) {
	return sinValue*cosStep + cosValue*sinStep, cosValue*cosStep - sinValue*sinStep
}

// Poses returns borrowed native positions for an inspector or alternate draw.
func (formation *RecurrentFormation) Poses() []motion.Point { return formation.poses }

func (formation *RecurrentFormation) Draw(dst *ebiten.Image) {
	if formation == nil || dst == nil || !formation.ready {
		return
	}
	c := formation.config
	for _, pose := range formation.poses {
		op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend}
		op.GeoM.Scale(c.ScaleX, c.ScaleY)
		op.GeoM.Translate(pose.X*c.OutputScaleX, pose.Y*c.OutputScaleY)
		if c.Opacity != 1 {
			op.ColorScale.ScaleAlpha(c.Opacity)
		}
		dst.DrawImage(c.Image, &op)
	}
}
