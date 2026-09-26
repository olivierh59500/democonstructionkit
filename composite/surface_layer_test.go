package composite

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestSurfaceLayerComposesSourcesRasterAndOutputCopies(t *testing.T) {
	red, blue := ebiten.NewImage(1, 1), ebiten.NewImage(1, 1)
	defer red.Deallocate()
	defer blue.Deallocate()
	red.Fill(color.RGBA{R: 255, A: 255})
	blue.Fill(color.RGBA{B: 255, A: 255})
	updates := 0
	layer, err := NewSurfaceLayer(SurfaceLayerConfig{
		Width: 1, Height: 1,
		Sources: []kit.Effect{kit.Func{
			OnUpdate: func(kit.Frame) error { updates++; return nil },
			OnDraw:   func(dst *ebiten.Image) { dst.DrawImage(red, nil) },
		}},
		Passes:  []SurfaceImagePass{{Image: blue, Blend: ebiten.BlendSourceAtop}},
		Outputs: []SurfaceOutput{{X: 1, Y: 2}, {X: 3, Y: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer layer.Close()
	if err := layer.Update(kit.Frame{}); err != nil || updates != 1 {
		t.Fatalf("source update count %d, error %v", updates, err)
	}
	dst := ebiten.NewImage(5, 4)
	defer dst.Deallocate()
	layer.Draw(dst)
	for _, x := range []int{1, 3} {
		if got := color.RGBAModel.Convert(dst.At(x, 2)).(color.RGBA); got != (color.RGBA{B: 255, A: 255}) {
			t.Fatalf("output copy at %d is %+v, want blue", x, got)
		}
	}
	if got := color.RGBAModel.Convert(dst.At(2, 2)).(color.RGBA); got.A != 0 {
		t.Fatalf("gap between copies is opaque: %+v", got)
	}
}
