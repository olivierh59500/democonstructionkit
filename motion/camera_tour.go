package motion

import (
	"fmt"
	"math"
)

type TourEase uint8

const (
	TourLinear TourEase = iota
	TourCubicInOut
	TourSineInOut
)

type TourCoordinates uint8

const (
	TourCameraOffset TourCoordinates = iota
	TourWorldCenter
)

// TourPose contains the position and zoom before the viewport-center offset.
// TourCameraOffset adds half the viewport after interpolation; TourWorldCenter
// interprets X and Y directly as the center of the world view.
type TourPose struct{ X, Y, Zoom float64 }

// TourSegment is a hold or an eased handoff. A direct scene index >= 0 avoids
// an intermediate surface during a fixed view; -1 selects the composite path.
// VisibleMask tells a renderer which retained scene canvases to refresh.
type TourSegment struct {
	Name        string
	Duration    float64
	From, To    TourPose
	Coordinates TourCoordinates
	Ease        TourEase
	Direct      int
	VisibleMask uint64
}

type CameraTourConfig struct {
	ViewportWidth, ViewportHeight float64
	StepSeconds                   float64
	Segments                      []TourSegment
	Initial                       int
}

type CameraTourState struct {
	Segment          int
	Elapsed          float64
	CenterX, CenterY float64
	Zoom             float64
	VisibleMask      uint64
	Direct           int
}

// CameraTour advances exactly one segment per tick. A handoff tick selects the
// next segment at elapsed zero; it does not also advance the destination.
// Scene updates are independent so every source may keep running continuously.
type CameraTour struct {
	config CameraTourConfig
	state  CameraTourState
}

func NewCameraTour(c CameraTourConfig) (*CameraTour, error) {
	if !finiteTour(c.ViewportWidth) || !finiteTour(c.ViewportHeight) ||
		c.ViewportWidth <= 0 || c.ViewportHeight <= 0 ||
		!finiteTour(c.StepSeconds) || c.StepSeconds <= 0 ||
		len(c.Segments) == 0 || len(c.Segments) > 1024 ||
		c.Initial < 0 || c.Initial >= len(c.Segments) {
		return nil, fmt.Errorf("motion: invalid camera tour")
	}
	segments := make([]TourSegment, len(c.Segments))
	for i, segment := range c.Segments {
		if segment.Name == "" || !finiteTour(segment.Duration) || segment.Duration <= 0 ||
			segment.Coordinates > TourWorldCenter || segment.Ease > TourSineInOut ||
			segment.Direct < -1 || segment.Direct >= 64 ||
			segment.Direct >= 0 && segment.VisibleMask&(uint64(1)<<segment.Direct) == 0 ||
			!finiteTourPose(segment.From) || !finiteTourPose(segment.To) ||
			segment.From.Zoom <= 0 || segment.To.Zoom <= 0 {
			return nil, fmt.Errorf("motion: invalid camera tour segment %d", i)
		}
		segments[i] = segment
	}
	c.Segments = segments
	tour := &CameraTour{config: c}
	tour.Reset()
	return tour, nil
}

func finiteTour(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func finiteTourPose(p TourPose) bool {
	return finiteTour(p.X) && finiteTour(p.Y) && finiteTour(p.Zoom)
}

func (t *CameraTour) Reset() {
	t.state = CameraTourState{Segment: t.config.Initial}
	t.sample()
}

func (t *CameraTour) Step() {
	segment := t.config.Segments[t.state.Segment]
	t.state.Elapsed += t.config.StepSeconds
	if t.state.Elapsed >= segment.Duration {
		t.state.Segment++
		if t.state.Segment == len(t.config.Segments) {
			t.state.Segment = 0
		}
		t.state.Elapsed = 0
	}
	t.sample()
}

func (t *CameraTour) sample() {
	segment := t.config.Segments[t.state.Segment]
	progress := t.state.Elapsed / segment.Duration
	if progress > 1 {
		progress = 1
	}
	eased := tourEase(segment.Ease, progress)
	x := segment.From.X + (segment.To.X-segment.From.X)*eased
	y := segment.From.Y + (segment.To.Y-segment.From.Y)*eased
	zoom := segment.From.Zoom + (segment.To.Zoom-segment.From.Zoom)*eased
	if segment.Coordinates == TourCameraOffset {
		x += t.config.ViewportWidth / 2
		y += t.config.ViewportHeight / 2
	}
	t.state.CenterX, t.state.CenterY, t.state.Zoom = x, y, zoom
	t.state.VisibleMask, t.state.Direct = segment.VisibleMask, segment.Direct
}

func tourEase(kind TourEase, value float64) float64 {
	switch kind {
	case TourCubicInOut:
		if value < .5 {
			return 4 * value * value * value
		}
		u := -2*value + 2
		return 1 - u*u*u/2
	case TourSineInOut:
		return (1 - math.Cos(math.Pi*value)) / 2
	default:
		return value
	}
}

func (t *CameraTour) State() CameraTourState { return t.state }
func (t *CameraTour) SegmentName() string    { return t.config.Segments[t.state.Segment].Name }
