//go:build dck_gpu_rendercheck

package sprites

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestImageSlotsGPURetainsSourceOrderCropsAndAtomicSelections(t *testing.T) {
	a, b := ebiten.NewImage(4, 3), ebiten.NewImage(2, 2)
	defer a.Deallocate()
	defer b.Deallocate()
	a.Fill(color.NRGBA{R: 255, A: 255})
	b.Fill(color.NRGBA{B: 255, A: 255})
	slots := []ImageSlot{{Image: 0, Source: image.Rect(1, 0, 4, 2), X: 1, Y: 1}, {Image: 1, X: 2, Y: 1}}
	s, err := NewImageSlots(ImageSlotsConfig{Images: []*ebiten.Image{a, b}, Slots: slots, MaxSlots: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	slots[0].X = 99
	dst := ebiten.NewImage(8, 6)
	defer dst.Deallocate()
	s.Draw(dst)
	pixels := make([]byte, 8*6*4)
	dst.ReadPixels(pixels)
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			want := [4]byte{}
			if y >= 1 && y < 3 {
				if x >= 1 && x < 4 {
					want = [4]byte{255, 0, 0, 255}
				}
				if x >= 2 && x < 4 {
					want = [4]byte{0, 0, 255, 255}
				}
			}
			at := (y*8 + x) * 4
			for c, v := range want {
				if pixels[at+c] != v {
					t.Fatal("slotorder/crop pixel", x, y, pixels[at:at+4], want)
				}
			}
		}
	}
	if s.SetSlots([]ImageSlot{{Image: 99}}) == nil {
		t.Fatal("invalid selection accepted")
	}
	dst.Clear()
	s.Draw(dst)
	again := make([]byte, len(pixels))
	dst.ReadPixels(again)
	for i, v := range pixels {
		if again[i] != v {
			t.Fatal("invalid window modifieddraw")
		}
	}
	calls := 0
	s.config.Select = func(_ kit.Frame, out []ImageSlot) (int, error) {
		calls++
		out[0] = ImageSlot{Image: 1, X: 5, Y: 4}
		return 1, nil
	}
	if err := s.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	s.Draw(dst)
	s.Draw(dst)
	if calls != 1 {
		t.Fatal("draw advancedselection")
	}
	s.Close()
	a.Fill(color.White)
	dst.Clear()
	dst.DrawImage(a, nil)
	dst.ReadPixels(pixels)
	if pixels[0] != 255 || pixels[3] != 255 {
		t.Fatal("close destroyed borrowedsource")
	}
}
