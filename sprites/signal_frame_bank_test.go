package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/modulation"
)

func signalTestAtlas(t *testing.T, count int) *Atlas {
	t.Helper()
	image := ebiten.NewImage(count, 1)
	t.Cleanup(image.Deallocate)
	atlas, err := NewAtlas(AtlasConfig{Image: image, TileW: 1, TileH: 1, Columns: count, Count: count})
	if err != nil {
		t.Fatal(err)
	}
	return atlas
}

func TestSignalFrameBankTracksIndependentYMVoicesAndDecay(t *testing.T) {
	bank, err := NewSignalFrameBank(SignalFrameBankConfig{
		Atlas:    signalTestAtlas(t, 8),
		Slots:    []SignalFrameSlot{{Channel: 0, X: 244, Y: 177}, {Channel: 1, X: 355, Y: 177}, {Channel: 2, X: 464, Y: 177}},
		Envelope: modulation.DecayConfig{Peak: 7, Rate: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		values [3]uint8
		want   [3]int
	}{
		{[3]uint8{0, 0, 0}, [3]int{0, 0, 0}},
		{[3]uint8{1, 0, 0}, [3]int{7, 0, 0}},
		{[3]uint8{1, 0, 0}, [3]int{6, 0, 0}},
		{[3]uint8{1, 0, 2}, [3]int{5, 0, 7}},
		{[3]uint8{1, 0, 2}, [3]int{4, 0, 6}},
		{[3]uint8{0, 0, 2}, [3]int{7, 0, 5}},
	} {
		if err := bank.StepYM(test.values[:]); err != nil {
			t.Fatal(err)
		}
		for index, want := range test.want {
			if got := bank.Frame(index); got != want {
				t.Fatalf("input %v slot %d frame = %d, want %d", test.values, index, got, want)
			}
		}
	}
	if err := bank.SetPosition(0, 250, 180); err != nil || bank.Frame(0) != 7 {
		t.Fatalf("live position changed frame: %v, %d", err, bank.Frame(0))
	}
	values := [...]uint8{0, 0, 2}
	if got := testing.AllocsPerRun(100, func() { _ = bank.StepYM(values[:]) }); got != 0 {
		t.Fatalf("YM signal update allocated %.2f objects", got)
	}
	bank.Reset()
	if bank.Frame(0) != 0 || bank.Frame(2) != 0 {
		t.Fatal("reset retained voice history or a visible frame")
	}
}

func TestSignalFrameBankAcceptsOtherMusicSignalsAndPaletteOffsets(t *testing.T) {
	bank, err := NewSignalFrameBank(SignalFrameBankConfig{
		Atlas: signalTestAtlas(t, 16),
		Slots: []SignalFrameSlot{{Channel: 1, FrameOffset: 8,
			Trigger: SignalOnRising, Threshold: .5,
			Envelope: &modulation.DecayConfig{Peak: 7, Rate: 1}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := bank.StepSignals([]float64{0, .8}, 1); err != nil || bank.Frame(0) != 15 {
		t.Fatalf("music signal frame = %d, error %v", bank.Frame(0), err)
	}
	if err := bank.StepSignals([]float64{0, .8}, .5); err != nil || bank.Frame(0) != 14 {
		t.Fatalf("half-step decay frame = %d, error %v", bank.Frame(0), err)
	}
	if err := bank.StepSignals([]float64{0, math.NaN()}, 1); err == nil || bank.Frame(0) != 14 {
		t.Fatal("invalid signal changed the frame")
	}
	if err := bank.StepSignals([]float64{0, .2}, 1); err != nil || bank.Frame(0) != 13 {
		t.Fatalf("falling signal frame = %d, error %v", bank.Frame(0), err)
	}
	if err := bank.StepSignals([]float64{0, .7}, 1); err != nil || bank.Frame(0) != 15 {
		t.Fatalf("rising signal did not retrigger: frame %d, error %v", bank.Frame(0), err)
	}
}

func TestSignalFrameBankRejectsMissingChannelsAndAtlasOverflow(t *testing.T) {
	atlas := signalTestAtlas(t, 8)
	config := SignalFrameBankConfig{Atlas: atlas, Slots: []SignalFrameSlot{{Channel: 2}},
		Envelope: modulation.DecayConfig{Peak: 7, Rate: 1}}
	bank, err := NewSignalFrameBank(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := bank.StepYM([]uint8{1, 2}); err == nil || bank.Frame(0) != 0 {
		t.Fatal("missing voice advanced an envelope")
	}
	config.Envelope.Peak = 8
	if _, err := NewSignalFrameBank(config); err == nil {
		t.Fatal("accepted a frame outside the atlas")
	}
	config.Envelope.Peak = 7
	config.Slots[0].X = math.Inf(1)
	if _, err := NewSignalFrameBank(config); err == nil {
		t.Fatal("accepted a nonfinite sprite position")
	}
	config.Slots[0].X = 0
	config.Slots[0].Trigger = SignalOnRising
	if _, err := NewSignalFrameBank(config); err == nil {
		t.Fatal("accepted a rising trigger without a positive threshold")
	}
}
