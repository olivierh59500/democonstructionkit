package modulation

import (
	"encoding/json"
	"math"
	"testing"
)

func TestSignalRoundTripAndTempoInputComposition(t *testing.T) {
	c := Spec{Base: 2, TimeBase: Beats, Oscillators: []Oscillator{{Shape: Cosine, Amplitude: 3, Frequency: 1}}, Keys: []Key{{Time: 0, Value: 0}, {Time: 4, Value: 8}}, Inputs: []Input{{Name: "voice.0", Gain: .5}}, Clamp: &Range{Min: 0, Max: 20}}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var restored Spec
	if err = json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	s, err := New(restored)
	if err != nil {
		t.Fatal(err)
	}
	ctx := MusicContext(1, 120, 0, map[string]float64{"voice.0": 4})
	if got := s.At(ctx); math.Abs(got-11) > 1e-12 {
		t.Fatal(got)
	}
	if allocs := testing.AllocsPerRun(100, func() { s.At(ctx) }); allocs != 0 {
		t.Fatal(allocs)
	}
}
func TestEnvelopeMatchesAuthoredVoiceFrames(t *testing.T) {
	d, _ := NewDecay(DecayConfig{Peak: 7, Rate: 1})
	var change Change[uint8]
	inputs := []uint8{0, 4, 4, 4, 4, 4, 4, 4, 4, 1, 1}
	want := []float64{0, 7, 6, 5, 4, 3, 2, 1, 0, 7, 6}
	for i, v := range inputs {
		if got := d.Step(change.Sample(v), 1); got != want[i] {
			t.Fatal(i, got, want[i])
		}
	}
}
func TestSignalOwnsConfigurationAndRejectsInvalidData(t *testing.T) {
	keys := []Key{{Time: 0, Value: 1}, {Time: 2, Value: 3}}
	s, _ := New(Spec{Keys: keys})
	keys[0].Value = 99
	if s.At(Context{Seconds: 1}) != 2 {
		t.Fatal("configuration aliases caller")
	}
	for _, c := range []Spec{{Base: math.NaN()}, {Loop: -1}, {Keys: []Key{{Time: 1}, {Time: 0}}}, {Oscillators: []Oscillator{{Shape: "unknown"}}}, {Clamp: &Range{Min: 2, Max: 1}}} {
		if _, err := New(c); err == nil {
			t.Fatal("invalid spec accepted")
		}
	}
}
