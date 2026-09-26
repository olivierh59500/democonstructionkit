package geometry

import (
	"fmt"
	"math"
)

// PointReader exposes model-space coordinates without requiring a particular
// sprite, mesh or ball-point storage layout.
type PointReader interface {
	Len() int
	XYZ(index int) Vec3
}

// PointWriter changes only model coordinates; artwork and palette indices in
// the caller's point collection remain untouched.
type PointWriter interface {
	PointReader
	SetXYZ(index int, position Vec3)
}

type MorphTail uint8

const (
	// MorphHold keeps source points that have no matching target in place.
	MorphHold MorphTail = iota
	// MorphRepeat cycles a shorter target over all source points.
	MorphRepeat
)

// PointMorphConfig chooses the number of fixed simulation steps and the policy
// for unequal point counts. SnapFinal is optional; leaving it false preserves
// incremental rounding in productions that accumulate a fixed per-point delta.
type PointMorphConfig struct {
	Frames    int
	Tail      MorphTail
	SnapFinal bool
}

// PointMorph starts from the current source pose, so a newly selected target
// does not teleport the shape at a timeline handoff. It reuses one delta bank.
type PointMorph struct {
	config  PointMorphConfig
	deltas  []Vec3
	targets []Vec3
	frame   int
}

func NewPointMorph(from, to PointReader, config PointMorphConfig) (*PointMorph, error) {
	if from == nil || to == nil || from.Len() < 0 || from.Len() > 1_000_000 ||
		to.Len() < 0 || to.Len() > 1_000_000 || config.Frames < 1 || config.Frames > 1_000_000 ||
		config.Tail > MorphRepeat {
		return nil, fmt.Errorf("motion: invalid point morph configuration")
	}
	morph := &PointMorph{config: config, deltas: make([]Vec3, from.Len())}
	if config.SnapFinal {
		morph.targets = make([]Vec3, from.Len())
	}
	targetCount := to.Len()
	finite := func(p Vec3) bool {
		return !math.IsNaN(p.X) && !math.IsInf(p.X, 0) && !math.IsNaN(p.Y) && !math.IsInf(p.Y, 0) &&
			!math.IsNaN(p.Z) && !math.IsInf(p.Z, 0)
	}
	for i := range morph.deltas {
		start := from.XYZ(i)
		if !finite(start) {
			return nil, fmt.Errorf("motion: nonfinite point morph source")
		}
		if i >= targetCount && (config.Tail == MorphHold || targetCount == 0) {
			if config.SnapFinal {
				morph.targets[i] = start
			}
			continue
		}
		index := i
		if config.Tail == MorphRepeat {
			index %= targetCount
		}
		target := to.XYZ(index)
		if !finite(target) {
			return nil, fmt.Errorf("motion: nonfinite point morph target")
		}
		morph.deltas[i] = Vec3{
			X: (target.X - start.X) / float64(config.Frames),
			Y: (target.Y - start.Y) / float64(config.Frames),
			Z: (target.Z - start.Z) / float64(config.Frames),
		}
		if !finite(morph.deltas[i]) {
			return nil, fmt.Errorf("motion: point morph delta overflows")
		}
		if config.SnapFinal {
			morph.targets[i] = target
		}
	}
	return morph, nil
}

// Step advances the supplied point collection exactly once. Calling it after
// completion leaves the final pose unchanged; drawing never advances the morph.
func (m *PointMorph) Step(points PointWriter) {
	if m == nil || points == nil || m.frame >= m.config.Frames {
		return
	}
	count := min(points.Len(), len(m.deltas))
	for i := 0; i < count; i++ {
		p, delta := points.XYZ(i), m.deltas[i]
		p.X += delta.X
		p.Y += delta.Y
		p.Z += delta.Z
		if m.config.SnapFinal && m.frame+1 == m.config.Frames {
			p = m.targets[i]
		}
		points.SetXYZ(i, p)
	}
	m.frame++
}

// Frame and Finished expose the current cue progress without changing points.
func (m *PointMorph) Frame() int     { return m.frame }
func (m *PointMorph) Finished() bool { return m.frame >= m.config.Frames }
