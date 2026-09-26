package motion

import "fmt"

// SteppedRevealConfig reveals a bounded number of blocks at regular tick
// intervals, then holds the completed picture until DoneTick. A zero DoneTick
// keeps the reveal available indefinitely after all blocks are visible.
type SteppedRevealConfig struct {
	Blocks, StartVisible, BlocksPerStep int
	StepTicks, DoneTick                 uint64
}

// SteppedReveal owns no images and can drive rows, columns or arbitrary block
// orders. Tick zero shows StartVisible blocks; Update/Step never allocates.
type SteppedReveal struct {
	config SteppedRevealConfig
	tick   uint64
}

func NewSteppedReveal(c SteppedRevealConfig) (*SteppedReveal, error) {
	if c.Blocks < 1 || c.Blocks > 1<<20 || c.StartVisible < 0 || c.StartVisible > c.Blocks ||
		c.BlocksPerStep < 1 || c.BlocksPerStep > 1<<20 || c.StepTicks == 0 {
		return nil, fmt.Errorf("motion: invalid stepped reveal timing")
	}
	return &SteppedReveal{config: c}, nil
}

func (r *SteppedReveal) Step() error {
	if r.tick == ^uint64(0) {
		return fmt.Errorf("motion: stepped reveal tick overflow")
	}
	r.tick++
	return nil
}

// SetTick supports deterministic seeking or a scene/timeline cue.
func (r *SteppedReveal) SetTick(tick uint64) { r.tick = tick }
func (r *SteppedReveal) Tick() uint64        { return r.tick }
func (r *SteppedReveal) Done() bool {
	return r.config.DoneTick > 0 && r.tick >= r.config.DoneTick
}
func (r *SteppedReveal) VisibleCount() int {
	steps := r.tick / r.config.StepTicks
	remaining := r.config.Blocks - r.config.StartVisible
	if steps >= uint64((remaining+r.config.BlocksPerStep-1)/r.config.BlocksPerStep) {
		return r.config.Blocks
	}
	return r.config.StartVisible + int(steps)*r.config.BlocksPerStep
}
