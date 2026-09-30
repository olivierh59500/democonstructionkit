//go:build dck_gpu_rendercheck

package composite

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicBandsFilledOutlineAndBorrowedMaterialGPU(t *testing.T) {
	testutil.RequireGPU(t)
	white := ebiten.NewImage(1, 1)
	defer white.Deallocate()
	white.Fill(color.White)
	paint := color.NRGBA{R: 200, G: 100, B: 50, A: 255}
	b, err := NewHarmonicBands(HarmonicBandsConfig{
		LeftX: 2, RightX: 18, Thickness: 4, Colors: []color.NRGBA{paint}, White: white,
		Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: 6, Y: 6}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	dst := ebiten.NewImage(24, 16)
	defer dst.Deallocate()
	pixels := make([]byte, 24*16*4)
	b.DrawAt(dst, 4, 3)
	dst.ReadPixels(pixels)
	for y := 0; y < 16; y++ {
		for x := 0; x < 24; x++ {
			want := color.NRGBA{}
			if x >= 6 && x < 22 && y >= 9 && y < 13 {
				want = paint
			}
			i := (y*24 + x) * 4
			if pixels[i] != want.R || pixels[i+1] != want.G || pixels[i+2] != want.B || pixels[i+3] != want.A {
				t.Fatalf("filled band at (%d,%d): got %v, want %v", x, y, pixels[i:i+4], want)
			}
		}
	}
	dst.Clear()
	b.DrawOutline(dst, 2)
	dst.ReadPixels(pixels)
	if pixels[(6*24+10)*4] != paint.R || pixels[(8*24+10)*4+3] != 0 {
		t.Fatal("outline lost its border or filled the strip interior")
	}
	b.Close()
	dst.Clear()
	dst.DrawImage(white, nil)
	dst.ReadPixels(pixels)
	if pixels[0] != 255 || pixels[3] != 255 {
		t.Fatal("closing bands deallocated the borrowed white material")
	}
}
