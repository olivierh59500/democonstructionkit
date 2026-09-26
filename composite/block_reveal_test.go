package composite

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestBlockRevealUsesEditableGridOrderAndTimedHold(t *testing.T) {
	source := ebiten.NewImage(4, 4)
	defer source.Deallocate()
	source.Fill(color.RGBA{R: 255, A: 255})
	reveal, err := NewBlockReveal(BlockRevealConfig{Image: source,
		CellWidth: 2, CellHeight: 2, Order: []int{3, 1, 2, 0},
		Timing:     motion.SteppedRevealConfig{BlocksPerStep: 1, StepTicks: 2, DoneTick: 12},
		Background: color.RGBA{G: 255, A: 255}})
	if err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(4, 4)
	defer dst.Deallocate()
	for tick := 1; tick <= 2; tick++ {
		if err := reveal.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	if reveal.VisibleCount() != 1 || reveal.Done() {
		t.Fatal("first block was not revealed at the second tick")
	}
	reveal.Draw(dst)
	if got := color.RGBAModel.Convert(dst.At(3, 3)).(color.RGBA); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("selected block pixel %+v", got)
	}
	if got := color.RGBAModel.Convert(dst.At(0, 0)).(color.RGBA); got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("unrevealed background pixel %+v", got)
	}
	for tick := 3; tick <= 12; tick++ {
		if err := reveal.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	if !reveal.Done() || reveal.VisibleCount() != 4 {
		t.Fatal("complete image was not held until the authored handoff")
	}
}
