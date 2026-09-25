package motion

import "fmt"

// LinearTickConfig samples a clamped value from an absolute simulation tick.
// Negative ticks hold Start; there is no per-frame state or allocation.
type LinearTickConfig struct{ Start, Step, Min, Max int }

type LinearTick struct{ config LinearTickConfig }

func NewLinearTick(config LinearTickConfig) (LinearTick, error) {
	if config.Min > config.Max || config.Start < config.Min || config.Start > config.Max {
		return LinearTick{}, fmt.Errorf("motion: invalid linear tick range")
	}
	return LinearTick{config: config}, nil
}

func (line LinearTick) At(tick int) int {
	if tick <= 0 || line.config.Step == 0 {
		return line.config.Start
	}
	if line.config.Step < 0 {
		remaining := line.config.Start - line.config.Min
		if tick >= remaining/(-line.config.Step)+1 {
			return line.config.Min
		}
	} else {
		remaining := line.config.Max - line.config.Start
		if tick >= remaining/line.config.Step+1 {
			return line.config.Max
		}
	}
	return line.config.Start + tick*line.config.Step
}
