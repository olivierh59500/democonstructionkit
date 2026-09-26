package presets

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

func TestNonamenoBottomScrollRetainsStrictRestartAndOptionalContinuousLoop(t *testing.T) {
	atlas := ebiten.NewImage(16, 8)
	defer atlas.Deallocate()
	metrics, err := font.NewGrid(font.Grid{Bounds: atlas.Bounds(), Cell: image.Point{X: 8, Y: 8}, Columns: 2, Order: "AB"})
	if err != nil {
		t.Fatal(err)
	}
	face := scrolling.Face{Atlas: atlas, Metrics: metrics}
	options := NonamenoScrollOptions{Width: 640, BaselineY: 442, TicksPerSecond: 60, PixelsPerTick: 1, Gap: 641}
	config, err := NonamenoBottomScroll("AB", face, options)
	if err != nil {
		t.Fatal(err)
	}
	scroll, err := scrolling.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	if got := scroll.StateAt(0).X; math.Abs(got-640) > 1e-9 {
		t.Fatalf("initial x %v", got)
	}
	if got := scroll.StateAt(1.0 / 60).X; math.Abs(got-639) > 1e-9 {
		t.Fatalf("first tick x %v", got)
	}
	if got := scroll.StateAt(656.0 / 60).X; math.Abs(got+16) > 1e-9 {
		t.Fatalf("last outgoing x %v", got)
	}
	if got := scroll.StateAt(657.0 / 60).X; math.Abs(got-640) > 1e-9 {
		t.Fatalf("strict restart x %v", got)
	}
	options.Gap = 0
	config, err = NonamenoBottomScroll("AB", face, options)
	if err != nil {
		t.Fatal(err)
	}
	continuous, err := scrolling.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer continuous.Close()
	if got := continuous.StateAt(16.0 / 60).X; math.Abs(got-640) > 1e-9 {
		t.Fatalf("continuous copy boundary x %v", got)
	}
}
