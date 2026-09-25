package scrolling

import (
	"errors"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// Pseudo3DBank is one independently moving text lane. InvertScale folds its
// apparent depth around ScaleSum; zero Speed uses the shared transport speed.
type Pseudo3DBank struct {
	Text         string
	BaseY        float64
	Start, Speed float64
	InvertScale  bool
}

// Pseudo3DPose contains editable harmonic terms and clipping. Lag and Rate
// retain authored arithmetic order, while IndexStep controls the exact sine/
// cosine recurrence used as glyphs are visited in reverse order.
type Pseudo3DPose struct {
	ZLag, ZRate, ZIndexStep      float64
	XTimeRate, XIndexRate        float64
	YLag, YRate, YIndexStep      float64
	ZBase, ZAmplitude, ScaleSum  float64
	HorizontalRate, VerticalRate float64
	XOrigin, XAmplitude          float64
	YAmplitude, YDepth           float64
	OutputScaleX, CullMargin     float64
	Alpha                        float32
}

// Pseudo3DConfig groups any number of text lanes under one fixed-step clock.
// Use Face for atlas metrics or Atlas for pre-sliced glyph images. The latter
// retains each source glyph crop and falls back to the atlas space image.
type Pseudo3DConfig struct {
	Banks                      []Pseudo3DBank
	Face                       Face
	Atlas                      *Atlas
	Advance, PixelsPerUpdate   float64
	Visible                    int
	TicksPerSecond, TimeOffset float64
	Width, Height              float64
	UseFrameTime               bool
	Pose                       Pseudo3DPose
}

// Pseudo3D owns the four (or any number of) scrolling programs and keeps
// transport, depth and wave clocks separate from Draw. No GPU surface is owned.
type Pseudo3D struct {
	config     Pseudo3DConfig
	programs   []*Scrolling
	positions  []float64
	speeds     []float64
	lengths    []int
	tick       uint64
	time       float64
	multiplier float64
	zSin, zCos float64
	xSin, xCos float64
	ySin, yCos float64
}

func NewPseudo3D(config Pseudo3DConfig) (*Pseudo3D, error) {
	if len(config.Banks) == 0 || len(config.Banks) > 64 || config.Visible < 1 || config.Visible > 1024 ||
		config.Advance <= 0 || config.PixelsPerUpdate < 0 || config.TicksPerSecond <= 0 ||
		config.Width <= 0 || config.Height <= 0 || config.Pose.Alpha < 0 || config.Pose.Alpha > 1 ||
		config.Atlas == nil && (config.Face.Atlas == nil || config.Face.Metrics == nil) {
		return nil, fmt.Errorf("scrolling: invalid pseudo-3d dimensions or font")
	}
	terms := [...]float64{
		config.Advance, config.PixelsPerUpdate, config.TicksPerSecond, config.TimeOffset, config.Width, config.Height,
		config.Pose.ZLag, config.Pose.ZRate, config.Pose.ZIndexStep,
		config.Pose.XTimeRate, config.Pose.XIndexRate, config.Pose.YLag, config.Pose.YRate, config.Pose.YIndexStep,
		config.Pose.ZBase, config.Pose.ZAmplitude, config.Pose.ScaleSum,
		config.Pose.HorizontalRate, config.Pose.VerticalRate,
		config.Pose.XOrigin, config.Pose.XAmplitude, config.Pose.YAmplitude, config.Pose.YDepth,
		config.Pose.OutputScaleX, config.Pose.CullMargin, float64(config.Pose.Alpha),
	}
	for _, term := range terms {
		if math.IsNaN(term) || math.IsInf(term, 0) {
			return nil, fmt.Errorf("scrolling: nonfinite pseudo-3d setting")
		}
	}
	if config.Pose.OutputScaleX <= 0 || config.Pose.CullMargin < 0 || config.Pose.ScaleSum <= 0 {
		return nil, fmt.Errorf("scrolling: invalid pseudo-3d projection")
	}
	program := &Pseudo3D{config: config, time: config.TimeOffset, multiplier: 1}
	program.config.Banks = append([]Pseudo3DBank(nil), config.Banks...)
	program.programs = make([]*Scrolling, len(config.Banks))
	program.positions = make([]float64, len(config.Banks))
	program.speeds = make([]float64, len(config.Banks))
	program.lengths = make([]int, len(config.Banks))
	program.zSin, program.zCos = math.Sincos(config.Pose.ZIndexStep)
	program.xSin, program.xCos = math.Sincos(config.Pose.XIndexRate)
	program.ySin, program.yCos = math.Sincos(config.Pose.YIndexStep)
	for index, bank := range config.Banks {
		if !finite(bank.BaseY) || !finite(bank.Start) || !finite(bank.Speed) || bank.Start < 0 {
			program.Close()
			return nil, fmt.Errorf("scrolling: invalid pseudo-3d bank %d", index)
		}
		program.lengths[index] = len([]rune(bank.Text))
		program.positions[index] = bank.Start
		program.speeds[index] = config.PixelsPerUpdate
		if bank.Speed != 0 {
			program.speeds[index] = bank.Speed
		}
		if bank.Text == "" {
			continue
		}
		var err error
		if config.Atlas != nil {
			glyphs := make([]*ebiten.Image, 0, program.lengths[index])
			for _, r := range bank.Text {
				img, _, ok := config.Atlas.Glyph(r)
				if !ok {
					img, _, _ = config.Atlas.Glyph(' ')
				}
				glyphs = append(glyphs, img)
			}
			program.programs[index], err = FromImages(glyphs, config.Advance)
		} else {
			program.programs[index], err = New(Config{Text: bank.Text, Advance: config.Advance,
				Fonts: map[string]Face{"default": config.Face}})
		}
		if err != nil {
			program.Close()
			return nil, fmt.Errorf("scrolling: pseudo-3d bank %d: %w", index, err)
		}
	}
	return program, nil
}

func (p *Pseudo3D) Update(frame kit.Frame) error {
	nextTime := float64(p.tick+1)/p.config.TicksPerSecond + p.config.TimeOffset
	if p.config.UseFrameTime {
		if !finite(frame.Time) {
			return fmt.Errorf("scrolling: nonfinite pseudo-3d time")
		}
		nextTime = frame.Time + p.config.TimeOffset
	}
	p.tick++
	p.time = nextTime
	for index, length := range p.lengths {
		if length == 0 {
			p.positions[index] = 0
			continue
		}
		p.positions[index] += p.speeds[index] * p.multiplier
		if limit := float64(length) * p.config.Advance; p.positions[index] >= limit {
			p.positions[index] -= limit
		}
	}
	return nil
}

func (p *Pseudo3D) Draw(dst *ebiten.Image) {
	if p == nil || dst == nil {
		return
	}
	t := p.time
	wave := math.Sin(t*p.config.Pose.HorizontalRate)*.5 + .5
	horizontalWave := math.Sqrt(1 - wave*wave)
	verticalWave := math.Sin(t*p.config.Pose.VerticalRate)*.5 + .5
	for index := range p.programs {
		p.drawBank(dst, index, t, horizontalWave, verticalWave)
	}
}

func (p *Pseudo3D) drawBank(dst *ebiten.Image, bankIndex int, t, horizontalWave, verticalWave float64) {
	program := p.programs[bankIndex]
	if program == nil {
		return
	}
	c := p.config
	bank := c.Banks[bankIndex]
	position := p.positions[bankIndex]
	first := int(position / c.Advance)
	last := first + c.Visible
	zs, zc := math.Sincos((t + float64(last)*c.Pose.ZLag) * c.Pose.ZRate)
	xs, xc := math.Sincos(t*c.Pose.XTimeRate + float64(last)*c.Pose.XIndexRate)
	ys, yc := math.Sincos((t + float64(last)*c.Pose.YLag) * c.Pose.YRate)
	advance := func() {
		zs, zc = pseudoSinCosBackward(zs, zc, p.zSin, p.zCos)
		xs, xc = pseudoSinCosBackward(xs, xc, p.xSin, p.xCos)
		ys, yc = pseudoSinCosBackward(ys, yc, p.ySin, p.yCos)
	}
	for index := last; index >= p.lengths[bankIndex]; index-- {
		advance()
	}
	state := IdentityState()
	state.First, state.End, state.Reverse = first, last+1, true
	state.Time, state.Position = t, position
	state.Map = func(sample Sample, op *ebiten.DrawImageOptions) bool {
		z, xSin, ySin := zs*c.Pose.ZAmplitude+c.Pose.ZBase, xs, ys
		advance()
		x := math.Floor((float64(sample.Index)*c.Advance - c.Pose.XOrigin - xSin*c.Pose.XAmplitude*horizontalWave - position) * c.Pose.OutputScaleX)
		y := math.Floor(ySin*c.Pose.YAmplitude*verticalWave + bank.BaseY - z*c.Pose.YDepth)
		scale := z
		if bank.InvertScale {
			scale = c.Pose.ScaleSum - z
		}
		if x < -c.Pose.CullMargin || x > c.Width+c.Pose.CullMargin || y < -c.Pose.CullMargin || y > c.Height+c.Pose.CullMargin || scale <= .1 {
			return false
		}
		op.GeoM.Reset()
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
		op.ColorScale.Scale(1, 1, 1, c.Pose.Alpha)
		return true
	}
	program.DrawAt(dst, state)
}

func pseudoSinCosBackward(sinValue, cosValue, sinStep, cosStep float64) (float64, float64) {
	return sinValue*cosStep - cosValue*sinStep, cosValue*cosStep + sinValue*sinStep
}

// BankPosition reports the current pixel transport of one text bank.
func (p *Pseudo3D) BankPosition(index int) float64 { return p.positions[index] }

// SetBankSpeed changes one bank without discarding its existing phase.
func (p *Pseudo3D) SetBankSpeed(index int, speed float64) error {
	if index < 0 || index >= len(p.speeds) || !finite(speed) {
		return fmt.Errorf("scrolling: invalid pseudo-3d bank speed")
	}
	p.speeds[index] = speed
	return nil
}

// SetSpeedMultiplier changes all bank speeds without resetting their phases.
func (p *Pseudo3D) SetSpeedMultiplier(multiplier float64) error {
	if !finite(multiplier) || multiplier < 0 {
		return fmt.Errorf("scrolling: invalid pseudo-3d speed multiplier")
	}
	p.multiplier = multiplier
	return nil
}

// SetBankPosition seeks one lane in its original pen coordinates.
func (p *Pseudo3D) SetBankPosition(index int, position float64) error {
	if index < 0 || index >= len(p.positions) || !finite(position) || position < 0 {
		return fmt.Errorf("scrolling: invalid pseudo-3d bank position")
	}
	p.positions[index] = position
	return nil
}

// SetTimeOffset changes a global phase or music synchronization cue.
func (p *Pseudo3D) SetTimeOffset(offset float64) error {
	if !finite(offset) {
		return fmt.Errorf("scrolling: invalid pseudo-3d time offset")
	}
	p.config.TimeOffset = offset
	return nil
}

func (p *Pseudo3D) Close() error {
	if p == nil {
		return nil
	}
	var err error
	for index, program := range p.programs {
		if program != nil {
			err = errors.Join(err, program.Close())
			p.programs[index] = nil
		}
	}
	return err
}
