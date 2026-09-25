package presets

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyMenuFormulas describes seven editable, composable sprite trajectories.
// Formulas retain their authored arithmetic order while width, height, count,
// time, and item index remain live inputs rather than fixed screenshot pixels.
func CuddlyMenuFormulas() []motion.FormulaFormationConfig {
	c, t, i := motion.ExprConst, motion.ExprTime(), motion.ExprIndex()
	w, h, count := motion.ExprWidth(), motion.ExprHeight(), motion.ExprCount()
	add, sub, mul, div := motion.ExprAdd, motion.ExprSub, motion.ExprMul, motion.ExprDiv
	sin, cos := motion.ExprSin, motion.ExprCos
	lag := func(spacing float64) motion.FormulaExpr { return sub(t, mul(i, c(spacing))) }
	twist := func(rate, gain float64) motion.FormulaExpr {
		phase := mul(t, c(rate))
		return mul(add(sin(phase), cos(phase)), c(gain))
	}

	// A centered row and one rotating horizontal harmonic.
	w0 := mul(mul(w, c(.5)), c(.5))
	h0 := mul(mul(h, c(.25)), c(.25))
	linear0 := mul(mul(sub(i, div(sub(count, c(1)), c(2))), c(.3)), w0)
	spin0 := sin(add(mul(t, c(6.1)), mul(mul(i, c(6.1)), c(3))))
	mode0 := motion.FormulaFormationConfig{
		X: mul(add(linear0, mul(spin0, c(25))), c(2.1)),
		Y: mul(mul(cos(mul(lag(.15), c(4.5))), h0), c(3)),
	}

	// Independent X/Y waves with one shared breathing envelope.
	mode1 := motion.FormulaFormationConfig{
		X: mul(mul(sin(mul(lag(.1), c(2.5*.7))), w), c(1.05)),
		Y: mul(mul(sin(mul(lag(.1), c(5.0*.7))), h), twist(1.6*.7, .75)),
	}

	// The horizontal axis folds through a second per-item cosine.
	x2 := mul(mul(sin(mul(lag(.07), c(4.0))), w), c(1.05))
	mode2 := motion.FormulaFormationConfig{
		X: mul(x2, cos(mul(sub(t, i), c(.25)))),
		Y: mul(mul(sin(mul(lag(.07), c(3.0))), h), twist(2.0, .75)),
	}

	// The shared envelope acts on X while a per-item wave folds Y.
	y3 := mul(cos(mul(lag(.07), c(3.0*.7))), h)
	mode3 := motion.FormulaFormationConfig{
		X: mul(mul(sin(mul(lag(.07), c(4.0*.7))), w), twist(2.0, .73)),
		Y: mul(y3, sin(mul(sub(t, i), c(.25)))),
	}

	// Each axis has an independent one-minus-sine aperture.
	xTwist4 := mul(sub(c(1), sin(mul(lag(.1), c(4.5)))), c(.52))
	yTwist4 := mul(sub(c(1), sin(mul(lag(.1), c(4.0)))), c(.52))
	mode4 := motion.FormulaFormationConfig{
		X: mul(mul(sin(mul(lag(.1), c(1.0))), w), xTwist4),
		Y: mul(mul(cos(mul(lag(.1), c(1.2))), h), yTwist4),
	}

	// Two equal-speed waves with a reduced vertical radius.
	mode5 := motion.FormulaFormationConfig{
		X: mul(mul(cos(mul(lag(.1), c(3.5*.7))), w), c(1.05)),
		Y: mul(mul(sin(mul(lag(.1), c(3.5*.7))), mul(h, c(.9))), twist(3.0*.5, .82)),
	}

	// A final X multiplier folds the cosine while Y keeps the shared envelope.
	x6 := mul(mul(cos(mul(lag(.1), c(4.0*.7))), w), c(1.04))
	mode6 := motion.FormulaFormationConfig{
		X: mul(x6, sin(mul(lag(.1), c(.5)))),
		Y: mul(mul(sin(mul(lag(.1), c(3.0*.7))), h), twist(3.0, .75)),
	}
	return []motion.FormulaFormationConfig{mode0, mode1, mode2, mode3, mode4, mode5, mode6}
}

// CuddlyMenuCarousel is an editable viewport-relative carousel preset. Atlas
// artwork, image order, hold/slide times and anchors may all be replaced.
func CuddlyMenuCarousel(atlas *sprites.Atlas) sprites.FormationCarouselConfig {
	return sprites.FormationCarouselConfig{
		Atlas: atlas, Count: 12, Modes: CuddlyMenuFormulas(),
		Hold: 8, Slide: 1, SlideIndex: .5,
		RadiusX: .9, RadiusY: .88,
		AnchorX: float64(atlas.TileW) / 2, AnchorY: float64(atlas.TileH) / 2,
		PixelSnap: true,
	}
}
