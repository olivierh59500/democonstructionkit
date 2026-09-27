package presets

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// BilizirScrollLoop keeps the original four-pixel leftward speed and uses a
// relative message-width wrap. Two virtual copies can then meet without the
// previous viewport-width blank interval. The initial pass is unchanged.
func BilizirScrollLoop(messageWidth float64) (motion.WrapBankConfig, error) {
	if math.IsNaN(messageWidth) || math.IsInf(messageWidth, 0) || messageWidth <= 0 {
		return motion.WrapBankConfig{}, fmt.Errorf("presets: invalid Bilizir message width")
	}
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-4},
		Lower: &motion.WrapLimit{Boundary: -messageWidth, Restart: messageWidth, Relative: true},
	}, nil
}

// BilizirOriginalScrollLoop reproduces the preserved Go production's strict
// reset to the right edge after its complete message leaves the viewport.
// Use one virtual text copy with this clock; it intentionally leaves a blank
// interval while the new message travels in from outside the viewport.
func BilizirOriginalScrollLoop(messageWidth, viewportWidth float64) (motion.WrapBankConfig, error) {
	if math.IsNaN(messageWidth) || math.IsInf(messageWidth, 0) || messageWidth <= 0 ||
		math.IsNaN(viewportWidth) || math.IsInf(viewportWidth, 0) || viewportWidth <= 0 {
		return motion.WrapBankConfig{}, fmt.Errorf("presets: invalid Bilizir original scroll width")
	}
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-4},
		Lower: &motion.WrapLimit{Boundary: -messageWidth, Restart: viewportWidth},
	}, nil
}
