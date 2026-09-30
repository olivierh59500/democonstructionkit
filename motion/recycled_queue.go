package motion

import "fmt"

type RecycledQueueConfig[T, P any] struct {
	Items                                  []T
	Count, Depth, Step, Spacing, DepthWrap int
	Grow                                   bool
	// Records are copied by value; any pointers inside T retain caller ownership.
	// Recycle changes the newly leading record; Project samples and may update
	// that record's animation phase. Both execute only during Step.
	Recycle func(tick int, item *T)
	Project func(index, depth int, item *T) (P, error)
}

// RecycledQueue owns bounded records, independent leading depth and cached
// poses. A strict spacing crossing rotates the last record into slot zero.
type RecycledQueue[T, P any] struct {
	config       RecycledQueueConfig[T, P]
	items        []T
	poses        []P
	count, depth int
}

func NewRecycledQueue[T, P any](c RecycledQueueConfig[T, P]) (*RecycledQueue[T, P], error) {
	if len(c.Items) < 1 || len(c.Items) > 65536 || c.Count < 0 || c.Count > len(c.Items) || c.Depth < 0 || c.Depth > c.Spacing || c.Step < 0 || c.Step > c.Spacing || c.Spacing < 1 || c.Spacing > 1<<20 || c.DepthWrap < c.Spacing || c.DepthWrap > 1<<30 || int64(len(c.Items))*int64(c.Spacing) > 1<<30 || c.Project == nil {
		return nil, fmt.Errorf("motion: invalid recycled queue")
	}
	q := &RecycledQueue[T, P]{config: c, items: append([]T(nil), c.Items...), poses: make([]P, len(c.Items)), count: c.Count, depth: c.Depth}
	q.config.Items = nil
	return q, nil
}

// Step advances records and cached poses once. A Project error stops the step
// without rolling back already advanced records.
func (q *RecycledQueue[T, P]) Step(tick int) error {
	if q == nil || tick < 0 {
		return fmt.Errorf("motion: invalid recycled queue tick")
	}
	q.depth += q.config.Step
	if q.depth > q.config.Spacing {
		q.depth -= q.config.Spacing
		if q.config.Grow {
			q.count = min(len(q.items), q.count+1)
		}
		last := q.items[len(q.items)-1]
		copy(q.items[1:], q.items[:len(q.items)-1])
		q.items[0] = last
		if q.config.Recycle != nil {
			q.config.Recycle(tick, &q.items[0])
		}
	}
	depth := q.depth
	for i := 0; i < q.count; i++ {
		if depth > q.config.DepthWrap {
			depth -= q.config.DepthWrap
		}
		p, err := q.config.Project(i, depth, &q.items[i])
		if err != nil {
			return err
		}
		q.poses[i] = p
		depth += q.config.Spacing
	}
	return nil
}

func (q *RecycledQueue[T, P]) Count() int {
	if q == nil {
		return 0
	}
	return q.count
}
func (q *RecycledQueue[T, P]) Depth() int {
	if q == nil {
		return 0
	}
	return q.depth
}

// Items borrows retained records until the next Step. This is a read-only
// inspection view; animation edits belong in Recycle or Project.
func (q *RecycledQueue[T, P]) Items() []T {
	if q == nil {
		return nil
	}
	return q.items
}

// Poses borrows the active cached poses until the next Step.
func (q *RecycledQueue[T, P]) Poses() []P {
	if q == nil {
		return nil
	}
	return q.poses[:q.count]
}
