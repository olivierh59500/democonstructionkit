package sprites

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// DepthQueueUpperPolicy selects how the leading depth crosses its upper bound.
type DepthQueueUpperPolicy uint8

const (
	QueueUpperStrict DepthQueueUpperPolicy = iota
	// QueueUpperInclusiveFirst includes equality on entry, then makes only
	// strict corrections. At Far it shifts once; at Far+Spacing it also shifts
	// once. This retains the phase of queues with distinct outer/inner tests.
	QueueUpperInclusiveFirst
	QueueUpperInclusive
)

// DepthQueueConfig describes an ordered ring of independently skinned points.
// Slot zero uses the leading Depth; each following slot subtracts Spacing. Head
// changes when the leading depth crosses Near/Far. Project can choose integer
// projection, visibility, atlas frames and scale. Without Project, Camera gives
// an ordinary perspective projection. Points and frame rectangles are copied;
// artwork remains borrowed. Construction does not advance or sample the queue.
type DepthQueueConfig struct {
	Points             []geometry.Vec2
	Head, Depth        int
	Near, Far, Spacing int
	DepthStep          int
	LowerInclusive     bool
	UpperPolicy        DepthQueueUpperPolicy
	Project            func(slot int, point geometry.Vec2, depth int) (FieldSample, bool)
	Camera             geometry.Camera
	Style              FieldStyle
	RendererCapacity   int
}

type DepthQueuePose struct {
	FieldSample
	PointIndex int
	Visible    bool
}

// DepthQueue owns the point ring, cached poses and visible draw population.
// Boundary correction uses bounded integer arithmetic rather than repeated
// subtraction. Sampling and drawing reuse storage; the shared FieldRenderer is
// created on first Draw so CPU clocks do not require a graphics resource.
type DepthQueue struct {
	Style                  FieldStyle
	config                 DepthQueueConfig
	points                 []geometry.Vec2
	head, depth, step      int
	poses, scratch         []DepthQueuePose
	samples, sampleScratch []FieldSample
	renderer               *FieldRenderer
	closed                 bool
}

func NewDepthQueue(c DepthQueueConfig) (*DepthQueue, error) {
	const limit = 1 << 30
	count := len(c.Points)
	if count < 1 || count > 65536 || c.Near < -limit || c.Far > limit || c.Far <= c.Near ||
		c.Spacing < 1 || c.Spacing > limit || int64(c.Far)-int64(c.Near) < int64(c.Spacing) ||
		int64(c.Near)-int64(count-1)*int64(c.Spacing) < -limit ||
		c.Depth < -limit || c.Depth > limit || c.DepthStep < -limit || c.DepthStep > limit ||
		c.UpperPolicy > QueueUpperInclusive || c.RendererCapacity < 0 || c.RendererCapacity > 65536 {
		return nil, fmt.Errorf("sprites: invalid depth queue bounds, population or transport")
	}
	for _, point := range c.Points {
		if !finiteField(point.X) || !finiteField(point.Y) {
			return nil, fmt.Errorf("sprites: nonfinite depth queue point")
		}
	}
	if c.Project == nil && (c.Camera.Focal <= 0 || c.Camera.Near <= 0 || !finiteField(c.Camera.Focal) || !finiteField(c.Camera.Near) ||
		!finiteField(c.Camera.Center.X) || !finiteField(c.Camera.Center.Y)) {
		return nil, fmt.Errorf("sprites: depth queue needs a projection callback or valid camera")
	}
	c.Style.Frames = append([]image.Rectangle(nil), c.Style.Frames...)
	points := append([]geometry.Vec2(nil), c.Points...)
	c.Points = nil
	head := c.Head % count
	if head < 0 {
		head += count
	}
	return &DepthQueue{Style: c.Style, config: c, points: points, head: head, depth: c.Depth, step: c.DepthStep,
		poses: make([]DepthQueuePose, count), scratch: make([]DepthQueuePose, count),
		samples: make([]FieldSample, 0, count), sampleScratch: make([]FieldSample, 0, count)}, nil
}

func depthQueueCeil(value, divisor int64) int64 { return (value + divisor - 1) / divisor }

func (q *DepthQueue) normalize(depth int64) (int, int) {
	near, far, spacing := int64(q.config.Near), int64(q.config.Far), int64(q.config.Spacing)
	head := int64(q.head)
	if depth < near || q.config.LowerInclusive && depth == near {
		shifts := depthQueueCeil(near-depth, spacing)
		if q.config.LowerInclusive {
			shifts = (near-depth)/spacing + 1
		}
		depth += shifts * spacing
		head -= shifts
	}
	if depth > far || q.config.UpperPolicy != QueueUpperStrict && depth == far {
		shifts := depthQueueCeil(depth-far, spacing)
		switch q.config.UpperPolicy {
		case QueueUpperInclusiveFirst:
			shifts = max(1, shifts)
		case QueueUpperInclusive:
			shifts = (depth-far)/spacing + 1
		}
		depth -= shifts * spacing
		head += shifts
	}
	count := int64(len(q.points))
	head %= count
	if head < 0 {
		head += count
	}
	return int(head), int(depth)
}

