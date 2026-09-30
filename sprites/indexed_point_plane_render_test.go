//go:build dck_gpu_rendercheck

package sprites

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func pointPlaneOver(source, destination color.NRGBA) color.NRGBA {
	a := int(source.A)
	return color.NRGBA{
		R: uint8((int(source.R)*a + int(destination.R)*(255-a)) / 255),
		G: uint8((int(source.G)*a + int(destination.G)*(255-a)) / 255),
		B: uint8((int(source.B)*a + int(destination.B)*(255-a)) / 255), A: 255,
	}
}

func TestIndexedPointPlaneGPUCollisionPaletteOffsetCropAndDedup(t *testing.T) {
	testutil.RequireGPU(t)
	points := []PointPlaneSample{
		{0, 0, 1}, {0, 0, 1},
		{2, 1, 1}, {2, 1, 1}, {2, 1, 2},
		{4, 2, 1}, {4, 2, 2}, {7, 4, 2},
		{-1, 0, 1}, {8, 0, 1}, {0, 5, 1},
	}
	palette := []color.NRGBA{{R: 9, G: 11, B: 13, A: 255},
		{R: 230, G: 40, B: 20, A: 255}, {R: 40, G: 200, B: 80, A: 128}, {R: 50, G: 30, B: 220, A: 64}}
	background := color.NRGBA{R: 20, G: 30, B: 40, A: 255}
	for _, collision := range []PointPlaneCollision{PointPlaneOR, PointPlaneXOR} {
		for _, drawZero := range []bool{false, true} {
			for _, crop := range []image.Rectangle{image.Rect(0, 0, 14, 10), image.Rect(4, 2, 10, 7)} {
				calls := 0
				config := pointPlaneFixture(collision, &points)
				config.Offset, config.DrawZero, config.Palette = image.Pt(3, 2), drawZero, palette
				config.Sample = func(i int) (PointPlaneSample, bool) { calls++; return points[i], true }
				plane, err := NewIndexedPointPlane(config)
				if err != nil {
					t.Fatal(err)
				}
				plane.Sample()
				if len(plane.Touched()) != 4 {
					t.Fatal("pixel touched more than once after XOR cancellation")
				}
				canvas := ebiten.NewImage(14, 10)
				canvas.Fill(background)
				view := canvas.SubImage(crop).(*ebiten.Image)
				plane.Draw(view)
				pixels := make([]byte, 14*10*4)
				canvas.ReadPixels(pixels)
				// Independently combine the population in an ordinary CPU map.
				masks := map[image.Point]byte{}
				for _, point := range points {
					position := image.Pt(point.X, point.Y)
					if !position.In(image.Rect(0, 0, 8, 5)) {
						continue
					}
					if collision == PointPlaneOR {
						masks[position] |= point.Mask
					} else {
						masks[position] ^= point.Mask
					}
				}
				for y := 0; y < 10; y++ {
					for x := 0; x < 14; x++ {
						want := background
						position := image.Pt(x-3, y-2)
						if mask, exists := masks[position]; exists && (mask != 0 || drawZero) && image.Pt(x, y).In(crop) {
							want = pointPlaneOver(palette[mask], want)
						}
						at := (y*14 + x) * 4
						for channel, expected := range []byte{want.R, want.G, want.B, want.A} {
							delta := int(pixels[at+channel]) - int(expected)
							if delta < -1 || delta > 1 {
								t.Fatalf("collision=%d zero=%v crop=%v at(%d,%d) channel%d=%d want%d", collision, drawZero, crop, x, y, channel, pixels[at+channel], expected)
							}
						}
					}
				}
				canvas.Fill(background)
				plane.DrawAt(view, image.Pt(3, 2))
				if calls != len(points) {
					t.Fatal("drawing advanced or resampled the population")
				}
				second := make([]byte, len(pixels))
				canvas.ReadPixels(second)
				for i, value := range pixels {
					if second[i] != value {
						t.Fatal("repeated draws changed the prepared raster")
					}
				}
				canvas.Deallocate()
				plane.Close()
			}
		}
	}
}

func TestIndexedPointPlaneGPUZeroPixelsPaletteReplacementAndBorrowedMaterial(t *testing.T) {
	testutil.RequireGPU(t)
	material := ebiten.NewImage(3, 3)
	defer material.Deallocate()
	white := material.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	white.Fill(color.White)
	points := []PointPlaneSample{{1, 1, 1}, {1, 1, 1}, {2, 1, 2}}
	config := pointPlaneFixture(PointPlaneXOR, &points)
	config.White, config.DrawZero, config.Blend = white, true, ebiten.BlendCopy
	config.Palette = []color.NRGBA{{R: 200, G: 100, B: 50, A: 128}, {}, {G: 255, A: 255}}
	plane, err := NewIndexedPointPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	plane.Sample()
	dst := ebiten.NewImage(8, 5)
	defer dst.Deallocate()
	dst.Fill(color.NRGBA{R: 255, B: 255, A: 255})
	plane.Draw(dst)
	pixels := make([]byte, 8*5*4)
	dst.ReadPixels(pixels)
	at := (1*8 + 1) * 4
	for channel, want := range []int{100, 50, 25, 128} {
		value := int(pixels[at+channel])
		if value < want-1 || value > want+1 {
			t.Fatal("DrawZero, nonzero source origin or premultiplied copy blend changed", pixels[at:at+4])
		}
	}
	plane.SetPalette([]color.NRGBA{{R: 255, A: 255}})
	dst.Clear()
	plane.Draw(dst)
	dst.ReadPixels(pixels)
	if pixels[at] != 255 || pixels[at+3] != 255 || pixels[(1*8+2)*4+3] != 0 {
		t.Fatal("palette update or unmapped transparent index changed")
	}
	plane.Close()
	plane.Close()
	white.Fill(color.NRGBA{B: 255, A: 255})
	borrowed := make([]byte, 4)
	white.ReadPixels(borrowed)
	if borrowed[2] != 255 || borrowed[3] != 255 {
		t.Fatal("closing disposed a borrowed material")
	}
	plane.Draw(dst)
	plane.DrawAt(nil, image.Point{})
	if plane.SetPalette(config.Palette) == nil || plane.Sample() == nil {
		t.Fatal("closed plane accepted updates")
	}
	wrongSize := ebiten.NewImage(2, 1)
	defer wrongSize.Deallocate()
	config.White = wrongSize
	if _, err := NewIndexedPointPlane(config); err == nil {
		t.Fatal("a multipixel white material was accepted")
	}
}
