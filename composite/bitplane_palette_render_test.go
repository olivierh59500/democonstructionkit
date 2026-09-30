//go:build dck_gpu_rendercheck

package composite

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/render"
)

func TestBitplanePaletteGPU(t *testing.T) {
	testutil.RequireGPU(t)
	for scenario := 0; scenario < 12; scenario++ {
		count := scenario%6 + 1
		mixed := scenario >= 6
		colors := make([]color.NRGBA, 1<<count)
		for i := range colors {
			colors[i] = color.NRGBA{R: uint8(i*3 + 7), G: uint8(255 - i*2), B: uint8(i*4 + 1), A: 255}
		}
		config := BitplanePaletteConfig{Width: 8, Height: 8, Planes: count, Palette: colors}
		if mixed {
			config.Channels = make([]BitplaneChannel, count)
			for i := range config.Channels {
				config.Channels[i] = BitplaneChannel(i % 4)
			}
		}
		effect, err := NewBitplanePalette(config)
		if err != nil {
			t.Fatal(err)
		}
		planes := make([]*ebiten.Image, count)
		for bit := range planes {
			pixels := make([]byte, 8*8*4)
			channel := BitplaneAlpha
			if mixed {
				channel = config.Channels[bit]
			}
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					at := (y*8 + x) * 4
					if channel != BitplaneAlpha {
						pixels[at+3] = 255 // Opaque art carries bits in RGB, not alpha.
					}
					if ((x + y*8) & (1 << bit)) != 0 {
						if channel == BitplaneAlpha {
							pixels[at], pixels[at+1], pixels[at+2], pixels[at+3] = 255, 255, 255, 255
						} else {
							pixels[at+int(channel)-1] = 255
						}
					}
				}
			}
			if mixed && bit == count-1 {
				planes[bit] = ebiten.NewImageFromImage(&image.NRGBA{Pix: pixels, Stride: 8 * 4, Rect: image.Rect(0, 0, 8, 8)})
			} else {
				planes[bit] = render.NewSurface(8, 8)
				planes[bit].WritePixels(pixels)
			}
		}
		dst := ebiten.NewImage(8, 8)
		if err := effect.Draw(dst, planes); err != nil {
			t.Fatal(err)
		}
		pixels := make([]byte, 8*8*4)
		dst.ReadPixels(pixels)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				index := (x + y*8) & ((1 << count) - 1)
				want := colors[index]
				at := (y*8 + x) * 4
				if pixels[at] != want.R || pixels[at+1] != want.G || pixels[at+2] != want.B || pixels[at+3] != want.A {
					t.Fatalf("%d planes at (%d,%d): got %v, want %v", count, x, y, pixels[at:at+4], want)
				}
			}
		}
		colors[0] = color.NRGBA{R: 200, A: 255}
		if err := effect.SetPalette(colors); err != nil {
			t.Fatal(err)
		}
		if err := effect.Draw(dst, planes); err != nil {
			t.Fatal(err)
		}
		dst.ReadPixels(pixels)
		if pixels[0] != 200 || pixels[1] != 0 || pixels[2] != 0 || pixels[3] != 255 {
			t.Fatalf("%d planes did not apply updated palette: %v", count, pixels[:4])
		}
		for _, plane := range planes {
			plane.Deallocate()
		}
		dst.Deallocate()
		effect.Close()
	}
}

func TestBitplanePaletteThresholdAndTransparentOverlay(t *testing.T) {
	testutil.RequireGPU(t)
	effect, err := NewBitplanePalette(BitplanePaletteConfig{Width: 2, Height: 1, Planes: 1,
		Palette: []color.NRGBA{{}, {R: 200, G: 100, A: 128}}})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	mask := ebiten.NewImage(2, 1)
	defer mask.Deallocate()
	mask.WritePixels([]byte{100, 100, 100, 100, 200, 200, 200, 200})
	dst := ebiten.NewImage(2, 1)
	defer dst.Deallocate()
	dst.Fill(color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	if err := effect.Draw(dst, []*ebiten.Image{mask}); err != nil {
		t.Fatal(err)
	}
	pixels := make([]byte, 8)
	dst.ReadPixels(pixels)
	if pixels[0] != 10 || pixels[1] != 20 || pixels[2] != 30 || pixels[3] != 255 {
		t.Fatalf("transparent zero index obscured the background: %v", pixels[:4])
	}
	for i, want := range []int{105, 60, 15, 255} {
		if got := int(pixels[4+i]); got < want-1 || got > want+1 {
			t.Fatalf("translucent palette entry at channel %d = %d, want about %d", i, got, want)
		}
	}
}
