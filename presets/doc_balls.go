package presets

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/effects"
)

// DOCBallPrograms returns eight independent movement functions. The sequence
// is editable: replace any function, reorder it or use another selector.
func DOCBallPrograms() []effects.BallMotionFunc {
	return []effects.BallMotionFunc{
		func(float64, int) effects.BallMotion { return effects.BallMotion{SpinSpeed: -5, Height: 40} },
		func(float64, int) effects.BallMotion { return effects.BallMotion{SpinSpeed: -5, Height: 40} },
		func(t float64, _ int) effects.BallMotion {
			return effects.BallMotion{SpinSpeed: -5, Height: -60 - math.Sin(t*7)*95, AngleSpacing: 35, Radius: 150}
		},
		func(t float64, i int) effects.BallMotion {
			return effects.BallMotion{SpinSpeed: 5, Height: math.Sin((t+float64(i))*0.5*13)*90 - 50, AngleSpacing: 16, Radius: 150}
		},
		func(t float64, i int) effects.BallMotion {
			q := (t + float64(i)) * 0.125 * 13.5
			return effects.BallMotion{SpinSpeed: 5, Height: 80 - math.Abs(math.Sin(q)*8*math.Cos(q)*42) - 50, AngleSpacing: 20, Radius: 150}
		},
		func(t float64, i int) effects.BallMotion {
			q := (t + float64(i)) * 0.25 * 13.5
			return effects.BallMotion{SpinSpeed: 5, Height: math.Sin(q)*8*math.Cos(q)*22 - 50, AngleSpacing: 20, Radius: 150}
		},
		func(t float64, i int) effects.BallMotion {
			q := (t + float64(i)) * 0.25 * 13.5
			return effects.BallMotion{SpinSpeed: -7, Height: math.Sin(q)*8*math.Cos(q)*22 - 50, AngleSpacing: 20, Radius: 150}
		},
		func(t float64, i int) effects.BallMotion {
			return effects.BallMotion{SpinSpeed: -8, Height: 10 - math.Abs(math.Sin((t*0.6+float64(i)*0.05)*1.75)*70)*2.3, AngleSpacing: 20, Radius: 150}
		},
	}
}

// DOCBallSequence preserves the standalone scene's opening replacement and
// later loop through movement programs two to seven.
func DOCBallSequence(t float64, _ int) (from, to int, alpha float64) {
	index := int(t/7) % 8
	if index < 2 {
		if t <= 21 {
			index = 7
		} else {
			index = 2 + int(t/7)%6
		}
	}
	from, to = index, index+1
	if to >= 8 {
		to = 2 + (to-2)%6
	}
	alpha = math.Min(1, math.Mod(t/7, 1)*7*.8)
	return
}

// CuddlyDOCBallSequence uses seven repeating segment slots, with the two
// entrance slots held at program seven after the opening.
func CuddlyDOCBallSequence(t float64, _ int) (from, to int, alpha float64) {
	index := int(math.Floor(math.Mod(t/7, 7)))
	resolve := func(i int) int {
		if i < 2 && t > 21 {
			return 7
		}
		return min(i, 7)
	}
	from, to = resolve(index), resolve(index+1)
	alpha = math.Min(1, math.Mod(t/7, 1)*7*1.3)
	return
}

// DOCProjectedBalls builds the standalone shadowed train from caller images.
func DOCProjectedBalls(ball *ebiten.Image, shadows []*ebiten.Image) effects.ProjectedBallTrainConfig {
	c := effects.DefaultProjectedBallTrainConfig()
	c.Ball, c.Shadows = ball, shadows
	c.Programs, c.Sequence = DOCBallPrograms(), DOCBallSequence
	return c
}

// CuddlyDOCProjectedBalls changes only the source-specific phase, sequence,
// rotation arithmetic and shadow painter order.
func CuddlyDOCProjectedBalls(ball *ebiten.Image, shadows []*ebiten.Image) effects.ProjectedBallTrainConfig {
	c := DOCProjectedBalls(ball, shadows)
	c.SpinStep = .2
	c.Sequence = CuddlyDOCBallSequence
	c.Rotation = effects.BallRotationDirect
	c.ShadowOrder = effects.BallSlotOrder
	return c
}
