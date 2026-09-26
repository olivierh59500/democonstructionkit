package composite

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestRasterOverlayWrapPolicies(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		start, velocity, boundary float64
		restart                   float64
		inclusive                 bool
		want                      float64
	}{
		{"negative inclusive", -175, -2, -177, 0, true, 0},
		{"negative strict at edge", -71.5, -.5, -72, 0, false, -72},
		{"negative strict beyond edge", -72, -.5, -72, 0, false, 0},
		{"positive inclusive", -54, 1, -53, -221, true, -221},
		{"positive strict at edge", -54, 1, -53, -221, false, -53},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := rasterNext(tc.start, tc.velocity, &RasterWrap{Boundary: tc.boundary, Restart: tc.restart, Inclusive: tc.inclusive})
			if got != tc.want {
				t.Fatalf("next phase = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRasterOverlayDrawsIndependentAuthoredCopies(t *testing.T) {
	texture := ebiten.NewImage(1, 1)
	defer texture.Deallocate()
	texture.Fill(color.RGBA{R: 255, A: 255})
	copies := []RasterCopy{{Y: -2}, {Y: 0}, {Y: 2}}
	overlay, err := NewRasterOverlay(RasterOverlayConfig{Image: texture,
		X: 1, Y: 3, ScaleX: 1, ScaleY: 1, Alpha: 1, Copies: copies})
	if err != nil {
		t.Fatal(err)
	}
	copies[1].Y = 99 // Construction keeps a private placement recipe.
	dst := ebiten.NewImage(4, 7)
	defer dst.Deallocate()
	overlay.Draw(dst)
	for _, y := range []int{1, 3, 5} {
		if got := color.RGBAModel.Convert(dst.At(1, y)).(color.RGBA); got != (color.RGBA{R: 255, A: 255}) {
			t.Fatalf("copy at row %d has pixel %+v", y, got)
		}
	}
}

func TestRasterOverlayPhaseAndValidation(t *testing.T) {
	texture := ebiten.NewImage(32, 180)
	t.Cleanup(texture.Deallocate)
	wrap := &RasterWrap{Boundary: -177, Restart: 0, Inclusive: true}
	r, err := NewRasterOverlay(RasterOverlayConfig{Image: texture, Source: image.Rect(0, 0, 16, 180), ScaleX: 85, ScaleY: 1,
		Alpha: 1, VelocityY: -2, WrapY: wrap, Blend: ebiten.BlendSourceAtop})
	if err != nil {
		t.Fatal(err)
	}
	wrap.Restart = 5 // Construction freezes authored wrap parameters.
	if err := r.SetPhase(0, -175); err != nil {
		t.Fatal(err)
	}
	if err := r.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if _, y := r.Phase(); y != 0 {
		t.Fatalf("wrapped phase = %v, want 0", y)
	}
	if err := r.SetVelocity(0, math.NaN()); err == nil {
		t.Fatal("accepted nonfinite raster speed")
	}
	if got := testing.AllocsPerRun(30, r.Step); got != 0 {
		t.Fatalf("step allocations = %v, want 0", got)
	}
	for _, c := range []RasterOverlayConfig{{}, {Image: texture},
		{Image: texture, ScaleX: 1, ScaleY: 1, Alpha: 1, Source: image.Rect(0, 0, 100, 1)},
		{Image: texture, ScaleX: 1, ScaleY: 1, Alpha: 1, WrapY: &RasterWrap{Boundary: math.Inf(1)}},
	} {
		if _, err := NewRasterOverlay(c); err == nil {
			t.Fatalf("accepted invalid raster overlay %+v", c)
		}
	}
}
