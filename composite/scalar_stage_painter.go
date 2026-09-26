package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/palette"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

type ScalarTintMode uint8

const (
	ScalarTintIdentity ScalarTintMode = iota
	ScalarTintRGB
	ScalarTintAlpha
	ScalarTintHSL
)

// ScalarTintConfig holds serializable formulas over the director's scalar
// value. RGB/HSL preserve source alpha; Alpha uses ColorScale.ScaleAlpha so
// premultiplied RGB fades together with alpha. HSL's H formula may read the
// secondary clock while Saturation and Lightness remain editable constants.
type ScalarTintConfig struct {
	Mode                  ScalarTintMode
	R, G, B               *motion.FormulaExpr
	A                     *motion.FormulaExpr
	H                     *motion.FormulaExpr
	Saturation, Lightness float64
}

// ScalarImagePass draws one borrowed image in declaration order. YFormula, if
// supplied, adds a scalar-derived offset to Y. ExprTime is the director value;
// ExprSecondaryTime is an independent caller-supplied position or cue.
type ScalarImagePass struct {
	Rule     ScalarStageRule
	Image    *ebiten.Image
	X, Y     float64
	YFormula *motion.FormulaExpr
	Tint     ScalarTintConfig
	Filter   ebiten.Filter
	Blend    ebiten.Blend
}

// ScalarStagePainterConfig binds a scalar stage director to ordered image
// materials. Backgrounds are selected by stage; nil clears to transparent.
type ScalarStagePainterConfig struct {
	Director    *timeline.ScalarStages
	Backgrounds []color.Color
	Passes      []ScalarImagePass
}

type scalarCompiledPass struct {
	config ScalarImagePass
	y      *motion.FormulaProgram
	colors [4]*motion.FormulaProgram
}

// ScalarStagePainter draws from the current director state without advancing
// it or allocating a new surface. The host owns the director, images and cues.
type ScalarStagePainter struct {
	director    *timeline.ScalarStages
	backgrounds []color.Color
	passes      [][]*scalarCompiledPass
}

func NewScalarStagePainter(config ScalarStagePainterConfig) (*ScalarStagePainter, error) {
	if config.Director == nil || len(config.Backgrounds) == 0 || len(config.Backgrounds) > 1024 || len(config.Passes) > 16384 {
		return nil, fmt.Errorf("composite: invalid scalar stage painter dimensions")
	}
	painter := &ScalarStagePainter{
		director:    config.Director,
		backgrounds: append([]color.Color(nil), config.Backgrounds...),
		passes:      make([][]*scalarCompiledPass, len(config.Backgrounds)),
	}
	for index, pass := range config.Passes {
		if pass.Image == nil || pass.Rule.From < 0 || pass.Rule.To < pass.Rule.From || pass.Rule.To >= len(painter.passes) ||
			pass.Rule.Compare > timeline.ScalarLessEqual || pass.Rule.DirectionSign < -1 || pass.Rule.DirectionSign > 1 ||
			math.IsNaN(pass.Rule.Threshold) || math.IsInf(pass.Rule.Threshold, 0) ||
			math.IsNaN(pass.X) || math.IsInf(pass.X, 0) || math.IsNaN(pass.Y) || math.IsInf(pass.Y, 0) ||
			pass.Tint.Mode > ScalarTintHSL {
			return nil, fmt.Errorf("composite: invalid scalar image pass %d", index)
		}
		compiled := &scalarCompiledPass{config: pass}
		var err error
		if pass.YFormula != nil {
			compiled.y, err = motion.CompileFormula(*pass.YFormula)
			if err != nil {
				return nil, fmt.Errorf("composite: pass %d Y formula: %w", index, err)
			}
		}
		if pass.Tint.Mode == ScalarTintRGB {
			for channel, expression := range [...]*motion.FormulaExpr{pass.Tint.R, pass.Tint.G, pass.Tint.B} {
				if expression == nil {
					return nil, fmt.Errorf("composite: missing RGB formula in pass %d", index)
				}
				compiled.colors[channel], err = motion.CompileFormula(*expression)
				if err != nil {
					return nil, fmt.Errorf("composite: pass %d color formula: %w", index, err)
				}
			}
		} else if pass.Tint.Mode == ScalarTintAlpha {
			if pass.Tint.A == nil {
				return nil, fmt.Errorf("composite: missing alpha formula in pass %d", index)
			}
			compiled.colors[3], err = motion.CompileFormula(*pass.Tint.A)
			if err != nil {
				return nil, fmt.Errorf("composite: pass %d alpha formula: %w", index, err)
			}
		} else if pass.Tint.Mode == ScalarTintHSL {
			if pass.Tint.H == nil || math.IsNaN(pass.Tint.Saturation) || math.IsInf(pass.Tint.Saturation, 0) ||
				math.IsNaN(pass.Tint.Lightness) || math.IsInf(pass.Tint.Lightness, 0) ||
				pass.Tint.Saturation < 0 || pass.Tint.Saturation > 1 || pass.Tint.Lightness < 0 || pass.Tint.Lightness > 1 {
				return nil, fmt.Errorf("composite: invalid HSL material in pass %d", index)
			}
			compiled.colors[0], err = motion.CompileFormula(*pass.Tint.H)
			if err != nil {
				return nil, fmt.Errorf("composite: pass %d hue formula: %w", index, err)
			}
		}
		for stage := pass.Rule.From; stage <= pass.Rule.To; stage++ {
			painter.passes[stage] = append(painter.passes[stage], compiled)
		}
	}
	return painter, nil
}

