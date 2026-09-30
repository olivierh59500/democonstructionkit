package sound

import (
	"math"
	"math/rand"
	"testing"
)

func nativePeriodMeterConfig() YMPeriodMeterConfig {
	return YMPeriodMeterConfig{Columns: 80,
		Levels: [16]float64{62, 161, 265, 377, 580, 774, 1155, 1575, 2260, 3088, 4570, 6233, 9330, 13187, 21220, 32767},
		Gain:   .003, Decay: 3, PeriodMin: 1, PeriodMaxExclusive: 4095, FrequencyScale: 10000,
		Envelope: true, EnvelopeShift: 6, EnvelopeLevel: 15, EnvelopeGain: 1,
	}
}

func TestYMPeriodMeterNativeRegisterFixture(t *testing.T) {
	meter, err := NewYMPeriodMeter(nativePeriodMeterConfig())
	if err != nil {
		t.Fatal(err)
	}
	first := [14]uint8{145, 1, 134, 3, 35, 3, 9, 248, 14, 15, 14, 0, 0, 0}
	if err := meter.Step(first); err != nil {
		t.Fatal(err)
	}
	for column, want := range map[int]float64{11: 95.301, 12: 60.66, 25: 60.66} {
		if math.Abs(meter.Level(column)-want) > .001 {
			t.Fatalf("column %d: %.3f; want %.3f", column, meter.Level(column), want)
		}
	}
	first[7] = 255
	meter.Step(first)
	if math.Abs(meter.Level(25)-57.66) > .001 {
		t.Fatal("muted tone did not decay during the same step")
	}
	meter.Reset()
	if meter.Level(25) != 0 || meter.Level(-1) != 0 || meter.Level(80) != 0 || meter.Len() != 80 {
		t.Fatal("reset or bounds are incorrect")
	}
}

func TestYMPeriodMeterMatchesIndependentNativeProgram(t *testing.T) {
	meter, err := NewYMPeriodMeter(nativePeriodMeterConfig())
	if err != nil {
		t.Fatal(err)
	}
	var expected [80]float64
	levels := [16]uint16{62, 161, 265, 377, 580, 774, 1155, 1575, 2260, 3088, 4570, 6233, 9330, 13187, 21220, 32767}
	charge := func(period, strength int) {
		if period <= 0 || period >= 4095 {
			return
		}
		column := int(math.Round(10000 / float64(period)))
		if column >= 80 {
			return
		}
		expected[column] = math.Max(expected[column], float64(levels[strength])*.003)
	}
	random := rand.New(rand.NewSource(501))
	for tick := 0; tick < 4096; tick++ {
		var registers [14]uint8
		for index := range registers {
			registers[index] = uint8(random.Intn(256))
		}
		envelope := (int(registers[12])*256 + int(registers[11]&0xc0)) >> 6
		for channel := 0; channel < 3; channel++ {
			volume := registers[8+channel] & 0x1f
			period := int(registers[channel*2]) + int(registers[channel*2+1]&0xf)*256
			strength := int(volume & 0xf)
			if volume&0x10 != 0 {
				strength = 15
			} else if registers[7]&(1<<channel) != 0 {
				strength = 0
			}
			charge(period, strength)
			if volume&0x10 != 0 {
				charge(envelope, 15)
			}
		}
		if err := meter.Step(registers); err != nil {
			t.Fatal(err)
		}
		for column := range expected {
			expected[column] = math.Max(0, expected[column]-3)
			if got := meter.Level(column); got != expected[column] {
				t.Fatalf("tick %d column %d: %v; want %v", tick, column, got, expected[column])
			}
		}
	}
}

func TestYMPeriodMeterRoundingAndOffset(t *testing.T) {
	for rounding, want := range map[YMPeriodRounding]int{
		YMPeriodRoundNearest: 7, YMPeriodRoundFloor: 6, YMPeriodRoundCeil: 7, YMPeriodRoundTruncate: 6,
	} {
		config := YMPeriodMeterConfig{Columns: 10, Levels: [16]float64{0, 1}, Gain: 2,
			PeriodMin: 1, PeriodMaxExclusive: 4096, FrequencyScale: 9, ColumnOffset: 2, Rounding: rounding}
		meter, err := NewYMPeriodMeter(config)
		if err != nil {
			t.Fatal(err)
		}
		registers := [14]uint8{2}
		registers[8] = 1
		meter.Step(registers)
		for column := 0; column < 10; column++ {
			level := 0.0
			if column == want {
				level = 2
			}
			if meter.Level(column) != level {
				t.Fatalf("rounding %d column %d: %v; want %v", rounding, column, meter.Level(column), level)
			}
		}
	}
}

