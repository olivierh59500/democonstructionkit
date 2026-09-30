//go:build dck_gpu_rendercheck

package effects

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/palette"
)

func indexedImageSource(width, height int, values []byte) *ebiten.Image {
	source := ebiten.NewImage(width, height)
	pixels := make([]byte, width*height*4)
	for i, value := range values {
		pixels[i*4], pixels[i*4+3] = value, 255
	}
	source.WritePixels(pixels)
	return source
}

func indexedImagePixel(t *testing.T, dst *ebiten.Image, want [4]byte) {
	t.Helper()
	pixels := make([]byte, 4)
	dst.ReadPixels(pixels)
	for c, value := range want {
		if pixels[c] != value {
			t.Fatalf("component %d = %d, want %d (whole pixel %v)", c, pixels[c], value, pixels)
		}
	}
}

func TestIndexedImageAbsoluteFadeAndBackwardSeekGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := indexedImageSource(1, 1, []byte{0})
	defer source.Deallocate()
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0xf00}, FPS: 50, Channel: composite.BitplaneRed})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	from, to := []uint32{0xf00}, []uint32{0}
	if err := e.Fade(from, to, 4, 5); err != nil {
		t.Fatal(err)
	}
	from[0], to[0] = 0, 0xfff
	dst := ebiten.NewImage(1, 1)
	defer dst.Deallocate()
	// The descending midpoint is eight, not seven: signed delta division
	// truncates toward zero before adding the initial channel.
	for _, sample := range []struct {
		tick int
		red  byte
	}{{-2, 255}, {3, 255}, {4, 255}, {5, 204}, {6, 136}, {7, 68}, {8, 0}, {30, 0}, {6, 136}, {3, 255}} {
		if err := e.Update(kit.Frame{Time: float64(sample.tick) / 50}); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			dst.Clear()
			e.Draw(dst)
			indexedImagePixel(t, dst, [4]byte{sample.red, 0, 0, 255})
		}
	}
	if err := e.Update(kit.Frame{Time: 6.49 / 50}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{136, 0, 0, 255})
	if err := e.Update(kit.Frame{Time: 6.51 / 50}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{68, 0, 0, 255})
	words := []uint32{0x0f0}
	if err := e.SetPalette(words); err != nil {
		t.Fatal(err)
	}
	words[0] = 0xfff
	if err := e.Update(kit.Frame{Time: 1}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{0, 255, 0, 255})
	if err := e.Fade([]uint32{0x00f}, []uint32{0xf00}, 4, 1); err != nil {
		t.Fatal(err)
	}
	for _, time := range []float64{0, .08, 10} {
		if err := e.Update(kit.Frame{Time: time}); err != nil {
			t.Fatal(err)
		}
		dst.Clear()
		e.Draw(dst)
		indexedImagePixel(t, dst, [4]byte{0, 0, 255, 255})
	}
}

func TestIndexedImageRGB565FadeGPU(t *testing.T) {
	testutil.RequireGPU(t)
	format, err := palette.NewPackedRGB(palette.PackedRGBConfig{Bits: [3]uint8{5, 6, 5}, Shift: [3]uint8{11, 5, 0}})
	if err != nil {
		t.Fatal(err)
	}
	source := indexedImageSource(1, 1, []byte{0})
	defer source.Deallocate()
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0xf800}, Format: format, Channel: composite.BitplaneRed})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if err := e.Fade([]uint32{0xf800}, []uint32{0x07ff}, 0, 3); err != nil {
		t.Fatal(err)
	}
	if err := e.Update(kit.Frame{Time: 1.0 / 60}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(1, 1)
	defer dst.Deallocate()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{131, 125, 123, 255})
}

func TestIndexedImageSubImageCropAndOutputPoseGPU(t *testing.T) {
	testutil.RequireGPU(t)
	parent := indexedImageSource(5, 4, []byte{0, 0, 0, 0, 0, 0, 0, 17, 34, 0, 0, 51, 68, 85, 0, 0, 0, 0, 0, 0})
	defer parent.Deallocate()
	source := parent.SubImage(image.Rect(1, 1, 4, 3)).(*ebiten.Image)
	colors := []uint32{0xf00, 0x0f0, 0x00f, 0xff0, 0x0ff, 0xf0f}
	options := ebiten.DrawImageOptions{}
	options.GeoM.Scale(2, 2)
	options.GeoM.Translate(1, 2)
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: colors, Channel: composite.BitplaneRed, Scale: 15,
		Crop: image.Rect(1, 0, 3, 2), Options: options})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	colors[1] = 0xfff
	dst := ebiten.NewImage(7, 8)
	defer dst.Deallocate()
	e.Draw(dst)
	pixels := make([]byte, 7*8*4)
	dst.ReadPixels(pixels)
	crop := [2][2][4]byte{{{0, 255, 0, 255}, {0, 0, 255, 255}}, {{0, 255, 255, 255}, {255, 0, 255, 255}}}
	for y := 0; y < 8; y++ {
		for x := 0; x < 7; x++ {
			want := [4]byte{}
			if x >= 1 && x < 5 && y >= 2 && y < 6 {
				want = crop[(y-2)/2][(x-1)/2]
			}
			for c, value := range want {
				if got := pixels[(y*7+x)*4+c]; got != value {
					t.Fatalf("crop/output pixel (%d,%d) component %d = %d, want %d", x, y, c, got, value)
				}
			}
		}
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	e.Draw(dst)
	dst.ReadPixels(pixels)
	for _, value := range pixels {
		if value != 0 {
			t.Fatal("closed indexed image still drew")
		}
	}
	parent.Fill(color.NRGBA{R: 80, A: 255})
	dst.DrawImage(source, nil)
	dst.ReadPixels(pixels)
	if pixels[0] != 80 || pixels[3] != 255 {
		t.Fatal("closing indexed image invalidated borrowed atlas or subimage")
	}
}

func TestIndexedImageSourceAlphaGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := ebiten.NewImage(1, 1)
	defer source.Deallocate()
	source.WritePixels([]byte{1, 0, 0, 128})
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0, 0xfff}, Channel: composite.BitplaneRed,
		Scale: 255, SourceAlpha: true})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	dst := ebiten.NewImage(1, 1)
	defer dst.Deallocate()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{128, 128, 128, 128})
}

func TestIndexedImageRejectsInvalidStateWithoutChangingFadeGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := indexedImageSource(1, 1, []byte{0})
	defer source.Deallocate()
	for _, config := range []IndexedImageConfig{{}, {Image: source}, {Image: source, Palette: make([]uint32, 257)},
		{Image: source, Palette: []uint32{0}, FPS: -1}, {Image: source, Palette: []uint32{0}, FPS: math.NaN()},
		{Image: source, Palette: []uint32{0}, FPS: math.Inf(1)},
		{Image: source, Palette: []uint32{0}, Crop: image.Rect(-1, 0, 1, 1)},
		{Image: source, Palette: []uint32{0}, Crop: image.Rect(0, 0, 2, 1)},
		{Image: source, Palette: []uint32{0}, Channel: composite.BitplaneBlue + 1},
		{Image: source, Palette: []uint32{0}, Scale: -1}} {
		if e, err := NewIndexedImage(config); err == nil {
			e.Close()
			t.Fatalf("accepted invalid config %+v", config)
		}
	}
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0xf00}, FPS: 1, Channel: composite.BitplaneRed})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if err := e.Fade([]uint32{0xf00}, []uint32{0}, 0, 5); err != nil {
		t.Fatal(err)
	}
	if err := e.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	for _, time := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64, float64(1 << 31)} {
		if err := e.Update(kit.Frame{Time: time}); err == nil {
			t.Fatalf("accepted invalid/overflowing clock %v", time)
		}
	}
	for _, fade := range []struct {
		start, ticks int
	}{{-1, 5}, {0, 0}} {
		if err := e.Fade([]uint32{0}, []uint32{0xfff}, fade.start, fade.ticks); err == nil {
			t.Fatalf("accepted invalid fade %+v", fade)
		}
	}
	// Keep overflow checks portable: these cases exist on 64-bit hosts.
	if ^uint(0)>>32 != 0 {
		limit := int64(1)<<31 - 1
		overflow := int(limit + 1)
		if e.Fade([]uint32{0}, []uint32{0xfff}, overflow, 5) == nil || e.Fade([]uint32{0}, []uint32{0xfff}, 0, overflow) == nil {
			t.Fatal("accepted overflowing fade bounds")
		}
	}
	if e.Fade(nil, []uint32{0}, 0, 5) == nil || e.SetPalette(nil) == nil {
		t.Fatal("accepted a mismatched palette bank")
	}
	dst := ebiten.NewImage(1, 1)
	defer dst.Deallocate()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{136, 0, 0, 255})
	if err := e.Update(kit.Frame{Time: 3}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{68, 0, 0, 255})
	const lastTick = 1<<31 - 1
	if err := e.Fade([]uint32{0x00f}, []uint32{0}, lastTick, lastTick); err != nil {
		t.Fatalf("rejected bounded fade: %v", err)
	}
	if err := e.Update(kit.Frame{Time: lastTick}); err != nil {
		t.Fatalf("rejected bounded clock: %v", err)
	}
	dst.Clear()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{0, 0, 255, 255})
	e.Close()
	if e.Update(kit.Frame{}) == nil || e.SetPalette([]uint32{0}) == nil || e.Fade([]uint32{0}, []uint32{0}, 0, 1) == nil {
		t.Fatal("accepted state changes after close")
	}
}

func TestIndexedImageSurfaceBudgetAndOverflowingClockGPU(t *testing.T) {
	testutil.RequireGPU(t)
	for _, dimensions := range [][2]int{{8193, 1}, {1, 8193}} {
		source := ebiten.NewImage(dimensions[0], dimensions[1])
		e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0}})
		if err == nil {
			e.Close()
			t.Fatalf("accepted oversized source %v", dimensions)
		}
		source.Deallocate()
	}
	source := indexedImageSource(1, 1, []byte{0})
	defer source.Deallocate()
	e, err := NewIndexedImage(IndexedImageConfig{Image: source, Palette: []uint32{0x0f0}, FPS: math.MaxFloat64, Channel: composite.BitplaneRed})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if err := e.Update(kit.Frame{Time: 2}); err == nil {
		t.Fatal("accepted an infinite product of finite time and FPS")
	}
	if err := e.Update(kit.Frame{Time: -math.MaxFloat64}); err != nil {
		t.Fatalf("finite negative time should clamp to zero: %v", err)
	}
	dst := ebiten.NewImage(1, 1)
	defer dst.Deallocate()
	e.Draw(dst)
	indexedImagePixel(t, dst, [4]byte{0, 255, 0, 255})
}
