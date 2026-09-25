package scrolling

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestRingLanesPreserveSynchronizedTextAndPacedVerticalWrap(t *testing.T) {
	image := ebiten.NewImage(16, 8)
	defer image.Deallocate()
	font := BitmapGrid{Image: image, Width: 8, Height: 8, Columns: 2, First: 'A'}
	configs := make([]RingConfig, 7)
	starts := []float64{-28, 52, 132, 212, 292, 372, 452}
	baseline := make([]*Ring, len(configs))
	for index := range configs {
		configs[index] = RingConfig{Text: "ABABABAB", Font: font, Viewport: 32, Speed: 2}
		var err error
		baseline[index], err = NewRing(configs[index])
		if err != nil {
			t.Fatal(err)
		}
	}
	lanes, err := NewRingLanes(RingLanesConfig{
		Rings: configs, Y: starts, ShiftEvery: 128, ShiftFirst: 130,
		ShiftVelocity: []float64{2},
		ShiftUpper:    &motion.WrapLimit{Boundary: 540, Restart: -28, Inclusive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	legacyY := append([]float64(nil), starts...)
	loop := 0
	for frame := 0; frame < 6000; frame++ {
		lanes.Step()
		for index, old := range baseline {
			old.Step()
			if got := lanes.Y(index); got != legacyY[index] {
				t.Fatalf("frame %d lane %d y = %v, want %v", frame, index, got, legacyY[index])
			}
			for letter, actual := range lanes.rings[index].letters {
				if actual != old.letters[letter] {
					t.Fatalf("frame %d lane %d letter %d = %+v, want %+v", frame, index, letter, actual, old.letters[letter])
				}
			}
		}
		if loop >= 128 {
			for index := range legacyY {
				legacyY[index] += 2
				if legacyY[index] >= 540 {
					legacyY[index] = -28
				}
			}
			loop = 0
		}
		loop++
	}
}

func TestRingLanesAllowDifferentFontsAndRejectInvalidCadence(t *testing.T) {
	a, b := ebiten.NewImage(16, 8), ebiten.NewImage(24, 8)
	defer a.Deallocate()
	defer b.Deallocate()
	configs := []RingConfig{
		{Text: "AAAA", Font: BitmapGrid{Image: a, Width: 8, Height: 8, Columns: 2, First: 'A'}, Viewport: 32, Speed: 2},
		{Text: "AAAA", Font: BitmapGrid{Image: b, Width: 12, Height: 8, Columns: 2, First: 'A'}, Viewport: 32, Speed: 3},
	}
	lanes, err := NewRingLanes(RingLanesConfig{Rings: configs, Y: []float64{0, 10}})
	if err != nil {
		t.Fatal(err)
	}
	if lanes.Len() != 2 || lanes.Y(1) != 10 || lanes.rings[0].config.Font.Width == lanes.rings[1].config.Font.Width {
		t.Fatal("independent font metrics were lost")
	}
	if _, err := NewRingLanes(RingLanesConfig{Rings: configs, Y: []float64{0}}); err == nil {
		t.Fatal("accepted missing lane position")
	}
	if _, err := NewRingLanes(RingLanesConfig{Rings: configs, Y: []float64{0, 10}, ShiftFirst: 1}); err == nil {
		t.Fatal("accepted cadence options without a shift period")
	}
}

func TestRingLanesUseTheScrollingFacade(t *testing.T) {
	image := ebiten.NewImage(16, 8)
	defer image.Deallocate()
	config := RingLanesConfig{
		Rings: []RingConfig{{Text: "ABAB", Font: BitmapGrid{Image: image, Width: 8, Height: 8, Columns: 2, First: 'A'}, Viewport: 32, Speed: 2}},
		Y:     []float64{10},
	}
	scroll, err := New(Config{RingLanes: &config})
	if err != nil {
		t.Fatal(err)
	}
	if err := scroll.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if got := scroll.backend.(*RingLanes).rings[0].letters[0].X; got != 38 {
		t.Fatalf("first ring letter after update = %v, want 38", got)
	}
	if _, err := New(Config{RingLanes: &config, Recycled: &RecycledConfig{}}); err == nil {
		t.Fatal("accepted two transports")
	}
}
