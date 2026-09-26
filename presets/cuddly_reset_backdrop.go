package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyResetBackdropTrajectory describes the Reset showcase's coupled
// horizontal waves and vertical orbit. The second phase wraps independently.
func CuddlyResetBackdropTrajectory() motion.FormulaTrajectoryConfig {
	angle := motion.ExprMul(motion.ExprConst(2*math.Pi/384), motion.ExprTime())
	x := motion.ExprSub(motion.ExprConst(335), motion.ExprMul(
		motion.ExprMul(motion.ExprConst(220), motion.ExprSin(angle)),
		motion.ExprSin(motion.ExprSecondaryTime()),
	))
	y := motion.ExprAdd(motion.ExprConst(250), motion.ExprMul(
		motion.ExprConst(170), motion.ExprSin(motion.ExprAdd(angle, motion.ExprConst(math.Pi/3))),
	))
	return motion.FormulaTrajectoryConfig{
		X: x, Y: y,
		Start:  [2]float64{12, 0},
		Step:   [2]float64{1, math.Pi / 32},
		Period: [2]float64{0, 2 * math.Pi},
	}
}
