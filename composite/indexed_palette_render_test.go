//go:build dck_gpu_rendercheck

package composite

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func indexedPalettePixels(t *testing.T, img *ebiten.Image, want [][4]byte) {
	t.Helper()
	pixels := make([]byte, img.Bounds().Dx()*img.Bounds().Dy()*4)
	img.ReadPixels(pixels)
	if len(want)*4 != len(pixels) {
		t.Fatal("invalid expected image dimensions")
	}
	for i, pixel := range want {
		for component, value := range pixel {
			if pixels[i*4+component] != value {
				t.Fatalf("pixel %d component %d = %d, want %d (whole pixel %v)", i, component, pixels[i*4+component], value, pixels[i*4:i*4+4])
			}
		}
	}
}

func TestIndexedPaletteChannelEncodingAndClampingGPU(t *testing.T) {
	testutil.RequireGPU(t)
	colors := []color.NRGBA{{R: 10, A: 255}, {G: 20, A: 255}, {B: 30, A: 255}, {R: 40, G: 50, A: 255}}
	for _, channel := range []BitplaneChannel{BitplaneAlpha, BitplaneRed, BitplaneGreen, BitplaneBlue} {
		t.Run(string(rune('0'+channel)), func(t *testing.T) {
			source := ebiten.NewImage(4, 1)
			defer source.Deallocate()
			pixels := make([]byte, 16)
			component := 3
			if channel != BitplaneAlpha {
				component = int(channel) - 1
			}
			for x, encoded := range []byte{0, 1, 2, 255} {
				pixels[x*4+3] = 255
				pixels[x*4+component] = encoded
			}
			source.WritePixels(pixels)
			p, err := NewIndexedPalette(IndexedPaletteConfig{Palette: colors, Channel: channel, Scale: 255, Offset: -1, Blend: ebiten.BlendCopy})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			dst := ebiten.NewImage(4, 1)
			defer dst.Deallocate()
			if err := p.Draw(dst, source); err != nil {
				t.Fatal(err)
			}
			indexedPalettePixels(t, dst, [][4]byte{{10, 0, 0, 255}, {10, 0, 0, 255}, {0, 20, 0, 255}, {40, 50, 0, 255}})
		})
	}
}

func TestIndexedPaletteDefaultRangeAndLiveColorsGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := ebiten.NewImage(4, 1)
	defer source.Deallocate()
	source.WritePixels([]byte{0, 0, 0, 0, 0, 0, 0, 85, 0, 0, 0, 170, 0, 0, 0, 255})
	colors := []color.NRGBA{{R: 10, A: 255}, {G: 20, A: 255}, {B: 30, A: 255}, {R: 40, A: 255}}
	p, err := NewIndexedPalette(IndexedPaletteConfig{Palette: colors, Blend: ebiten.BlendCopy})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	// Caller storage stays independent of uploaded uniforms.
	colors[0] = color.NRGBA{R: 255, A: 255}
	dst := ebiten.NewImage(4, 1)
	defer dst.Deallocate()
	if err := p.Draw(dst, source); err != nil {
		t.Fatal(err)
	}
	indexedPalettePixels(t, dst, [][4]byte{{10, 0, 0, 255}, {0, 20, 0, 255}, {0, 0, 30, 255}, {40, 0, 0, 255}})
	if err := p.SetPalette(colors); err != nil {
		t.Fatal(err)
	}
	if err := p.Draw(dst, source); err != nil {
		t.Fatal(err)
	}
	indexedPalettePixels(t, dst, [][4]byte{{255, 0, 0, 255}, {0, 20, 0, 255}, {0, 0, 30, 255}, {40, 0, 0, 255}})
}

func TestIndexedPaletteSubImageAndIndependentAlphaGPU(t *testing.T) {
	testutil.RequireGPU(t)
	parent := ebiten.NewImage(5, 3)
	defer parent.Deallocate()
	pixels := make([]byte, 5*3*4)
	for x, alpha := range []byte{64, 128, 255} {
		at := (5 + x + 1) * 4
		pixels[at], pixels[at+3] = 1, alpha
	}
	parent.WritePixels(pixels)
	source := parent.SubImage(image.Rect(1, 1, 4, 2)).(*ebiten.Image)
	for _, sourceAlpha := range []bool{false, true} {
		p, err := NewIndexedPalette(IndexedPaletteConfig{Palette: []color.NRGBA{{}, {R: 255, G: 128, B: 64, A: 128}},
			Channel: BitplaneRed, Scale: 255, SourceAlpha: sourceAlpha, Blend: ebiten.BlendCopy})
		if err != nil {
			t.Fatal(err)
		}
		dst := ebiten.NewImage(3, 1)
		if err := p.Draw(dst, source); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, 12)
		dst.ReadPixels(got)
		for x, alpha := range []byte{64, 128, 255} {
			factor := 1.0
			if sourceAlpha {
				factor = float64(alpha) / 255
			}
			for c, base := range []float64{128, 128 * 128.0 / 255, 64 * 128.0 / 255, 128} {
				want := int(math.Round(base * factor))
				if value := int(got[x*4+c]); value < want-1 || value > want+1 {
					t.Fatalf("source alpha %t, pixel %d component %d = %d, want about %d", sourceAlpha, x, c, value, want)
				}
			}
		}
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		dst.Clear()
		dst.DrawImage(source, nil)
		indexedPalettePixels(t, dst, [][4]byte{{1, 0, 0, 64}, {1, 0, 0, 128}, {1, 0, 0, 255}})
		dst.Deallocate()
	}
}

func TestIndexedPaletteRejectsInvalidConfigAndClosedUseGPU(t *testing.T) {
	testutil.RequireGPU(t)
	valid := IndexedPaletteConfig{Palette: []color.NRGBA{{A: 255}}}
	for _, config := range []IndexedPaletteConfig{{}, {Palette: make([]color.NRGBA, 257)},
		{Palette: valid.Palette, Channel: BitplaneBlue + 1}, {Palette: valid.Palette, Scale: -1},
		{Palette: valid.Palette, Scale: float32(math.Inf(1))}, {Palette: valid.Palette, Offset: float32(math.NaN())}} {
		if p, err := NewIndexedPalette(config); err == nil {
			p.Close()
			t.Fatalf("accepted invalid config %+v", config)
		}
	}
	p, err := NewIndexedPalette(valid)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	image := ebiten.NewImage(1, 1)
	defer image.Deallocate()
	if p.Draw(image, image) == nil || p.Draw(nil, image) == nil || p.Draw(image, nil) == nil || p.SetPalette(nil) == nil {
		t.Fatal("accepted invalid draw or palette")
	}
	p.Close()
	if p.Draw(image, image) == nil || p.SetPalette(valid.Palette) == nil {
		t.Fatal("accepted use after close")
	}
}
