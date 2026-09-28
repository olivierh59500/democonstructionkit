package authoring

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestPostWindowRepeatsWithoutMovingItsFirstCue(t *testing.T) {
	clock := postClock{window: timeline.Window{Start: 1, Duration: 7, FadeIn: .5, FadeOut: .5}, period: 9}
	for _, check := range []struct {
		seconds, local, alpha float64
		active                bool
	}{
		{.5, -.5, 0, false},
		{1, 0, 0, true},
		{1.25, .25, .5, true},
		{1.5, .5, 1, true},
		{7.75, 6.75, .5, true},
		{8, 7, 0, false},
		{10, 0, 0, true},
		{10.25, .25, .5, true},
		{10.5, .5, 1, true},
		{16.75, 6.75, .5, true},
		{17, 7, 0, false},
	} {
		local, alpha, active := clock.At(check.seconds)
		if active != check.active || math.Abs(local-check.local) > 1e-12 || math.Abs(alpha-check.alpha) > 1e-12 {
			t.Fatalf("at %.2f: local %.12f alpha %.12f active %v; want %+v", check.seconds, local, alpha, active, check)
		}
	}
}
