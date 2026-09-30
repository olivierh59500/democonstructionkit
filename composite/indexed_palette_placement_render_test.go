//go:build dck_gpu_rendercheck

package composite

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func TestIndexedPalettePlacedTintedSubimagesAndDefaultIsolationGPU(t *testing.T) {
	testutil.RequireGPU(t)
	parent := ebiten.NewImage(5, 3)
	defer parent.Deallocate()
	pixels := make([]byte, 5*3*4)
	pixels[(5+1)*4+3] = 255
	pixels[(5+2)*4], pixels[(5+2)*4+3] = 1, 255
	parent.WritePixels(pixels)
	source := parent.SubImage(image.Rect(1, 1, 3, 2)).(*ebiten.Image)
	colors := []color.NRGBA{{R: 200, G: 100, B: 40, A: 255}, {R: 80, G: 200, B: 120, A: 255}}
	p, err := NewIndexedPalette(IndexedPaletteConfig{Palette: colors, Channel: BitplaneRed, Scale: 255, Blend: ebiten.BlendCopy})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	dst := ebiten.NewImage(12, 8)
	defer dst.Deallocate()
	view := dst.SubImage(image.Rect(2, 1, 10, 7)).(*ebiten.Image)
	options := ebiten.DrawRectShaderOptions{Uniforms: map[string]any{"Palette": 0}, Images: [4]*ebiten.Image{dst}}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(3, 2)
	options.ColorScale.Scale(.5, 1, 0, 1)
	if err := p.DrawWith(view, source, options); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, 12*8*4)
	dst.ReadPixels(got)
	for y := 0; y < 8; y++ {
		for x := 0; x < 12; x++ {
			want := [4]byte{}
			if x >= 3 && x < 7 && y >= 2 && y < 4 {
				want = [4]byte{100, 100, 0, 255}
				if x >= 5 {
					want = [4]byte{40, 200, 0, 255}
				}
			}
			for component, value := range want {
				if got[(y*12+x)*4+component] != value {
					t.Fatalf("placed palette pixel %d,%d: got %v, want %v", x, y, got[(y*12+x)*4:(y*12+x+1)*4], want)
				}
			}
		}
	}
	plain := ebiten.NewImage(2, 1)
	defer plain.Deallocate()
	if err := p.Draw(plain, source); err != nil {
		t.Fatal(err)
	}
	indexedPalettePixels(t, plain, [][4]byte{{200, 100, 40, 255}, {80, 200, 120, 255}})
	if p.DrawWith(nil, source, options) == nil || p.DrawWith(source, source, options) == nil {
		t.Fatal("placed draw accepted an invalid destination")
	}
	p.Close()
	if p.DrawWith(plain, source, options) == nil {
		t.Fatal("placed draw accepted a closed shader")
	}
}
