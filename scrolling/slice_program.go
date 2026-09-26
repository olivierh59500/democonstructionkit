package scrolling

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// SliceProgramConfig combines an existing bounded strip transport with a
// configurable text-pause/rotation clock. Offsets are indexed in visible
// painter order. OnControl runs after the configured cue has changed the clock.
type SliceProgramConfig struct {
	Stream    SliceStreamConfig
	Clock     motion.CuedScrollClockConfig
	Offsets   []float64
	OnControl func(SliceControl)
}

// SliceProgram owns transport and film-frame clocks, while borrowing the
// optional DNAFrames artwork passed to Draw. Text insertion, control events,
// pause expiry and rotation advance exactly once per Step.
type SliceProgram struct {
	stream    *SliceStream
	clock     *motion.CuedScrollClock
	offsets   []float64
	onControl func(SliceControl)
	insert    func(int)
	handle    func(SliceControl) bool
}

func NewSliceProgram(c SliceProgramConfig) (*SliceProgram, error) {
	stream, err := NewSliceStream(c.Stream)
	if err != nil {
		return nil, err
	}
	clock, err := motion.NewCuedScrollClock(c.Clock)
	if err != nil {
		return nil, err
	}
	if len(c.Offsets) != 0 && len(c.Offsets) != len(stream.Slices()) {
		return nil, fmt.Errorf("scrolling: invalid slice program phase profile")
	}
	for _, offset := range c.Offsets {
		if math.IsNaN(offset) || math.IsInf(offset, 0) {
			return nil, fmt.Errorf("scrolling: nonfinite slice program phase profile")
		}
	}
	p := &SliceProgram{stream: stream, clock: clock,
		offsets: append([]float64(nil), c.Offsets...), onControl: c.OnControl}
	p.handle = p.handleControl
	p.insert = func(count int) { p.Insert(count) }
	return p, nil
}

func (p *SliceProgram) handleControl(event SliceControl) bool {
	stop := p.clock.OnControl(event.Name)
	if p.onControl != nil {
		p.onControl(event)
	}
	return stop
}

// Insert bypasses the pause gate. Use it for authored intro pre-rolls and
// explicit control of the number of newly inserted source strips.
func (p *SliceProgram) Insert(count int) int {
	return p.stream.Step(count, p.handle)
}

// AdvanceRotation changes film frames without inserting text.
func (p *SliceProgram) AdvanceRotation(step float64) error {
	if err := p.clock.AdvanceRotation(step); err != nil {
		return err
	}
	return p.stream.SetFrames(p.clock.State().Rotation, p.offsets, p.clock.RotationFrames())
}

// Step inserts text unless paused, applies controls on that same tick, then
// advances film rotation. Its borrowed samples are ready for Draw afterward.
func (p *SliceProgram) Step() error {
	if err := p.clock.Step(p.insert); err != nil {
		return err
	}
	return p.stream.SetFrames(p.clock.State().Rotation, p.offsets, p.clock.RotationFrames())
}

// Warmup performs insertion and rotation without applying the pause gate. A
// control encountered during pre-roll still changes the following phase step.
func (p *SliceProgram) Warmup(ticks, slicesPerTick int) error {
	if ticks < 0 || slicesPerTick < 0 {
		return fmt.Errorf("scrolling: invalid slice program pre-roll")
	}
	for range ticks {
		p.Insert(slicesPerTick)
		if err := p.AdvanceRotation(p.clock.State().RotationStep); err != nil {
			return err
		}
	}
	return nil
}

func (p *SliceProgram) Reset() {
	p.stream.Reset()
	p.clock.Reset()
}

func (p *SliceProgram) Stream() *SliceStream           { return p.stream }
func (p *SliceProgram) Clock() *motion.CuedScrollClock { return p.clock }
