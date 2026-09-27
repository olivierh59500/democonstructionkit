package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/modulation"
)

// SignalFrameTrigger selects when a channel retriggers its sprite envelope.
// SignalOnChange preserves register-driven animation. SignalOnRising handles
// continuous PCM or module levels that change every sample.
type SignalFrameTrigger uint8

const (
	SignalOnChange SignalFrameTrigger = iota
	SignalOnRising
)

// SignalFrameSlot connects one sprite to an input channel. FrameOffset selects
// another palette or animation sequence within the same borrowed atlas.
// Zero scales and opacity mean one, matching ordinary DrawImage defaults.
type SignalFrameSlot struct {
	Channel, FrameOffset                  int
	X, Y, AnchorX, AnchorY                float64
	ScaleX, ScaleY, AngleDegrees, Opacity float64
	Trigger                               SignalFrameTrigger
	Threshold                             float64
	Envelope                              *modulation.DecayConfig // Nil uses SignalFrameBankConfig.Envelope.
}

// SignalFrameBankConfig controls a bank of music- or input-triggered sprites.
// Atlas regions, positions, channel routing, envelopes and draw material are
// independent, so several slots may share one input but animate differently.
type SignalFrameBankConfig struct {
	Atlas    *Atlas
	Slots    []SignalFrameSlot
	Envelope modulation.DecayConfig
	Filter   ebiten.Filter
	Blend    ebiten.Blend
}

// SignalFrameBank owns channel-change detectors and decay state. The host
// supplies current YM register values or arbitrary finite numeric signals once
// per update; Draw borrows the atlas and never advances a clock.
type SignalFrameBank struct {
	config     SignalFrameBankConfig
	changes    []modulation.Change[float64]
	previous   []float64
	envelopes  []*modulation.Decay
	frames     []int
	maxChannel int
}

func NewSignalFrameBank(config SignalFrameBankConfig) (*SignalFrameBank, error) {
	if config.Atlas == nil || config.Atlas.Image == nil || config.Atlas.Count() == 0 ||
		len(config.Slots) == 0 || len(config.Slots) > 4096 {
		return nil, fmt.Errorf("sprites: invalid signal frame atlas or slot count")
	}
	config.Slots = append([]SignalFrameSlot(nil), config.Slots...)
	bank := &SignalFrameBank{config: config,
		changes:   make([]modulation.Change[float64], len(config.Slots)),
		previous:  make([]float64, len(config.Slots)),
		envelopes: make([]*modulation.Decay, len(config.Slots)),
		frames:    make([]int, len(config.Slots)),
	}
	for index := range bank.config.Slots {
		slot := &bank.config.Slots[index]
		if slot.Channel < 0 || slot.Channel > 65535 || slot.FrameOffset < 0 ||
			slot.FrameOffset >= config.Atlas.Count() ||
			slot.Trigger > SignalOnRising || !signalFrameFinite(slot.Threshold) ||
			slot.Trigger == SignalOnRising && slot.Threshold <= 0 ||
			!signalFrameFinite(slot.X, slot.Y, slot.AnchorX, slot.AnchorY,
				slot.ScaleX, slot.ScaleY, slot.AngleDegrees, slot.Opacity) {
			return nil, fmt.Errorf("sprites: invalid signal frame slot %d", index)
		}
		if slot.ScaleX == 0 {
			slot.ScaleX = 1
		}
		if slot.ScaleY == 0 {
			slot.ScaleY = 1
		}
		if slot.Opacity == 0 {
			slot.Opacity = 1
		}
		if slot.Opacity < 0 || slot.Opacity > 1 {
			return nil, fmt.Errorf("sprites: invalid signal sprite opacity")
		}
		bank.maxChannel = max(bank.maxChannel, slot.Channel)
		envelope := config.Envelope
		if slot.Envelope != nil {
			envelope = *slot.Envelope
		}
		if envelope.Floor < 0 || envelope.Peak >= float64(config.Atlas.Count()-slot.FrameOffset) {
			return nil, fmt.Errorf("sprites: signal envelope exceeds atlas frames")
		}
		decay, err := modulation.NewDecay(envelope)
		if err != nil {
			return nil, err
		}
		bank.envelopes[index] = decay
		bank.frames[index] = slot.FrameOffset + int(decay.Value())
		slot.Envelope = nil
	}
	if bank.config.Blend == (ebiten.Blend{}) {
		bank.config.Blend = ebiten.BlendSourceOver
	}
	return bank, nil
}

