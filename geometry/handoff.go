package geometry

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// Handoff matches a new deformation to the preceding displayed vertices, then
// removes the positional correction smoothly. It includes translation, rotation,
// scale and local deformation, without depending on Euler-angle conventions.
// The zero value applies no correction. Keep vertex correspondence stable.
type Handoff struct {
	previous, offsets []Vec3
	start, duration   float64
}

func (h *Handoff) Begin(previous, incoming []Vec3, seconds, duration float64) error {
	if len(previous) != len(incoming) || len(previous) == 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) || math.IsNaN(duration) || math.IsInf(duration, 0) || duration <= 0 {
		return fmt.Errorf("geometry: invalid handoff endpoints or duration")
	}
	if cap(h.previous) < len(previous) {
		h.previous = make([]Vec3, len(previous))
		h.offsets = make([]Vec3, len(previous))
	} else {
		h.previous = h.previous[:len(previous)]
		h.offsets = h.offsets[:len(previous)]
	}
	copy(h.previous, previous)
	for i := range previous {
		h.offsets[i] = previous[i].Sub(incoming[i])
	}
	h.start, h.duration = seconds, duration
	return nil
}

// Apply may operate in place and never advances time. At the handoff instant it
// copies the preceding vertices exactly; after duration it copies the incoming
// effect exactly. A second Begin may start from an already corrected pose.
func (h *Handoff) Apply(dst, current []Vec3, seconds float64) {
	n := min(len(dst), len(current))
	if len(h.offsets) != n || h.duration <= 0 || seconds >= h.start+h.duration {
		copy(dst, current)
		return
	}
	if seconds <= h.start {
		copy(dst, h.previous)
		return
	}
	weight := 1 - motion.Smooth((seconds-h.start)/h.duration)
	for i := 0; i < n; i++ {
		dst[i] = current[i].Add(h.offsets[i].Scale(weight))
	}
}
