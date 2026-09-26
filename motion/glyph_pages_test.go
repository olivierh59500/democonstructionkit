package motion

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestGlyphPageCycleCompletionBarriersAndDepthOrder(t *testing.T) {
	cycle, err := NewGlyphPageCycle(GlyphPageConfig{
		Pages: []string{"ABCD", "WXYZ"}, DelayPatterns: [][]int{{0, 1, 2, 3}, {3, 2, 1, 0}}, Columns: 2,
		Origin: GlyphPoint{X: 0, Y: 0, Z: 1}, Center: GlyphPoint{X: 10, Y: 20, Z: .1},
		ColumnStep: 10, RowStep: 10, EnterDurationMS: 100, ExitDurationMS: 50, DelayStepMS: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cycle.UpdateAt(50); err != nil {
		t.Fatal(err)
	}
	if got := cycle.Glyph(0).Position; got != (GlyphPoint{X: 5, Y: 10, Z: .55}) {
		t.Fatalf("first glyph midpoint %+v", got)
	}
	if cycle.Phase() != GlyphPageEntering || cycle.PatternIndex() != 0 {
		t.Fatal("entry barrier crossed before the last delay completed")
	}
	if err := cycle.UpdateAt(130); err != nil {
		t.Fatal(err)
	}
	if cycle.Phase() != GlyphPageExiting || cycle.PatternIndex() != 1 || cycle.PageIndex() != 0 {
		t.Fatalf("entry barrier state: phase=%d pattern=%d page=%d", cycle.Phase(), cycle.PatternIndex(), cycle.PageIndex())
	}
	if got := cycle.Glyph(3).Position; got != (GlyphPoint{X: 10, Y: 10, Z: 1}) {
		t.Fatalf("exit started from a discontinuous pose: %+v", got)
	}
	if err := cycle.UpdateAt(210); err != nil {
		t.Fatal(err)
	}
	if cycle.Phase() != GlyphPageEntering || cycle.PatternIndex() != 0 || cycle.PageIndex() != 1 || cycle.Glyph(0).Rune != 'W' {
		t.Fatal("exit barrier did not rotate page and delay pattern")
	}
	for i := 0; i < cycle.Count(); i++ {
		if cycle.Glyph(i).Position != (GlyphPoint{X: 10, Y: 20, Z: .1}) {
			t.Fatalf("page %d entered from wrong center", i)
		}
	}
	cycle.poses[0].Position.Z = 2
	cycle.poses[1].Position.Z = 1
	cycle.poses[2].Position.Z = 1
	cycle.poses[3].Position.Z = 0
	if got := cycle.DrawOrder(); !reflect.DeepEqual(got, []int{3, 1, 2, 0}) {
		t.Fatalf("depth order %v", got)
	}
	if err := cycle.SetPageAt(0, 1, 500); err != nil || cycle.PageIndex() != 0 || cycle.PatternIndex() != 1 {
		t.Fatalf("timeline page cue failed: %v", err)
	}
	if got := testing.AllocsPerRun(100, func() { _ = cycle.UpdateAt(501); cycle.DrawOrder() }); got != 0 {
		t.Fatalf("update and sort allocate %.2f objects per frame", got)
	}
}

func TestGlyphPageCycleRejectsInvalidInputs(t *testing.T) {
	base := GlyphPageConfig{Pages: []string{"AB"}, DelayPatterns: [][]int{{0, 1}}, Columns: 2,
		EnterDurationMS: 100, ExitDurationMS: 50, DelayStepMS: 10}
	bad := base
	bad.Pages = []string{"AB", "X"}
	if _, err := NewGlyphPageCycle(bad); err == nil {
		t.Fatal("accepted unequal page lengths")
	}
	bad = base
	bad.DelayPatterns = [][]int{{0, -1}}
	if _, err := NewGlyphPageCycle(bad); err == nil {
		t.Fatal("accepted negative delay")
	}
	bad = base
	bad.DelayStepMS = math.NaN()
	if _, err := NewGlyphPageCycle(bad); err == nil {
		t.Fatal("accepted nonfinite timing")
	}
	cycle, err := NewGlyphPageCycle(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := cycle.SetPageAt(9, 0, 0); err == nil {
		t.Fatal("accepted invalid timeline page")
	}
	if err := cycle.UpdateAt(math.Inf(1)); err == nil {
		t.Fatal("accepted nonfinite clock")
	}
}

func TestGlyphPageElasticEasesMatchAuthoredEquations(t *testing.T) {
	for _, progress := range []float64{0, .123, .5, .987, 1} {
		if progress == 0 || progress == 1 {
			continue
		}
		out := math.Exp2(-10*progress)*math.Sin((progress-.075)*2*math.Pi/.3) + 1
		in := -math.Exp2(10*(progress-1)) * math.Sin((progress-1-.075)*2*math.Pi/.3)
		if math.Abs(ElasticOut(progress)-out) > 1e-14 || math.Abs(ElasticIn(progress)-in) > 1e-14 {
			t.Fatalf("elastic easing differs from authored page tween at %.3f", progress)
		}
	}
}

func BenchmarkGlyphPageCycle160(b *testing.B) {
	pattern, err := SpiralGlyphDelays(20, 8)
	if err != nil {
		b.Fatal(err)
	}
	cycle, err := NewGlyphPageCycle(GlyphPageConfig{
		Pages: []string{strings.Repeat("A", 160), strings.Repeat("B", 160)},
		DelayPatterns: [][]int{pattern}, Columns: 20,
		Origin: GlyphPoint{X: 16, Y: 160, Z: 1}, Center: GlyphPoint{X: 320, Y: 240, Z: 1e-9},
		ColumnStep: 32, RowStep: 32, EnterDurationMS: 2000, ExitDurationMS: 1000,
		DelayStepMS: 40, EnterEase: ElasticOut, ExitEase: ElasticIn,
	})
	if err != nil {
		b.Fatal(err)
	}
	nowMS := 0.0
	b.ReportAllocs()
	for b.Loop() {
		nowMS += 16.0
		if err := cycle.UpdateAt(nowMS); err != nil {
			b.Fatal(err)
		}
		cycle.DrawOrder()
	}
}
