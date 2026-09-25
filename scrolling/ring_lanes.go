package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// RingLanesConfig combines independent bitmap rings in one synchronized
// presentation. Every ring may choose its own font, message, speed and control
// commands. Y contains one initial vertical position per ring. ShiftEvery
// advances the optional wrapped position bank in simulation ticks; ShiftFirst
// is the first tick to shift, defaulting to ShiftEvery when zero.
type RingLanesConfig struct {
	Rings         []RingConfig
	Y             []float64
	ShiftEvery    int
	ShiftFirst    int
	ShiftVelocity []float64
	ShiftLower    *motion.WrapLimit
	ShiftUpper    *motion.WrapLimit
}

// RingLanes owns each ring's glyph transport and the optional paced lane
// positions. DrawAt and DrawLaneAt never advance simulation state.
type RingLanes struct {
	rings      []*Ring
	y          []float64
	shift      *motion.WrapBank
	shiftEvery int
	shiftFirst int
	tick       int
}

func NewRingLanes(config RingLanesConfig) (*RingLanes, error) {
	if len(config.Rings) == 0 || len(config.Rings) != len(config.Y) || len(config.Rings) > 16384 || config.ShiftEvery < 0 || config.ShiftFirst < 0 {
		return nil, fmt.Errorf("scrolling: invalid ring lane dimensions or cadence")
	}
	for _, y := range config.Y {
		if math.IsNaN(y) || math.IsInf(y, 0) {
			return nil, fmt.Errorf("scrolling: nonfinite ring lane position")
		}
	}
	if config.ShiftEvery == 0 && (config.ShiftFirst != 0 || len(config.ShiftVelocity) != 0 || config.ShiftLower != nil || config.ShiftUpper != nil) {
		return nil, fmt.Errorf("scrolling: shift parameters require a cadence")
	}
	lanes := &RingLanes{rings: make([]*Ring, len(config.Rings)), y: append([]float64(nil), config.Y...), shiftEvery: config.ShiftEvery, shiftFirst: config.ShiftFirst}
	for index, ringConfig := range config.Rings {
		ring, err := NewRing(ringConfig)
		if err != nil {
			return nil, fmt.Errorf("scrolling: ring lane %d: %w", index, err)
		}
		lanes.rings[index] = ring
	}
	if config.ShiftEvery > 0 {
		if lanes.shiftFirst == 0 {
			lanes.shiftFirst = config.ShiftEvery
		}
		var err error
		lanes.shift, err = motion.NewWrapBank(motion.WrapBankConfig{
			Start: config.Y, Velocity: config.ShiftVelocity,
			Lower: config.ShiftLower, Upper: config.ShiftUpper,
		})
		if err != nil {
			return nil, err
		}
	}
	return lanes, nil
}

func (lanes *RingLanes) Len() int { return len(lanes.rings) }

// Update makes RingLanes usable as a scrolling.New transport.
func (lanes *RingLanes) Update(kit.Frame) error { lanes.Step(); return nil }

// Draw renders at the configured lane positions without an additional offset.
func (lanes *RingLanes) Draw(dst *ebiten.Image) { lanes.DrawAt(dst, 0, 0) }

// Step advances all text transports, then moves the lane bank on its authored
// cadence. Calling Draw more than once for the same tick does not move text.
func (lanes *RingLanes) Step() {
	for _, ring := range lanes.rings {
		ring.Step()
	}
	lanes.tick++
	if lanes.shift != nil && lanes.tick >= lanes.shiftFirst && (lanes.tick-lanes.shiftFirst)%lanes.shiftEvery == 0 {
		lanes.shift.Step()
	}
}

func (lanes *RingLanes) Y(index int) float64 {
	if lanes.shift != nil {
		return lanes.shift.At(index)
	}
	return lanes.y[index]
}

// DrawAt renders all rings into one destination at their independent Y values.
func (lanes *RingLanes) DrawAt(dst *ebiten.Image, x, y float64) {
	for index := range lanes.rings {
		lanes.DrawLaneAt(index, dst, x, y)
	}
}

// DrawLaneAt renders one ring, for screens where each font has its own mask or
// destination surface. All lanes still share the same Step boundary.
func (lanes *RingLanes) DrawLaneAt(index int, dst *ebiten.Image, x, y float64) {
	lanes.rings[index].DrawAt(dst, x, y+lanes.Y(index))
}
