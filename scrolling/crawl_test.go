package scrolling

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"math"
	"testing"
)

func TestCrawlSharedConstructorAndBoundedWindowTransport(t *testing.T) {
	atlas := ebiten.NewImage(64, 16)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 16, Height: 16, Order: "ABCD"}).Grid(atlas, ebiten.FilterNearest)
	projection, _ := composite.NewRowProjection([]composite.Row{{Source: composite.Region{Width: 64, Height: 1}, Width: 64, Height: 1}})
	cfg := CrawlConfig{Paragraph: BitmapParagraphConfig{Font: grid, Lines: []string{"A", "B", "C", "D"}, LineAdvance: 17}, VisibleLines: 2, Width: 64, Height: 34, ProjectionWidth: 64, ProjectionHeight: 34, Projection: projection, Output: composite.Region{Width: 64, Height: 34}, PixelsPerUpdate: 1}
	effect, err := New(Config{Crawl: &cfg})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	crawl := effect.backend.(*Crawl)
	line, offset := 0, 0
	for tick := 0; tick < 10000; tick++ {
		offset--
		if offset < -16 {
			offset = 0
			line++
			if line > 2 {
				line = 0
			}
		}
		if err := effect.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
			t.Fatal(err)
		}
		if crawl.line != line || crawl.offset != float64(offset) {
			t.Fatalf("tick %d: got %d,%g want %d,%d", tick, crawl.line, crawl.offset, line, offset)
		}
	}
	if err := crawl.SetPosition(2, -16); err != nil {
		t.Fatal(err)
	}
	crawl.Step()
	if crawl.line != 0 || crawl.offset != 0 {
		t.Fatal("final window failed to wrap")
	}
	for _, position := range []struct {
		line   int
		offset float64
	}{{-1, 0}, {3, 0}, {0, -17}, {0, math.NaN()}} {
		if crawl.SetPosition(position.line, position.offset) == nil {
			t.Fatal("invalid seek accepted")
		}
	}
	if _, err := New(Config{Crawl: &cfg, Text: "A"}); err == nil {
		t.Fatal("conflicting regular transport accepted")
	}
	if err := crawl.SetSpeed(0); err != nil {
		t.Fatal(err)
	}
	lineBefore, offsetBefore := crawl.line, crawl.offset
	crawl.Step()
	if crawl.line != lineBefore || crawl.offset != offsetBefore {
		t.Fatal("paused crawl moved")
	}
	if crawl.SetSpeed(-1) == nil {
		t.Fatal("negative crawl speed accepted")
	}
	if allocations := testing.AllocsPerRun(100, func() { crawl.Step() }); allocations != 0 {
		t.Fatalf("transport allocated %g times", allocations)
	}
}
