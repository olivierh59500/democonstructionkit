package motion

import (
	"fmt"
	"math"
	"sort"
)

// FormationFrame gives every item an absolute position at one time. Ease
// controls interpolation to the next frame; nil selects linear interpolation.
type FormationFrame struct {
	Time   float64
	Points []Point
	Ease   Ease
}

// KeyframedFormationConfig describes an arbitrary, editable path for every
// sprite. Loop is measured in seconds; the first and final point banks must
// match when it is positive.
type KeyframedFormationConfig struct {
	Count  int
	Loop   float64
	Frames []FormationFrame
}

// KeyframedFormation owns a bounded copy of its position banks. Sampling an
// item's pose needs no allocation or GPU resource and can serve several views.
type KeyframedFormation struct {
	count  int
	loop   float64
	frames []FormationFrame
}

func NewKeyframedFormation(config KeyframedFormationConfig) (*KeyframedFormation, error) {
	if config.Count < 1 || config.Count > 10000 || len(config.Frames) < 1 ||
		len(config.Frames) > 4096 || config.Count > 1_048_576/len(config.Frames) ||
		math.IsNaN(config.Loop) || math.IsInf(config.Loop, 0) || config.Loop < 0 {
		return nil, fmt.Errorf("motion: invalid keyframed formation size or loop")
	}
	frames := make([]FormationFrame, len(config.Frames))
	for i, frame := range config.Frames {
		if math.IsNaN(frame.Time) || math.IsInf(frame.Time, 0) || frame.Time < 0 ||
			i > 0 && frame.Time <= config.Frames[i-1].Time || len(frame.Points) != config.Count {
			return nil, fmt.Errorf("motion: invalid formation frame %d", i)
		}
		for _, point := range frame.Points {
			if math.IsNaN(point.X) || math.IsInf(point.X, 0) ||
				math.IsNaN(point.Y) || math.IsInf(point.Y, 0) {
				return nil, fmt.Errorf("motion: nonfinite formation point in frame %d", i)
			}
		}
		frames[i] = FormationFrame{
			Time: frame.Time, Points: append([]Point(nil), frame.Points...), Ease: frame.Ease,
		}
	}
	if config.Loop > 0 {
		if len(frames) < 2 || frames[0].Time != 0 || frames[len(frames)-1].Time != config.Loop {
			return nil, fmt.Errorf("motion: formation loop needs keys at zero and its end")
		}
		for i, first := range frames[0].Points {
			if first != frames[len(frames)-1].Points[i] {
				return nil, fmt.Errorf("motion: formation item %d jumps at loop boundary", i)
			}
		}
	}
	return &KeyframedFormation{count: config.Count, loop: config.Loop, frames: frames}, nil
}

// At returns the absolute anchor position of one item at a project time.
func (formation *KeyframedFormation) At(seconds float64, index int) Point {
	if formation == nil || index < 0 || index >= formation.count ||
		math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return Point{}
	}
	if formation.loop > 0 {
		seconds = Wrap(seconds, formation.loop)
	}
	frames := formation.frames
	if seconds <= frames[0].Time {
		return frames[0].Points[index]
	}
	next := sort.Search(len(frames), func(i int) bool { return frames[i].Time > seconds })
	if next == len(frames) {
		return frames[next-1].Points[index]
	}
	a, b := frames[next-1], frames[next]
	fraction := (seconds - a.Time) / (b.Time - a.Time)
	if a.Ease != nil {
		fraction = a.Ease(fraction)
	}
	return Point{
		X: Lerp(a.Points[index].X, b.Points[index].X, fraction),
		Y: Lerp(a.Points[index].Y, b.Points[index].Y, fraction),
	}
}
