//go:build dck_gpu_rendercheck

package sprites

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestProjectedObjectGPUNearSpritesOccludeFarSprites(t *testing.T) {
	near, far := ebiten.NewImage(4, 4), ebiten.NewImage(4, 4)
	defer near.Deallocate()
	defer far.Deallocate()
	near.Fill(color.NRGBA{R: 255, A: 255})
	far.Fill(color.NRGBA{B: 255, A: 255})
	for _, batch := range []bool{false, true} {
		object, err := NewProjectedObject(ProjectedObjectConfig{
			// The input deliberately lists the near ball first. Correct
			// depth sorting must still put it in front at this crossing.
			Points: []Point{{Z: 0, Image: 0}, {Z: 100, Image: 1}},
			Scale:  1, Focal: 100, CenterX: 4, CenterY: 4, Batch: batch,
		})
		if err != nil {
			t.Fatal(err)
		}
		dst := ebiten.NewImage(8, 8)
		object.Draw(dst, []*ebiten.Image{near, far})
		pixels := make([]byte, 8*8*4)
		dst.ReadPixels(pixels)
		dst.Deallocate()
		at := (4*8 + 4) * 4
		if pixels[at] != 255 || pixels[at+1] != 0 || pixels[at+2] != 0 || pixels[at+3] != 255 {
			t.Fatalf("near ball is obscured (batch %t): %v", batch, pixels[at:at+4])
		}
	}
}
