package timeline

import (
	"fmt"
	"image/color"
	"math"
)

// CardTintRamp interpolates RGBA channels over Frames samples. Reverse counts
// down from To using the remaining frame count; this preserves authored fade
// formulas whose rounding differs from a forward interpolation.
type CardTintRamp struct {
	Frames   int
	From, To color.RGBA
	Reverse  bool
}

type TintedCardsConfig struct {
	Count        int
	Ramps        []CardTintRamp
	HandoffTicks int // Zero defaults to one black handoff tick.
}

type TintedCardFrame struct {
	Card      int
	Tint      color.RGBA
	Cue       int // -1 ordinarily; next card index on the handoff tick.
	Handoff   bool
	Completed bool
}

// TintedCards prepares one card's color cycle and emits cue events exactly on
// handoff ticks. Assets, audio filenames and the screen's layer order remain
// caller-owned. Next advances one simulation tick and allocates nothing.
type TintedCards struct {
	colors     []color.RGBA
	count      int
	handoff    int
	card, tick int
	hold       int
	completed  bool
}

func NewTintedCards(config TintedCardsConfig) (*TintedCards, error) {
	if config.Count < 1 || config.Count > 1<<16 || len(config.Ramps) == 0 || len(config.Ramps) > 1<<16 ||
		config.HandoffTicks < 0 || config.HandoffTicks > 1<<20 {
		return nil, fmt.Errorf("timeline: invalid tinted card count or handoff")
	}
	if config.HandoffTicks == 0 {
		config.HandoffTicks = 1
	}
	count := 0
	for _, ramp := range config.Ramps {
		if ramp.Frames < 1 || ramp.Frames > 1<<20-count {
			return nil, fmt.Errorf("timeline: invalid card tint ramp length")
		}
		count += ramp.Frames
	}
	sequence := &TintedCards{colors: make([]color.RGBA, 0, count), count: config.Count, handoff: config.HandoffTicks}
	for _, ramp := range config.Ramps {
		for frame := 0; frame < ramp.Frames; frame++ {
			sequence.colors = append(sequence.colors, sampleCardTint(ramp, frame))
		}
	}
	return sequence, nil
}

func sampleCardTint(ramp CardTintRamp, frame int) color.RGBA {
	channel := func(from, to uint8) uint8 {
		if ramp.Frames == 1 {
			return to
		}
		last := float64(ramp.Frames - 1)
		value := float64(from) + float64(frame)*(float64(to)-float64(from))/last
		if ramp.Reverse {
			value = float64(to) + float64(ramp.Frames-1-frame)*(float64(from)-float64(to))/last
		}
		return uint8(max(0, min(255, math.Floor(value))))
	}
	return color.RGBA{
		R: channel(ramp.From.R, ramp.To.R),
		G: channel(ramp.From.G, ramp.To.G),
		B: channel(ramp.From.B, ramp.To.B),
		A: channel(ramp.From.A, ramp.To.A),
	}
}

func (sequence *TintedCards) Next() TintedCardFrame {
	if sequence.completed {
		return TintedCardFrame{Card: sequence.count, Cue: -1, Completed: true}
	}
	if sequence.hold > 0 {
		sequence.hold--
		return TintedCardFrame{Card: sequence.card, Tint: color.RGBA{A: 255}, Cue: -1, Handoff: true}
	}
	if sequence.tick == len(sequence.colors) {
		sequence.tick = 0
		sequence.card++
		sequence.hold = sequence.handoff - 1
		sequence.completed = sequence.card == sequence.count
		return TintedCardFrame{Card: sequence.card, Tint: color.RGBA{A: 255}, Cue: sequence.card,
			Handoff: true, Completed: sequence.completed}
	}
	frame := TintedCardFrame{Card: sequence.card, Tint: sequence.colors[sequence.tick], Cue: -1}
	sequence.tick++
	return frame
}

func (sequence *TintedCards) Reset() {
	sequence.card, sequence.tick, sequence.hold = 0, 0, 0
	sequence.completed = false
}

func (sequence *TintedCards) Completed() bool { return sequence.completed }
func (sequence *TintedCards) Card() int       { return sequence.card }
