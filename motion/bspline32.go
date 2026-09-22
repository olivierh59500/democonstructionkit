package motion

import (
	"fmt"
	"math"
	"sort"
)

// BSpline32 is a uniform cubic B-spline evaluated with float32 arithmetic. Keys
// are interleaved time and component values: [time, x, y, z, time, x, y, z, ...]
// for three components. At least four keys are required. Repeated interior key
// times are supported and retain their authored order. Unlike an interpolating
// Catmull-Rom path, the curve generally does not pass through its control points.
//
// A segment uses four consecutive keys and the times of the first two. The final
// three keys are control support, not additional segments. Outside the authored
// time range the first/last segment extrapolates; no implicit clamping occurs.
// This convention preserves imported camera tracks and supports negative starts.
type BSpline32 struct {
	keys       []float32
	components int
	count      int
}

func NewBSpline32(keys []float32, components int) (*BSpline32, error) {
	if components < 1 || components > 64 || len(keys)%(components+1) != 0 || len(keys)/(components+1) < 4 {
		return nil, fmt.Errorf("motion: invalid B-spline key layout")
	}
	stride := components + 1
	for i, value := range keys {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || (i%stride == 0 && i > 0 && value < keys[i-stride]) {
			return nil, fmt.Errorf("motion: B-spline values must be finite and key times nondecreasing")
		}
	}
	count := len(keys) / stride
	last := (count - 4) * stride
	if keys[stride] == keys[0] || keys[last+stride] == keys[last] {
		return nil, fmt.Errorf("motion: B-spline extrapolation segments need distinct key times")
	}
	return &BSpline32{keys: append([]float32(nil), keys...), components: components, count: len(keys) / stride}, nil
}

// Sample writes Components values without allocation. It returns false for a
// short output or non-finite time and leaves output unchanged on that failure.
// Exact key times belong to the preceding segment, preserving knot boundaries.
func (s *BSpline32) Sample(out []float32, at float32) bool {
	if s == nil || s.count < 4 || s.components < 1 || len(out) < s.components || math.IsNaN(float64(at)) || math.IsInf(float64(at), 0) {
		return false
	}
	stride := s.components + 1
	segment := sort.Search(s.count-1, func(i int) bool { return s.keys[(i+1)*stride] >= at })
	if segment > s.count-4 {
		segment = s.count - 4
	}
	base := segment * stride
	t := (at - s.keys[base]) / (s.keys[base+stride] - s.keys[base])
	t2 := t * t
	t3 := t2 * t
	w0 := ((-1.0 / 6.0) * t3) + (0.5 * t2) - (0.5 * t) + (1.0 / 6.0)
	w1 := (0.5 * t3) - t2 + (2.0 / 3.0)
	w2 := ((-0.5) * t3) + (0.5 * t2) + (0.5 * t) + (1.0 / 6.0)
	w3 := (1.0 / 6.0) * t3
	for i := 0; i < s.components; i++ {
		out[i] = s.keys[base+i+1]*w0 +
			s.keys[base+i+1+stride]*w1 +
			s.keys[base+i+1+stride*2]*w2 +
			s.keys[base+i+1+stride*3]*w3
	}
	return true
}

func (s *BSpline32) Components() int { return s.components }
