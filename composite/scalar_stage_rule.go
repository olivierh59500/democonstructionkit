package composite

import (
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// ScalarStageRule gates a visual pass by stage, scalar value and direction.
// From and To are inclusive; DirectionSign 0 accepts either direction.
type ScalarStageRule struct {
	From, To      int
	Compare       timeline.ScalarCompare
	Threshold     float64
	DirectionSign int
}

func (rule ScalarStageRule) Matches(stage int, value, direction float64) bool {
	if stage < rule.From || stage > rule.To ||
		rule.DirectionSign > 0 && direction <= 0 ||
		rule.DirectionSign < 0 && direction >= 0 {
		return false
	}
	switch rule.Compare {
	case timeline.ScalarGreater:
		return value > rule.Threshold
	case timeline.ScalarGreaterEqual:
		return value >= rule.Threshold
	case timeline.ScalarLess:
		return value < rule.Threshold
	case timeline.ScalarLessEqual:
		return value <= rule.Threshold
	default:
		return true
	}
}