// Step applies an explicit leading-depth change, then prepares every slot once.
// Invalid input or a nonfinite visible pose keeps the previous transport and
// prepared frame. Caller callbacks may have side effects of their own.
func (q *DepthQueue) Step(delta int) error {
	if q == nil || q.closed || delta < -(1<<30) || delta > 1<<30 {
		return fmt.Errorf("sprites: invalid or closed depth queue step")
	}
	head, depth := q.normalize(int64(q.depth) + int64(delta))
	visible := q.sampleScratch[:0]
	for slot := range q.points {
		pointIndex := (head + slot) % len(q.points)
		pointDepth := depth - slot*q.config.Spacing
		point := q.points[pointIndex]
		var sample FieldSample
		var shown bool
		if q.config.Project != nil {
			sample, shown = q.config.Project(slot, point, pointDepth)
		} else {
			position, scale, ok := q.config.Camera.Project(geometry.Vec3{X: point.X, Y: point.Y, Z: float64(pointDepth)})
			sample, shown = FieldSample{X: position.X, Y: position.Y, Z: float64(pointDepth), Scale: scale}, ok
		}
		sample.Index = pointIndex
		if shown && (!finiteField(sample.X) || !finiteField(sample.Y) || !finiteField(sample.Z) || !finiteField(sample.Scale) || !finiteField(sample.PreviousX) || !finiteField(sample.PreviousY)) {
			return fmt.Errorf("sprites: nonfinite visible depth queue pose at slot %d", slot)
		}
		q.scratch[slot] = DepthQueuePose{FieldSample: sample, PointIndex: pointIndex, Visible: shown}
		if shown {
			visible = append(visible, sample)
		}
	}
	q.head, q.depth = head, depth
	q.poses, q.scratch = q.scratch, q.poses
	q.samples, q.sampleScratch = visible, q.samples
	return nil
}

func (q *DepthQueue) Update(kit.Frame) error {
	if q == nil {
		return fmt.Errorf("sprites: absent depth queue")
	}
	return q.Step(q.step)
}

// SetDepthStep changes the Update transport without resetting the ring or pose.
func (q *DepthQueue) SetDepthStep(step int) error {
	if q == nil || q.closed || step < -(1<<30) || step > 1<<30 {
		return fmt.Errorf("sprites: invalid depth queue velocity")
	}
	q.step = step
	return nil
}

// Reset restores the configured head, depth and update velocity without
// resampling. It retains CPU storage and any already-created renderer.
func (q *DepthQueue) Reset() error {
	if q == nil || q.closed {
		return fmt.Errorf("sprites: closed depth queue reset")
	}
	q.head = q.config.Head % len(q.points)
	if q.head < 0 {
		q.head += len(q.points)
	}
	q.depth, q.step = q.config.Depth, q.config.DepthStep
	clear(q.poses)
	clear(q.scratch)
	q.samples, q.sampleScratch = q.samples[:0], q.sampleScratch[:0]
	return nil
}

func (q *DepthQueue) Head() int  { return q.head }
func (q *DepthQueue) Depth() int { return q.depth }

// Poses and Samples borrow read-only prepared storage until the next Step or Reset.
func (q *DepthQueue) Poses() []DepthQueuePose { return q.poses }
func (q *DepthQueue) Samples() []FieldSample  { return q.samples }

func (q *DepthQueue) Draw(dst *ebiten.Image) {
	if q != nil {
		q.DrawStyle(dst, q.Style)
	}
}

// DrawStyle reuses the sampled population with another sprite, pixel, trail or
// outline material. Atlas frames may have different dimensions per depth.
func (q *DepthQueue) DrawStyle(dst *ebiten.Image, style FieldStyle) {
	if q == nil || q.closed || dst == nil || len(q.samples) == 0 {
		return
	}
	if q.renderer == nil {
		capacity := q.config.RendererCapacity
		if capacity == 0 {
			capacity = len(q.points)
		}
		q.renderer = NewFieldRenderer(capacity)
	}
	q.renderer.Draw(dst, q.samples, style)
}

func (q *DepthQueue) Close() error {
	if q == nil || q.closed {
		return nil
	}
	q.closed = true
	if q.renderer != nil {
		q.renderer.Close()
	}
	q.renderer, q.points, q.poses, q.scratch = nil, nil, nil, nil
	q.samples, q.sampleScratch = nil, nil
	q.config.Project, q.config.Points, q.config.Style = nil, nil, FieldStyle{}
	q.Style = FieldStyle{}
	return nil
}

var _ kit.Effect = (*DepthQueue)(nil)
