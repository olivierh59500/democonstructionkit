package timeline

import (
	"image/color"
	"testing"
)

func TestTintedCardsHoldAndResetRemainIndependentOfMusic(t *testing.T) {
	sequence, err := NewTintedCards(TintedCardsConfig{
		Count: 2, HandoffTicks: 2,
		Ramps: []CardTintRamp{{Frames: 2, From: color.RGBA{A: 255}, To: color.RGBA{R: 10, A: 255}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		card, red, cue int
		handoff, done  bool
	}{
		{0, 0, -1, false, false}, {0, 10, -1, false, false},
		{1, 0, 1, true, false}, {1, 0, -1, true, false},
		{1, 0, -1, false, false}, {1, 10, -1, false, false},
		{2, 0, 2, true, true},
	}
	for tick, expected := range want {
		frame := sequence.Next()
		if frame.Card != expected.card || int(frame.Tint.R) != expected.red || frame.Cue != expected.cue ||
			frame.Handoff != expected.handoff || frame.Completed != expected.done {
			t.Fatalf("tick %d frame = %+v, want %+v", tick, frame, expected)
		}
	}
	sequence.Reset()
	if frame := sequence.Next(); frame.Card != 0 || frame.Cue != -1 || frame.Completed {
		t.Fatalf("restart frame = %+v", frame)
	}
}

func TestTintedCardsRejectsInvalidLengths(t *testing.T) {
	for _, config := range []TintedCardsConfig{
		{Count: 0, Ramps: []CardTintRamp{{Frames: 2}}},
		{Count: 1, Ramps: []CardTintRamp{{Frames: 0}}},
		{Count: 1, Ramps: []CardTintRamp{{Frames: 2}}, HandoffTicks: -1},
	} {
		if _, err := NewTintedCards(config); err == nil {
			t.Fatalf("accepted invalid tinted cards: %+v", config)
		}
	}
}
