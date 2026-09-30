//go:build dck_gpu_rendercheck

package sprites

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicFieldBatchedImagesAndBorrowedLifetimeGPU(t *testing.T) {
	testutil.RequireGPU(t)
	image := ebiten.NewImage(2, 2)
	defer image.Deallocate()
	image.Fill(color.NRGBA{R: 230, G: 160, A: 255})
	f, err := NewHarmonicField(HarmonicFieldConfig{
		Count: 2, Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: 3, Y: 4}, Spacing: motion.Point{X: 4}},
		Style: FieldStyle{Image: image, Appearance: FieldAppearance{AnchorX: .5, AnchorY: .5}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dst := ebiten.NewImage(12, 8)
	defer dst.Deallocate()
	f.Draw(dst)
	pixels := make([]byte, 12*8*4)
	dst.ReadPixels(pixels)
	for y := 0; y < 8; y++ {
		for x := 0; x < 12; x++ {
			inside := y >= 3 && y < 5 && ((x >= 2 && x < 4) || (x >= 6 && x < 8))
			i := (y*12 + x) * 4
			if inside && (pixels[i] != 230 || pixels[i+1] != 160 || pixels[i+3] != 255) || !inside && pixels[i+3] != 0 {
				t.Fatalf("sprite coverage at (%d,%d): %v", x, y, pixels[i:i+4])
			}
		}
	}
	f.Close()
	dst.Clear()
	dst.DrawImage(image, nil)
	dst.ReadPixels(pixels)
	if pixels[0] != 230 || pixels[3] != 255 {
		t.Fatal("closing field deallocated borrowed artwork")
	}
}
