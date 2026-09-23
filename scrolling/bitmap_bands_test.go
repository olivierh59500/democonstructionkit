package scrolling

import (
	"image"
	"math"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestBitmapBandsKeepLongMessagesOffTheGPU(t *testing.T) {
	atlas := ebiten.NewImage(24, 4)
	defer atlas.Deallocate()
	font, _ := (BitmapSpec{Width: 8, Height: 4, Order: "ABC"}).Grid(atlas, ebiten.FilterLinear)
	c := BitmapBandsConfig{Font: font, Text: strings.Repeat("ABC", 40000), Width: 64, Height: 16, Repeat: true, UseTicks: true, Lanes: []BitmapLane{{Speed: 3}, {Y: 8, Speed: 7}}}
	effect, err := New(Config{Bands: &c})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	bands := effect.backend.(*BitmapBands)
	if bands.text.Width() != 960000 || bands.surface.Bounds() != image.Rect(0, 0, 64, 16) {
		t.Fatalf("message changed GPU dimensions: %v", bands.surface.Bounds())
	}
	c.Lanes[0].Speed = 99
	if bands.config.Lanes[0].Speed != 3 {
		t.Fatal("band retained mutable config slice")
	}
	if err := effect.Update(kit.Frame{Tick: 123456789}); err != nil {
		t.Fatal(err)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = effect.Update(kit.Frame{Tick: 123456789}) }); allocations != 0 {
		t.Fatalf("band Update allocated %g times", allocations)
	}
}
func TestBitmapBandsRejectDenseWorkAndInvalidClock(t *testing.T) {
	atlas := ebiten.NewImage(8, 4)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 8, Height: 4, Order: "A"}).Grid(atlas, ebiten.FilterNearest)
	c := BitmapBandsConfig{Font: grid, Text: "A", Width: 64, Height: 16, Advance: 1e-100, Lanes: []BitmapLane{{Speed: 1}}}
	if _, err := NewBitmapBands(c); err == nil {
		t.Fatal("unbounded glyph density accepted")
	}
	c.Advance = 8
	bands, err := NewBitmapBands(c)
	if err != nil {
		t.Fatal(err)
	}
	defer bands.Close()
	if err := bands.Update(kit.Frame{Time: math.Inf(1)}); err == nil {
		t.Fatal("nonfinite time accepted")
	}
	if _, err := New(Config{Bands: &c, Text: "A"}); err == nil {
		t.Fatal("conflicting regular text transport accepted")
	}
}
