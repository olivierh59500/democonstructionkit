//go:build dck_gpu_rendercheck

package composite

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/render"
)

// The oracle samples CPU source pixels independently of shader coordinates.
// Several planes borrow the same atlased material with different offsets.
func TestBitplanePaletteIndependentOffsetsGPU(t *testing.T) {
	testutil.RequireGPU(t)
	const size = 8
	mask := render.NewSurface(size, size)
	defer mask.Deallocate()
	maskPixels := make([]byte, size*size*4)
	materialPixels := make([]byte, size*size*4)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			at := (y*size + x) * 4
			if (x+y)%3 == 0 {
				maskPixels[at+3] = 255
			}
			materialPixels[at] = byte(((x + 2*y) / 2 % 2) * 255)
			materialPixels[at+1] = byte(((3*x + y) / 3 % 2) * 255)
			materialPixels[at+2] = byte(((x ^ y) % 2) * 255)
			materialPixels[at+3] = 255
		}
	}
	mask.WritePixels(maskPixels)
	material := ebiten.NewImageFromImage(&image.NRGBA{Pix: materialPixels, Stride: size * 4, Rect: image.Rect(0, 0, size, size)})
	defer material.Deallocate()
	for count := 1; count <= 6; count++ {
		colors := make([]color.NRGBA, 1<<count)
		for i := range colors {
			colors[i] = color.NRGBA{R: uint8(i*3 + 7), G: uint8(250 - i*2), B: uint8(i*4 + 1), A: 255}
		}
		planes := make([]*ebiten.Image, count)
		channels := make([]BitplaneChannel, count)
		planes[0] = mask
		for i := 1; i < count; i++ {
			planes[i], channels[i] = material, BitplaneChannel((i-1)%3+1)
		}
		lookup, err := NewBitplanePalette(BitplanePaletteConfig{
			Width: size, Height: size, Planes: count, Palette: colors, Channels: channels,
		})
		if err != nil {
			t.Fatal(err)
		}
		dst := ebiten.NewImage(size, size)
		pixels := make([]byte, size*size*4)
		for _, offsets := range [][][2]float32{
			{{1, -1}, {-2, 1}, {3, -2}, {-1, 2}, {2, 3}, {-3, -1}},
			nil, // Draw must reset the offsets left by the preceding draw.
			{{0.25, -0.75}, {-1.75, 1.25}, {2.25, -1.75}, {-0.75, 2.25}, {1.25, 2.25}, {-2.75, -0.75}},
		} {
			if offsets == nil {
				err = lookup.Draw(dst, planes)
			} else {
				err = lookup.DrawOffsets(dst, planes, offsets[:count])
			}
			if err != nil {
				t.Fatal(err)
			}
			dst.ReadPixels(pixels)
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					index := 0
					for bit := 0; bit < count; bit++ {
						sx, sy := x, y
						if offsets != nil {
							sx = int(math.Floor(float64(x) + 0.5 + float64(offsets[bit][0])))
							sy = int(math.Floor(float64(y) + 0.5 + float64(offsets[bit][1])))
						}
						if sx < 0 || sy < 0 || sx >= size || sy >= size {
							continue
						}
						component, source := 3, maskPixels
						if bit > 0 {
							component, source = int(channels[bit])-1, materialPixels
						}
						if source[(sy*size+sx)*4+component] >= 128 {
							index |= 1 << bit
						}
					}
					want, at := colors[index], (y*size+x)*4
					if pixels[at] != want.R || pixels[at+1] != want.G || pixels[at+2] != want.B || pixels[at+3] != want.A {
						t.Fatalf("%d planes with offsets %v at (%d,%d): got %v, want %v", count, offsets, x, y, pixels[at:at+4], want)
					}
				}
			}
		}
		dst.Deallocate()
		lookup.Close()
	}
}

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
