package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// PhenomenaMaterialFormulas preserves the authored arithmetic order of its
// brightness, white-mask alpha and red-photon channel ramps.
type PhenomenaMaterialFormulas struct {
	Value, Percent, BrightDown, WhiteMask, Green, Secondary motion.FormulaExpr
}

func PhenomenaStageColorFormulas() PhenomenaMaterialFormulas {
	value := motion.ExprTime()
	percent := motion.ExprDiv(value, motion.ExprConst(100))
	return PhenomenaMaterialFormulas{
		Value: value, Percent: percent,
		BrightDown: motion.ExprSub(motion.ExprConst(2), percent),
		WhiteMask:  motion.ExprDiv(motion.ExprSub(motion.ExprConst(200), value), motion.ExprConst(100)),
		Green:      motion.ExprMul(percent, motion.ExprConst(.5)),
		Secondary:  motion.ExprSecondaryTime(),
	}
}
