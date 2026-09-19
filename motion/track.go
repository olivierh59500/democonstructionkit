package motion

import (
	"fmt"
	"math"
	"sort"
)

// Key is an absolute-time value. Ease controls the segment starting at this key.
type Key[T any] struct {
	Time  float64
	Value T
	Ease  Ease
}

// Track interpolates immutable ordered keyframes, including camera paths.
type Track[T any] struct {
	keys        []Key[T]
	interpolate func(T, T, float64) T
}

func NewTrack[T any](keys []Key[T], interpolate func(T, T, float64) T) (*Track[T], error) {
	if len(keys) == 0 || interpolate == nil {
		return nil, fmt.Errorf("motion: empty track or missing interpolation")
	}
	for i, k := range keys {
		if math.IsNaN(k.Time) || math.IsInf(k.Time, 0) || (i > 0 && keys[i-1].Time >= k.Time) {
			return nil, fmt.Errorf("motion: key times must be finite and strictly increasing")
		}
	}
	return &Track[T]{keys: append([]Key[T](nil), keys...), interpolate: interpolate}, nil
}
func (t *Track[T]) At(seconds float64) T {
	if seconds <= t.keys[0].Time || math.IsNaN(seconds) {
		return t.keys[0].Value
	}
	i := sort.Search(len(t.keys), func(i int) bool { return t.keys[i].Time > seconds })
	if i == len(t.keys) {
		return t.keys[i-1].Value
	}
	a, b := t.keys[i-1], t.keys[i]
	ease := a.Ease
	if ease == nil {
		ease = Linear
	}
	return t.interpolate(a.Value, b.Value, ease((seconds-a.Time)/(b.Time-a.Time)))
}
