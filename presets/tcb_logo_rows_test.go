package presets

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestTCBLogoRowProfileKeepsStrictCounterWrap(t *testing.T) {
	values, err := motion.CompileWaveTable(TCBLogoWaveSections()...)
	if err != nil {
		t.Fatal(err)
	}
	image := ebiten.NewImage(303, 32)
	defer image.Deallocate()
	config := TCBLogoRowProfile(values, 303)
	rows, err := composite.NewProfileImage(image, config)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !config.Batch || config.BaseX != 159.5 || config.BaseY != 96 || config.RowStep != 1 || config.PhaseWrap != len(values)-80 {
		t.Fatalf("unexpected TCB logo placement %+v", config)
	}
	rows.SetPhase(40)
	if rows.RowOffset(0) != values[40] || rows.RowOffset(1) != values[41] || rows.RowOffset(0) == rows.RowOffset(1) {
		t.Fatal("adjacent logo rows lost their spatial wave phase")
	}
	rows.SetPhase(0)
	counter := 0
	for tick := 0; tick < 4000; tick++ {
		counter++
		if counter > len(values)-80 {
			counter = 0
		}
		rows.Advance()
		if rows.Phase() != counter {
			t.Fatalf("tick %d phase %d, want %d", tick, rows.Phase(), counter)
		}
	}
}
