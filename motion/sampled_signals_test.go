package motion

import (
	"math"
	"testing"
	"time"
)

func TestSampledSignalsFollowExternalClockAndOwnTheirMasks(t *testing.T) {
	data := []byte{0, 1, 3, 2}
	s, err := NewSampledSignals(SampledSignalsConfig{Channels: 2, Rate: 50, Masks: data})
	if err != nil {
		t.Fatal(err)
	}
	clear(data)
	for _, sample := range []struct {
		position time.Duration
		frame    int
		on       [2]bool
	}{
		{-time.Second, 0, [2]bool{}},
		{20*time.Millisecond - 1, 0, [2]bool{}},
		{20 * time.Millisecond, 1, [2]bool{true, false}},
		{60 * time.Millisecond, 3, [2]bool{false, true}},
		{40 * time.Millisecond, 2, [2]bool{true, true}},
		{80 * time.Millisecond, -1, [2]bool{}},
		{time.Duration(math.MaxInt64), -1, [2]bool{}},
	} {
		if got := s.FrameAt(sample.position); got != sample.frame {
			t.Fatalf("at %v: frame %d, want %d", sample.position, got, sample.frame)
		}
		for channel, want := range sample.on {
			if got := s.At(sample.position, channel); got != want {
				t.Fatalf("at %v channel %d: got %v, want %v", sample.position, channel, got, want)
			}
		}
	}
	if s.At(0, -1) || s.At(0, 2) || s.Duration() != 80*time.Millisecond {
		t.Fatal("invalid channel or duration")
	}
	if allocations := testing.AllocsPerRun(1000, func() { s.At(40*time.Millisecond, 1) }); allocations != 0 {
		t.Fatalf("sampling allocated %g objects", allocations)
	}
}

func TestSampledSignalsLoopAfterIntroductionAtAnyHostRate(t *testing.T) {
	s, err := NewSampledSignals(SampledSignalsConfig{
		Channels: 1, Rate: 50, Masks: []byte{0, 0, 1, 0, 1}, Loop: true, LoopStart: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, hostRate := range []int64{50, 60, 144} {
		for tick := int64(0); tick < hostRate*10; tick++ {
			position := time.Duration((tick*int64(time.Second) + hostRate - 1) / hostRate)
			nativeFrame := tick * 50 / hostRate
			if nativeFrame >= 5 {
				nativeFrame = 2 + (nativeFrame-2)%3
			}
			if got := s.FrameAt(position); got != int(nativeFrame) {
				t.Fatalf("host %d tick %d: got %d, want %d", hostRate, tick, got, nativeFrame)
			}
		}
	}
	if s.FrameAt(100*time.Millisecond) != 2 || s.FrameAt(160*time.Millisecond) != 2 {
		t.Fatal("loop replayed the introduction")
	}
}

func TestSampledSignalsSupportSixtyFourChannelsAndRejectInvalidBanks(t *testing.T) {
	s, err := NewSampledSignals(SampledSignalsConfig{
		Channels: 64, Rate: 1, Masks: []byte{1, 0, 0, 0, 0, 0, 0, 128},
	})
	if err != nil || !s.At(0, 0) || !s.At(0, 63) || s.At(0, 1) {
		t.Fatal("invalid wide-channel sample", err)
	}
	for _, config := range []SampledSignalsConfig{
		{Channels: 0, Rate: 50, Masks: []byte{0}},
		{Channels: 65, Rate: 50, Masks: []byte{0}},
		{Channels: 1, Rate: 0, Masks: []byte{0}},
		{Channels: 1, Rate: 8001, Masks: []byte{0}},
		{Channels: 1, Rate: 50},
		{Channels: 9, Rate: 50, Masks: []byte{0}},
		{Channels: 1, Rate: 50, Masks: []byte{2}},
		{Channels: 1, Rate: 50, Masks: []byte{0}, LoopStart: -1},
		{Channels: 1, Rate: 50, Masks: []byte{0}, LoopStart: 1},
	} {
		if _, err := NewSampledSignals(config); err == nil {
			t.Fatalf("accepted invalid config %+v", config)
		}
	}
	var missing *SampledSignals
	if missing.FrameAt(0) != -1 || missing.At(0, 0) {
		t.Fatal("nil bank returned an active signal")
	}
}
