package composite

import (
	"image/color"
	"math"
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

func TestSurfaceLayerMovesImagePassWithoutRebuilding(t *testing.T) {
	image := ebiten.NewImage(1, 1)
	defer image.Deallocate()
	layer, err := NewSurfaceLayer(SurfaceLayerConfig{
		Width: 16, Height: 8, Background: color.Black,
		Passes:  []SurfaceImagePass{{Image: image, X: 2, Y: 3}},
		Outputs: []SurfaceOutput{{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer layer.Close()
	canvas := layer.Canvas()
	if err := layer.SetPassPosition(0, 7.5, -1); err != nil {
		t.Fatal(err)
	}
	if layer.Canvas() != canvas || layer.passes[0].config.X != 7.5 || layer.passes[0].config.Y != -1 {
		t.Fatal("moving a pass rebuilt the surface or lost its position")
	}
	if err := layer.SetPassPosition(1, 0, 0); err == nil {
		t.Fatal("accepted a missing pass")
	}
	if err := layer.SetPassPosition(0, math.NaN(), 0); err == nil {
		t.Fatal("accepted a nonfinite pass position")
	}
	if err := layer.SetPassTransform(0, 4, 5, .75, 2, .2); err != nil {
		t.Fatal(err)
	}
	if err := layer.SetPassFilter(0, ebiten.FilterLinear); err != nil {
		t.Fatal(err)
	}
	if got := layer.passes[0].config; layer.Canvas() != canvas || got.X != 4 || got.Y != 5 ||
		got.ScaleX != .75 || got.ScaleY != 2 || got.Angle != .2 || got.Filter != ebiten.FilterLinear {
		t.Fatalf("dynamic pass transform changed the surface or lost its pose: %+v", got)
	}
	if err := layer.SetPassTransform(0, 0, 0, math.NaN(), 1, 0); err == nil {
		t.Fatal("accepted a nonfinite pass scale")
	}
}
