package sound

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/modulation"
)

// YMPeriodRounding selects how a reciprocal period becomes a column index.
type YMPeriodRounding uint8

const (
	YMPeriodRoundNearest YMPeriodRounding = iota
	YMPeriodRoundFloor
	YMPeriodRoundCeil
	YMPeriodRoundTruncate
)

// YMPeriodGating selects which voices obey the YM mixer tone-disable bits.
// A gated voice selects DisabledLevel rather than losing its period entirely.
type YMPeriodGating uint8

const (
	YMPeriodGateFixed YMPeriodGating = iota // Envelope-driven tone and envelope charges bypass the mixer.
	YMPeriodGateAll                         // Gate envelope-driven tones too, and omit their envelope charges.
	YMPeriodGateNone                        // Ignore mixer tone-disable bits.
)

// YMPeriodMeterConfig maps register snapshots to a bank of visual peaks. Levels
// is an arbitrary nonnegative sixteen-entry volume curve. Gain converts it into
// visual units; Decay removes units after all charges in each Step. Periods use
// [PeriodMin, PeriodMaxExclusive); FrequencyScale/period+ColumnOffset is rounded
// with Rounding, and columns outside the bank are ignored.
//
// Envelope adds a charge from the sixteen-bit envelope period, shifted right by
// EnvelopeShift. EnvelopeLevel also selects the tone strength when the volume
// register's envelope flag is set; EnvelopeGain scales only the extra envelope
// charge. No decoder or playback state is advanced by this adapter.
type YMPeriodMeterConfig struct {
	Columns                       int
	Levels                        [16]float64
	Gain, Decay                   float64
	PeriodMin, PeriodMaxExclusive int
	FrequencyScale                float64
	ColumnOffset                  int
	Rounding                      YMPeriodRounding
	Gating                        YMPeriodGating
	DisabledLevel                 uint8
	Envelope                      bool
	EnvelopeShift, EnvelopeLevel  uint8
	EnvelopeGain                  float64
}

// YMPeriodMeter owns a reusable peak bank and consumes the public fourteen-byte
// snapshot returned by Stream.YMRegisters. Step performs no allocation or audio
// decoding. Level can drive bars, sprites, lights or any other effect.
type YMPeriodMeter struct {
	config  YMPeriodMeterConfig
	peaks   *modulation.PeakBank
	charges [6]modulation.PeakCharge
}

func NewYMPeriodMeter(c YMPeriodMeterConfig) (*YMPeriodMeter, error) {
	if c.Columns < 1 || c.Columns > 65536 || c.PeriodMin < 1 || c.PeriodMaxExclusive > 65536 ||
		c.PeriodMaxExclusive <= c.PeriodMin || !finitePeriodValue(c.Gain) || c.Gain < 0 ||
		!finitePeriodValue(c.Decay) || c.Decay < 0 || !finitePeriodValue(c.FrequencyScale) || c.FrequencyScale < 0 ||
		c.ColumnOffset < -1<<20 || c.ColumnOffset > 1<<20 || c.Rounding > YMPeriodRoundTruncate ||
		c.Gating > YMPeriodGateNone || c.DisabledLevel > 15 || c.EnvelopeShift > 15 || c.EnvelopeLevel > 15 ||
		!finitePeriodValue(c.EnvelopeGain) || c.EnvelopeGain < 0 {
		return nil, fmt.Errorf("sound: invalid YM period meter")
	}
	ceiling := 0.0
	for _, level := range c.Levels {
		if !finitePeriodValue(level) || level < 0 {
			return nil, fmt.Errorf("sound: invalid YM period volume curve")
		}
		ceiling = math.Max(ceiling, level)
	}
	ceiling *= c.Gain
	if c.Envelope {
		ceiling *= math.Max(1, c.EnvelopeGain)
	}
	if !finitePeriodValue(ceiling) {
		return nil, fmt.Errorf("sound: YM period meter gain overflows")
	}
	peaks, err := modulation.NewPeakBank(modulation.PeakBankConfig{Count: c.Columns, Ceiling: ceiling, Decay: c.Decay})
	if err != nil {
		return nil, err
	}
	return &YMPeriodMeter{config: c, peaks: peaks}, nil
}

// Step charges tone and optional envelope columns before the same-tick decay.
// Call once per simulation update; drawing must only read Level.
func (m *YMPeriodMeter) Step(registers [14]uint8) error {
	if m == nil || m.peaks == nil {
		return fmt.Errorf("sound: unavailable YM period meter")
	}
	count := 0
	charge := func(period, level int, gain float64) {
		column, ok := m.column(period)
		if !ok {
			return
		}
		m.charges[count] = modulation.PeakCharge{Index: column, Level: m.config.Levels[level] * m.config.Gain * gain}
		count++
	}
	envelopePeriod := (int(registers[11]) | int(registers[12])<<8) >> m.config.EnvelopeShift
	for channel := 0; channel < 3; channel++ {
		volume := registers[8+channel] & 0x1f
		envelope := volume&0x10 != 0
		disabled := registers[7]&(1<<channel) != 0
		level := int(volume & 0x0f)
		if envelope {
			level = int(m.config.EnvelopeLevel)
		}
		gated := disabled && (m.config.Gating == YMPeriodGateAll || m.config.Gating == YMPeriodGateFixed && !envelope)
		if gated {
			level = int(m.config.DisabledLevel)
		}
		period := int(registers[channel*2]) | int(registers[channel*2+1]&0xf)<<8
		charge(period, level, 1)
		if envelope && m.config.Envelope && !gated {
			charge(envelopePeriod, int(m.config.EnvelopeLevel), m.config.EnvelopeGain)
		}
	}
	return m.peaks.Step(m.charges[:count], 1)
}

func (m *YMPeriodMeter) column(period int) (int, bool) {
	if period < m.config.PeriodMin || period >= m.config.PeriodMaxExclusive {
		return 0, false
	}
	value := m.config.FrequencyScale/float64(period) + float64(m.config.ColumnOffset)
	switch m.config.Rounding {
	case YMPeriodRoundNearest:
		value = math.Round(value)
	case YMPeriodRoundFloor:
		value = math.Floor(value)
	case YMPeriodRoundCeil:
		value = math.Ceil(value)
	case YMPeriodRoundTruncate:
		value = math.Trunc(value)
	}
	if value < 0 || value >= float64(m.config.Columns) {
		return 0, false
	}
	return int(value), true
}

func (m *YMPeriodMeter) Level(column int) float64 {
	if m == nil {
		return 0
	}
	return m.peaks.Level(column)
}

func (m *YMPeriodMeter) Len() int {
	if m == nil {
		return 0
	}
	return m.peaks.Len()
}

func (m *YMPeriodMeter) Reset() {
	if m != nil {
		m.peaks.Reset()
	}
}

func finitePeriodValue(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
