package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyColorshockOrbit is a serializable, editable two-frequency backdrop
// trajectory. Expression order retains the source's multiply-then-divide phase.
func CuddlyColorshockOrbit() motion.FormulaFormationConfig {
	c, time := motion.ExprConst, motion.ExprTime()
	xPhase := motion.ExprDiv(motion.ExprMul(time, c(math.Pi)), c(100))
	yPhase := motion.ExprDiv(motion.ExprMul(time, c(math.Pi)), c(200))
	return motion.FormulaFormationConfig{
		X: motion.ExprSub(c(400), motion.ExprMul(motion.ExprSin(xPhase), c(400))),
		Y: motion.ExprSub(c(180), motion.ExprMul(motion.ExprCos(yPhase), c(300))),
	}
}

// CuddlyColorshockTableClock indexes the supplied position table two entries
// per frame and resets only after passing index 1872.
func CuddlyColorshockTableClock() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{2},
		Upper: &motion.WrapLimit{Boundary: 1872, Restart: 0},
	}
}
