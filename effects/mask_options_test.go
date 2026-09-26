package effects

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestMaskWithPlacesAlphaAndCropsTopBeforeOutput(t *testing.T) {
	content := ebiten.NewImage(2, 4)
	alpha := ebiten.NewImage(1, 1)
	defer content.Deallocate()
	defer alpha.Deallocate()
	content.Fill(color.RGBA{R: 255, A: 255})
	alpha.Fill(color.White)
	var updateOrder []int
	mask, err := NewMaskWith(MaskConfig{
		Content: kit.Func{OnUpdate: func(kit.Frame) error { updateOrder = append(updateOrder, 1); return nil },
			OnDraw: func(dst *ebiten.Image) { dst.DrawImage(content, nil) }},
		Alpha: kit.Func{OnUpdate: func(kit.Frame) error { updateOrder = append(updateOrder, 2); return nil },
			OnDraw: func(dst *ebiten.Image) { dst.DrawImage(alpha, nil) }},
		Width: 2, Height: 4, Blend: ebiten.BlendDestinationIn,
		MaskY: 1, OutputX: 2, OutputY: 3, ClearTop: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mask.Close()
	if err := mask.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if len(updateOrder) != 2 || updateOrder[0] != 1 || updateOrder[1] != 2 {
		t.Fatalf("mask update order %v", updateOrder)
	}
	dst := ebiten.NewImage(6, 8)
	defer dst.Deallocate()
	mask.Draw(dst)
	if got := color.RGBAModel.Convert(dst.At(2, 4)).(color.RGBA); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("masked output pixel %+v, want red", got)
	}
	if got := color.RGBAModel.Convert(dst.At(2, 3)).(color.RGBA); got.A != 0 {
		t.Fatalf("top clear did not remove row: %+v", got)
	}
}
