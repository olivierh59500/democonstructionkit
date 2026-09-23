package scrolling

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestRowBandsKeepsCursorControlsOutsideDraw(t *testing.T) {
	atlas := scanlineTestAtlas(t)
	effect, err := NewRowBands(RowBandsConfig{
		Font: atlas, Text: "A\\B", Advance: 10,
		FallbackRect: image.Rect(0, 0, 10, 2),
		WorkWidth:    20, Height: 2, TextSpeed: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	if got := effect.CursorRune(); got != 'A' {
		t.Fatalf("initial cursor = %q", got)
	}
	for _, want := range []rune{'\\', 'B', 'A'} {
		if err := effect.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		if got := effect.CursorRune(); got != want {
			t.Fatalf("cursor = %q, want %q", got, want)
		}
		before := effect.position
		canvas := ebiten.NewImage(20, 2)
		effect.Draw(canvas)
		effect.Draw(canvas)
		if effect.position != before || effect.CursorRune() != want {
			t.Fatal("repeated Draw advanced the text or its control cue")
		}
	}
}

func TestRowColumnVariableSpeedPreservesSeparateClocks(t *testing.T) {
	atlas := scanlineTestAtlas(t)
	effect, err := NewRowColumn(RowColumnConfig{
		Font: atlas, Text: "AB", Advance: 10,
		ViewportWidth: 10, Height: 2, WorkWidth: 20,
		RowHeight: 1, ColumnWidth: 5, Wave: []float64{0}, RowPhaseStep: 1,
		TextSpeed: 2, TextRestartX: 10, InitialMultiplier: 1.4,
		ColumnAmplitude: 1, ColumnSpeed: .1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	if err := effect.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if math.Abs(effect.x+2.8) > 1e-12 || math.Abs(effect.columnPhase-.14) > 1e-12 || effect.rowPhase != 1 {
		t.Fatalf("first clocks: text=%v column=%v row=%d", effect.x, effect.columnPhase, effect.rowPhase)
	}
	if err := effect.SetSpeedMultiplier(2); err != nil {
		t.Fatal(err)
	}
	if err := effect.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if math.Abs(effect.x+6.8) > 1e-12 || math.Abs(effect.columnPhase-.34) > 1e-12 || effect.rowPhase != 2 {
		t.Fatalf("changed clocks: text=%v column=%v row=%d", effect.x, effect.columnPhase, effect.rowPhase)
	}
	x, column, row := effect.x, effect.columnPhase, effect.rowPhase
	canvas := ebiten.NewImage(10, 4)
	effect.Draw(canvas)
	effect.Draw(canvas)
	if effect.x != x || effect.columnPhase != column || effect.rowPhase != row {
		t.Fatal("repeated Draw advanced a transport clock")
	}
	if err := effect.SetSpeedMultiplier(-1); err == nil {
		t.Fatal("negative multiplier was accepted")
	}
}