func TestYMPeriodMeterEnvelopeAndMixerPolicies(t *testing.T) {
	registers := [14]uint8{232, 3, 0, 0, 0, 0, 0, 1, 0x10, 0, 0, 0, 100}
	for _, gating := range []YMPeriodGating{YMPeriodGateFixed, YMPeriodGateAll, YMPeriodGateNone} {
		config := nativePeriodMeterConfig()
		config.Gating = gating
		config.EnvelopeGain = .5
		meter, err := NewYMPeriodMeter(config)
		if err != nil {
			t.Fatal(err)
		}
		meter.Step(registers)
		wantTone, wantEnvelope := 95.301, 46.1505
		if gating == YMPeriodGateAll {
			wantTone, wantEnvelope = 0, 0
		}
		if math.Abs(meter.Level(10)-wantTone) > 1e-9 || math.Abs(meter.Level(25)-wantEnvelope) > 1e-9 {
			t.Fatalf("gate %d: tone %v envelope %v", gating, meter.Level(10), meter.Level(25))
		}
	}
	config := nativePeriodMeterConfig()
	config.Envelope = false
	meter, err := NewYMPeriodMeter(config)
	if err != nil {
		t.Fatal(err)
	}
	meter.Step(registers)
	if meter.Level(10) == 0 || meter.Level(25) != 0 {
		t.Fatal("disabling envelope columns should retain the envelope-driven tone")
	}
	config.EnvelopeLevel = 7
	config.DisabledLevel = 5
	config.Gating = YMPeriodGateAll
	meter, err = NewYMPeriodMeter(config)
	if err != nil {
		t.Fatal(err)
	}
	meter.Step(registers)
	if meter.Level(10) != math.Max(0, config.Levels[5]*config.Gain-config.Decay) {
		t.Fatal("custom disabled volume level was not used")
	}
}

func TestYMPeriodMeterValidatesConfiguration(t *testing.T) {
	mutations := []func(*YMPeriodMeterConfig){
		func(c *YMPeriodMeterConfig) { c.Columns = 0 },
		func(c *YMPeriodMeterConfig) { c.Columns = 65537 },
		func(c *YMPeriodMeterConfig) { c.Gain = math.NaN() },
		func(c *YMPeriodMeterConfig) { c.Decay = -1 },
		func(c *YMPeriodMeterConfig) { c.Levels[5] = -1 },
		func(c *YMPeriodMeterConfig) { c.Levels[5] = math.Inf(1) },
		func(c *YMPeriodMeterConfig) { c.PeriodMin = 0 },
		func(c *YMPeriodMeterConfig) { c.PeriodMaxExclusive = 65537 },
		func(c *YMPeriodMeterConfig) { c.PeriodMaxExclusive = c.PeriodMin },
		func(c *YMPeriodMeterConfig) { c.FrequencyScale = math.Inf(1) },
		func(c *YMPeriodMeterConfig) { c.ColumnOffset = 1<<20 + 1 },
		func(c *YMPeriodMeterConfig) { c.Rounding = 4 },
		func(c *YMPeriodMeterConfig) { c.Gating = 3 },
		func(c *YMPeriodMeterConfig) { c.DisabledLevel = 16 },
		func(c *YMPeriodMeterConfig) { c.EnvelopeShift = 16 },
		func(c *YMPeriodMeterConfig) { c.EnvelopeLevel = 16 },
		func(c *YMPeriodMeterConfig) { c.EnvelopeGain = math.NaN() },
		func(c *YMPeriodMeterConfig) { c.Gain = math.MaxFloat64 },
		func(c *YMPeriodMeterConfig) { c.EnvelopeGain = math.MaxFloat64 },
	}
	for index, mutate := range mutations {
		config := nativePeriodMeterConfig()
		mutate(&config)
		if _, err := NewYMPeriodMeter(config); err == nil {
			t.Fatalf("invalid config %d accepted", index)
		}
	}
	var meter *YMPeriodMeter
	if meter.Len() != 0 || meter.Level(0) != 0 || meter.Step([14]uint8{}) == nil {
		t.Fatal("nil meter was not handled")
	}
	meter.Reset()
}

func TestYMPeriodMeterDoesNotAllocatePerSnapshot(t *testing.T) {
	meter, err := NewYMPeriodMeter(nativePeriodMeterConfig())
	if err != nil {
		t.Fatal(err)
	}
	registers := [14]uint8{145, 1, 134, 3, 35, 3, 9, 248, 14, 15, 14}
	if allocations := testing.AllocsPerRun(100, func() { meter.Step(registers) }); allocations != 0 {
		t.Fatalf("allocated %v times per snapshot", allocations)
	}
}
