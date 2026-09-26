package presets

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// SpreadpointBallPathConfig keeps the two screen-space waves and their
// source-order arithmetic editable. Snap applies floor(value+0.5) per axis.
type SpreadpointBallPathConfig struct {
	BaseX, BaseY              float64
	AmplitudeX, AmplitudeY    float64
	FrequencyX, FrequencyY    float64
	TimeDivisor, IndexDivisor float64
	Snap                      bool
}

func DefaultSpreadpointBallPath() SpreadpointBallPathConfig {
	return SpreadpointBallPathConfig{
		BaseX: 203, BaseY: 79, AmplitudeX: 151, AmplitudeY: 50,
		FrequencyX: 5, FrequencyY: 8, TimeDivisor: 71, IndexDivisor: 47, Snap: true,
	}
}

// SpreadpointBallFormula compiles a data-only path usable with any sprite bank.
func SpreadpointBallFormula(c SpreadpointBallPathConfig) (*motion.FormulaFormation, error) {
	for _, value := range [...]float64{c.BaseX, c.BaseY, c.AmplitudeX, c.AmplitudeY,
		c.FrequencyX, c.FrequencyY, c.TimeDivisor, c.IndexDivisor} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("presets: nonfinite Spreadpoint ball path")
		}
	}
	if c.TimeDivisor == 0 || c.IndexDivisor == 0 {
		return nil, fmt.Errorf("presets: zero Spreadpoint ball divisor")
	}
	phase := motion.ExprAdd(
		motion.ExprDiv(motion.ExprTime(), motion.ExprConst(c.TimeDivisor)),
		motion.ExprDiv(motion.ExprIndex(), motion.ExprConst(c.IndexDivisor)),
	)
	axis := func(base, amplitude, frequency float64) motion.FormulaExpr {
		wave := motion.ExprMul(motion.ExprConst(amplitude),
			motion.ExprSin(motion.ExprMul(motion.ExprConst(frequency), phase)))
		if c.Snap {
			wave = motion.ExprFloor(motion.ExprAdd(wave, motion.ExprConst(.5)))
		}
		return motion.ExprAdd(motion.ExprConst(base), wave)
	}
	return motion.NewFormulaFormation(motion.FormulaFormationConfig{
		X: axis(c.BaseX, c.AmplitudeX, c.FrequencyX),
		Y: axis(c.BaseY, c.AmplitudeY, c.FrequencyY),
	})
}

// CuddlySpreadpointBallFormation supplies the original twenty-slot scene
// clock. Count and artwork may be replaced without changing the path formula.
func CuddlySpreadpointBallFormation(image *ebiten.Image, count int) (sprites.GroupConfig, error) {
	formula, err := SpreadpointBallFormula(DefaultSpreadpointBallPath())
	if err != nil {
		return sprites.GroupConfig{}, err
	}
	return sprites.GroupConfig{
		Frames: []*ebiten.Image{image}, Count: count,
		Formula: formula, Phase: -1, PhaseStep: 1,
	}, nil
}
