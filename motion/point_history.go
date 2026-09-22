package motion

import "fmt"

// PointHistory stores a fixed number of recent positions for delayed sprites,
// logos or pointer trails. Each Push is one caller-defined animation tick; Draw
// must not push. Sampling never changes history and does not allocate.
type PointHistory[T any] struct {
	points []T
	head   int
}

// NewPointHistory fills every delay with initial, avoiding an empty trail during
// startup. Capacity includes the current point: a delay of 60 needs 61 entries.
func NewPointHistory[T any](capacity int, initial T) (*PointHistory[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("motion: point history capacity must be positive")
	}
	h := &PointHistory[T]{points: make([]T, capacity)}
	h.Reset(initial)
	return h, nil
}

// Reset fills all delays with initial and starts a new history without allocating.
func (h *PointHistory[T]) Reset(initial T) {
	for i := range h.points {
		h.points[i] = initial
	}
	h.head = 0
}

// Push records the current point in constant time, discarding the oldest one.
func (h *PointHistory[T]) Push(point T) {
	if len(h.points) == 0 {
		return
	}
	h.head++
	if h.head == len(h.points) {
		h.head = 0
	}
	h.points[h.head] = point
}

// At returns the point delay pushes before the most recent point. Delays outside
// [0, Capacity()) are invalid rather than wrapping into unrelated recent data.
func (h *PointHistory[T]) At(delay int) (T, bool) {
	if delay < 0 || delay >= len(h.points) {
		var zero T
		return zero, false
	}
	i := h.head - delay
	if i < 0 {
		i += len(h.points)
	}
	return h.points[i], true
}

func (h *PointHistory[T]) Capacity() int { return len(h.points) }
