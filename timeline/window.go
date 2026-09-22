package timeline

import (
	"fmt"
	"math"
)

// Window enables a layer or pass on [Start,Start+Duration). Zero Duration keeps
// it active indefinitely. Times and fades are in seconds; sampling is stateless.
type Window struct{ Start, Duration, FadeIn, FadeOut float64 }

func (w Window) Validate() error {
	for _, v := range []float64{w.Start, w.Duration, w.FadeIn, w.FadeOut} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("timeline: nonfinite window")
		}
	}
	if w.Duration < 0 || w.FadeIn < 0 || w.FadeOut < 0 || (w.Duration == 0 && w.FadeOut > 0) {
		return fmt.Errorf("timeline: invalid window duration or fade")
	}
	return nil
}
func (w Window) At(seconds float64) (local, alpha float64, active bool) {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, 0, false
	}
	local = seconds - w.Start
	if local < 0 || (w.Duration > 0 && local >= w.Duration) {
		return local, 0, false
	}
	alpha = 1
	if w.FadeIn > 0 {
		alpha = math.Min(alpha, local/w.FadeIn)
	}
	if w.Duration > 0 && w.FadeOut > 0 {
		alpha = math.Min(alpha, (w.Duration-local)/w.FadeOut)
	}
	return local, math.Max(0, alpha), true
}
