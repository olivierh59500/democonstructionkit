//go:build dck_gpu_rendercheck

package sprites

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func TestFieldOutlineMetricsRotationAndMirroringGPU(t *testing.T) {
	testutil.RequireGPU(t)
	r := NewFieldRenderer(1)
	defer r.Close()
	art := ebiten.NewImage(12, 10)
	defer art.Deallocate()
	art.Fill(color.NRGBA{R: 255, A: 255})
	// The frame's 6x4 dimensions, not the atlas dimensions, define the border.
	style := FieldStyle{Image: art, Frames: []image.Rectangle{image.Rect(3, 2, 9, 6)},
		Appearance: FieldAppearance{AnchorX: .5, AnchorY: .5}, Outline: &FieldOutline{}}
	dst := ebiten.NewImage(24, 20)
	defer dst.Deallocate()
	pixels := make([]byte, 24*20*4)
	for _, appearance := range []FieldAppearance{
		{AnchorX: .5, AnchorY: .5},
		{AnchorX: .5, AnchorY: .5, ScaleX: -1},
		{AnchorX: .5, AnchorY: .5, Angle: math.Pi / 2},
	} {
		style.Appearance = appearance
		dst.Clear()
		r.Draw(dst, []FieldSample{{X: 12, Y: 10}}, style)
		dst.ReadPixels(pixels)
		left, right, top, bottom := 9, 15, 8, 12
		if appearance.Angle != 0 {
			left, right, top, bottom = 10, 14, 7, 13
		}
		for y := 0; y < 20; y++ {
			for x := 0; x < 24; x++ {
				inside := x >= left && x < right && y >= top && y < bottom
				border := inside && (x == left || x == right-1 || y == top || y == bottom-1)
				at := (y*24 + x) * 4
				if border && (pixels[at] != 255 || pixels[at+1] != 255 || pixels[at+2] != 255 || pixels[at+3] != 255) || !border && pixels[at+3] != 0 {
					t.Fatalf("outline %+v at (%d,%d): got %v, border=%t", appearance, x, y, pixels[at:at+4], border)
				}
			}
		}
	}
}

func TestFieldOutlineTranslucencyDepthAndSkinSwitchGPU(t *testing.T) {
	testutil.RequireGPU(t)
	r := NewFieldRenderer(1)
	defer r.Close()
	art := ebiten.NewImage(6, 6)
	defer art.Deallocate()
	art.Fill(color.NRGBA{G: 220, A: 255})
	style := FieldStyle{Image: art, Appearance: FieldAppearance{Width: 3, Height: 3}, ScaleByDepth: true,
		Outline:    &FieldOutline{Width: 1, Color: color.NRGBA{R: 200, G: 100, B: 50, A: 128}},
		DrawImages: true, VectorRects: true} // Outline selects its own material.
	dst := ebiten.NewImage(16, 16)
	defer dst.Deallocate()
	pixels := make([]byte, 16*16*4)
	samples := []FieldSample{{X: 4, Y: 4, Scale: 2}}
	r.Draw(dst, samples, style)
	dst.ReadPixels(pixels)
	for _, point := range [][2]int{{4, 4}, {6, 4}, {4, 6}, {9, 9}} {
		at := (point[1]*16 + point[0]) * 4
		for i, want := range []int{100, 50, 25, 128} {
			got := int(pixels[at+i])
			if got < want-1 || got > want+1 {
				t.Fatalf("border/corner at %v channel %d = %d, want about %d", point, i, got, want)
			}
		}
	}
	if pixels[(6*16+6)*4+3] != 0 {
		t.Fatal("outlined skin filled its interior")
	}
	style.Outline, style.VectorRects = nil, false
	dst.Clear()
	r.Draw(dst, samples, style)
	dst.ReadPixels(pixels)
	if pixels[(6*16+6)*4+1] != 220 || pixels[(6*16+6)*4+3] != 255 {
		t.Fatal("switching skins failed to restore the borrowed image")
	}
}