// Draw reads current stage/value/direction and an optional second scalar such
// as a bouncing sprite's Y. Repeated calls do not advance the presentation.
func (painter *ScalarStagePainter) Draw(dst *ebiten.Image, secondary float64) {
	if painter == nil || dst == nil {
		return
	}
	state := painter.director.State()
	if state.Stage < 0 || state.Stage >= len(painter.backgrounds) {
		return
	}
	if background := painter.backgrounds[state.Stage]; background != nil {
		dst.Fill(background)
	} else {
		dst.Clear()
	}
	for _, pass := range painter.passes[state.Stage] {
		if !pass.config.Rule.Matches(state.Stage, state.Value, state.Direction) {
			continue
		}
		x, y := pass.config.X, pass.config.Y
		if pass.y != nil {
			y += scalarStageValue(pass.y, state.Value, secondary)
		}
		if pass.config.Tint.Mode == ScalarTintIdentity && x == 0 && y == 0 && pass.config.Filter == 0 && pass.config.Blend == (ebiten.Blend{}) {
			dst.DrawImage(pass.config.Image, nil)
			continue
		}
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = pass.config.Filter, pass.config.Blend
		options.GeoM.Translate(x, y)
		switch pass.config.Tint.Mode {
		case ScalarTintRGB:
			options.ColorScale.Scale(
				float32(scalarStageValue(pass.colors[0], state.Value, secondary)),
				float32(scalarStageValue(pass.colors[1], state.Value, secondary)),
				float32(scalarStageValue(pass.colors[2], state.Value, secondary)), 1)
		case ScalarTintAlpha:
			options.ColorScale.ScaleAlpha(float32(scalarStageValue(pass.colors[3], state.Value, secondary)))
		case ScalarTintHSL:
			hue := scalarStageValue(pass.colors[0], state.Value, secondary)
			r, g, b := palette.HSLToRGB(hue, pass.config.Tint.Saturation, pass.config.Tint.Lightness)
			options.ColorScale.Scale(float32(r), float32(g), float32(b), 1)
		}
		dst.DrawImage(pass.config.Image, &options)
	}
}

func (painter *ScalarStagePainter) Director() *timeline.ScalarStages { return painter.director }

func scalarStageValue(program *motion.FormulaProgram, value, secondary float64) float64 {
	return program.AtWithSecondaryTime(value, secondary, 0, 0, 0, 0)
}
