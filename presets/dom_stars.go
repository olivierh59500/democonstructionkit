package presets

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// DOMStarOptions keeps the sprite count, grid, lifetime, frame speed and
// placement editable. XMaxCell is the inclusive maximum grid index. Random is
// called in sprite-index order, four times for each initial sprite and three
// times for each respawn.
type DOMStarOptions struct {
	Count, XMaxCell            int
	XStep, YMax                float64
	RateDenominator, RateRange float64
	InitialFrameMax, EndPhase  float64
	OffsetX, OffsetY           float64
	Random                     func() float64
}

// DefaultDOMStarOptions reproduces the authored eight-sprite animation.
func DefaultDOMStarOptions(random func() float64) DOMStarOptions {
	return DOMStarOptions{Count: 8, XMaxCell: 9, XStep: 64, YMax: 354,
		RateDenominator: 4, RateRange: 4, InitialFrameMax: 10, EndPhase: 9,
		OffsetX: 64, OffsetY: 60, Random: random}
}

// DOMAnimatedStars binds the editable spawn recipe to a shared animated field.
// Frames are borrowed; they may come from any sprite atlas with compatible art.
func DOMAnimatedStars(frames []*ebiten.Image, c DOMStarOptions) (sprites.AnimatedFieldConfig, error) {
	if c.Count < 1 || c.XMaxCell < 0 || c.Random == nil ||
		!finiteDOMStar(c.XStep) || !finiteDOMStar(c.YMax) || c.YMax < 0 ||
		!finiteDOMStar(c.RateDenominator) || c.RateDenominator <= 0 ||
		!finiteDOMStar(c.RateRange) || c.RateRange < 0 ||
		!finiteDOMStar(c.InitialFrameMax) || c.InitialFrameMax < 0 ||
		!finiteDOMStar(c.EndPhase) || c.EndPhase <= 0 ||
		!finiteDOMStar(c.OffsetX) || !finiteDOMStar(c.OffsetY) {
		return sprites.AnimatedFieldConfig{}, fmt.Errorf("presets: invalid DOM star recipe")
	}
	spawn := func(_ int, reset bool) motion.FrameParticle {
		x := math.Round(c.Random()*float64(c.XMaxCell)) * c.XStep
		y := math.Round(c.Random() * c.YMax)
		denominator := math.Round(c.Random()*c.RateRange) + c.RateDenominator
		phase := 0.0
		if !reset {
			phase = math.Round(c.Random() * c.InitialFrameMax)
		}
		return motion.FrameParticle{X: x, Y: y, Phase: phase, Rate: 1 / denominator}
	}
	return sprites.AnimatedFieldConfig{
		Motion: motion.FrameFieldConfig{Count: c.Count, EndPhase: c.EndPhase, Spawn: spawn},
		Frames: frames, OffsetX: c.OffsetX, OffsetY: c.OffsetY,
	}, nil
}

func finiteDOMStar(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