// StepYM consumes byte-sized voice values without converting a whole register
// bank or allocating. The zero previous value preserves silent startup.
func (bank *SignalFrameBank) StepYM(values []uint8) error {
	if bank == nil || len(values) <= bank.maxChannel {
		return fmt.Errorf("sprites: missing signal channels")
	}
	for index, slot := range bank.config.Slots {
		trigger := bank.trigger(index, float64(values[slot.Channel]))
		bank.frames[index] = slot.FrameOffset + int(bank.envelopes[index].Step(trigger, 1))
	}
	return nil
}

// StepSignals accepts module envelopes, analyzed PCM levels or user input.
// Delta is in the same units as each slot's decay rate. Invalid inputs leave
// every slot unchanged so a malformed music frame cannot desynchronize them.
func (bank *SignalFrameBank) StepSignals(values []float64, delta float64) error {
	if bank == nil || len(values) <= bank.maxChannel || !signalFrameFinite(delta) || delta < 0 {
		return fmt.Errorf("sprites: invalid signal frame input")
	}
	for _, slot := range bank.config.Slots {
		if !signalFrameFinite(values[slot.Channel]) {
			return fmt.Errorf("sprites: nonfinite signal channel")
		}
	}
	for index, slot := range bank.config.Slots {
		trigger := bank.trigger(index, values[slot.Channel])
		bank.frames[index] = slot.FrameOffset + int(bank.envelopes[index].Step(trigger, delta))
	}
	return nil
}

func (bank *SignalFrameBank) trigger(index int, value float64) bool {
	slot := bank.config.Slots[index]
	if slot.Trigger == SignalOnRising {
		triggered := bank.previous[index] < slot.Threshold && value >= slot.Threshold
		bank.previous[index] = value
		return triggered
	}
	return bank.changes[index].Sample(value)
}

// SetPosition composes a live DCK trajectory with one sprite without resetting
// its signal envelope or material.
func (bank *SignalFrameBank) SetPosition(index int, x, y float64) error {
	if bank == nil || index < 0 || index >= len(bank.config.Slots) || !signalFrameFinite(x, y) {
		return fmt.Errorf("sprites: invalid signal sprite position")
	}
	bank.config.Slots[index].X, bank.config.Slots[index].Y = x, y
	return nil
}

// Frame reports one prepared atlas index without exposing mutable state.
func (bank *SignalFrameBank) Frame(index int) int {
	if bank == nil || index < 0 || index >= len(bank.frames) {
		return -1
	}
	return bank.frames[index]
}

func (bank *SignalFrameBank) Reset() {
	if bank == nil {
		return
	}
	for index, slot := range bank.config.Slots {
		bank.changes[index] = modulation.Change[float64]{}
		bank.previous[index] = 0
		bank.envelopes[index].Reset()
		bank.frames[index] = slot.FrameOffset + int(bank.envelopes[index].Value())
	}
}

func (bank *SignalFrameBank) Draw(dst *ebiten.Image) {
	if bank == nil || dst == nil {
		return
	}
	for index, slot := range bank.config.Slots {
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = bank.config.Filter, bank.config.Blend
		options.GeoM.Translate(-slot.AnchorX, -slot.AnchorY)
		options.GeoM.Scale(slot.ScaleX, slot.ScaleY)
		options.GeoM.Rotate(slot.AngleDegrees * math.Pi / 180)
		options.GeoM.Translate(slot.X, slot.Y)
		if slot.Opacity != 1 {
			options.ColorScale.ScaleAlpha(float32(slot.Opacity))
		}
		composite.DrawRegion(dst, bank.config.Atlas.Image, bank.config.Atlas.Region(bank.frames[index]), &options)
	}
}

func (bank *SignalFrameBank) Close() error { return nil }

func signalFrameFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
