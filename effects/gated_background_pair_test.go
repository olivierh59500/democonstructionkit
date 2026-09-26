package effects

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestGatedBackgroundPairDrawsSecondLayerAboveFirst(t *testing.T) {
	red, blue := ebiten.NewImage(2, 2), ebiten.NewImage(2, 2)
	defer red.Deallocate()
	defer blue.Deallocate()
	red.Fill(color.RGBA{R: 255, A: 255})
	blue.Fill(color.RGBA{B: 255, A: 255})
	background := composite.BackgroundConfig{CopiesX: 1, CopiesY: 1}
	pair, err := NewGatedBackgroundPair(GatedBackgroundPairConfig{
		Images:      [2]*ebiten.Image{red, blue},
		Backgrounds: [2]composite.BackgroundConfig{background, background},
		Motion: motion.GatedBackgroundPairConfig{GateOpen: 1, GateReset: 2,
			FirstX:     motion.ThresholdAxis{Lower: -1, Upper: 1},
			FirstY:     motion.ThresholdAxis{Lower: -1, Upper: 1},
			SecondY:    motion.ThresholdAxis{Lower: -1, Upper: 1},
			SecondXMin: -1, SecondXMax: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(2, 2)
	defer dst.Deallocate()
	pair.Draw(dst)
	if got := color.RGBAModel.Convert(dst.At(0, 0)).(color.RGBA); got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("paired background order produced %+v", got)
	}
}
