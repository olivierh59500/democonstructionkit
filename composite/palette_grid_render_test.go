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

func TestPaletteGridMappedCellsAndMaskOffsetsGPU(t *testing.T) {
	testutil.RequireGPU(t)
	const width, height = 8, 8
	controlCPU := image.NewNRGBA(image.Rect(0, 0, width, height))
	bodyCPU := image.NewNRGBA(controlCPU.Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			a := uint8(127)
			if (x+y)%3 == 0 {
				a = 128
			}
			controlCPU.SetNRGBA(x, y, color.NRGBA{A: a})
			blue := uint8(0)
			if x == 3 && y >= 2 && y <= 4 {
				blue = 255
			}
			bodyCPU.SetNRGBA(x, y, color.NRGBA{B: blue, A: 255})
		}
	}
	control := ebiten.NewImageFromImage(controlCPU) // Exercise an atlased source.
	body := ebiten.NewImageFromImage(bodyCPU)
	defer control.Deallocate()
	defer body.Deallocate()
	replacement := color.NRGBA{R: 220, G: 40, B: 180, A: 255}
	p, err := NewPaletteGrid(PaletteGridConfig{Width: width, Height: height, Columns: 2, Rows: 4,
		X:                PaletteGridAxis{CellSize: 4, Offset: 1},
		Y:                PaletteGridAxis{CellSize: 2, FirstSpan: 3, Indices: []int{0, 2, 3}},
		ControlThreshold: .5, BodyColor: &replacement, BodyChannel: BitplaneBlue})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	first, second := make([]color.NRGBA, 8), make([]color.NRGBA, 8)
	for i := range first {
		first[i] = color.NRGBA{R: uint8(10 + i*20), G: 30, B: 80, A: 255}
		second[i] = color.NRGBA{R: 40, G: uint8(20 + i*20), B: 150, A: 255}
	}
	if err := p.SetColors(first, second); err != nil {
		t.Fatal(err)
	}
	dst := render.NewSurface(width, height)
	defer dst.Deallocate()
	pixels := make([]byte, width*height*4)
	rowAt := [8]int{0, 0, 0, 2, 2, 3, 3, 3}
	for _, state := range []PaletteGridState{{}, {ControlOffset: [2]float32{1, 0}, BodyOffset: [2]float32{0, 1}}} {
		if err := p.Draw(dst, control, body, state); err != nil {
			t.Fatal(err)
		}
		dst.ReadPixels(pixels)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				col, row := min(1, (x+1)/4), rowAt[y]
				want := first[row*2+col]
				cx, cy := x+int(state.ControlOffset[0]), y+int(state.ControlOffset[1])
				if cx >= 0 && cx < width && cy >= 0 && cy < height && controlCPU.NRGBAAt(cx, cy).A >= 128 {
					want = second[row*2+col]
				}
				bx, by := x+int(state.BodyOffset[0]), y+int(state.BodyOffset[1])
				if bx >= 0 && bx < width && by >= 0 && by < height && bodyCPU.NRGBAAt(bx, by).B >= 128 {
					want = replacement
				}
				at := (y*width + x) * 4
				if pixels[at] != want.R || pixels[at+1] != want.G || pixels[at+2] != want.B || pixels[at+3] != want.A {
					t.Fatalf("grid at (%d,%d) state %+v: got %v, want %v", x, y, state, pixels[at:at+4], want)
				}
			}
		}
	}
}

func TestPaletteGridContinuousRowsAndTranslucencyGPU(t *testing.T) {
	testutil.RequireGPU(t)
	const width, height = 4, 3
	control := render.NewSurface(width, height)
	defer control.Deallocate()
	values := [4]byte{0, 64, 128, 255}
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x, value := range values {
			at := (y*width + x) * 4
			pixels[at], pixels[at+3] = value, 255
		}
	}
	control.WritePixels(pixels)
	p, err := NewPaletteGrid(PaletteGridConfig{Width: width, Height: height, Columns: 1, Rows: height,
		ControlChannel: BitplaneRed, Blend: ebiten.BlendCopy})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	first := []color.NRGBA{{R: 200, G: 40, A: 128}, {G: 220, B: 60, A: 255}, {}}
	second := []color.NRGBA{{G: 180, B: 200, A: 64}, {R: 120, B: 180, A: 255}, {R: 255, A: 128}}
	if err := p.SetColors(first, second); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(width, height)
	defer dst.Deallocate()
	if err := p.Draw(dst, control, nil, PaletteGridState{}); err != nil {
		t.Fatal(err)
	}
	dst.ReadPixels(pixels)
	for y := 0; y < height; y++ {
		r1, g1, b1, a1 := first[y].RGBA()
		r2, g2, b2, a2 := second[y].RGBA()
		one, two := [4]uint32{r1 >> 8, g1 >> 8, b1 >> 8, a1 >> 8}, [4]uint32{r2 >> 8, g2 >> 8, b2 >> 8, a2 >> 8}
		for x, value := range values {
			for c := 0; c < 4; c++ {
				want := int(math.Round(float64(one[c])*(1-float64(value)/255) + float64(two[c])*float64(value)/255))
				got := int(pixels[(y*width+x)*4+c])
				if got < want-1 || got > want+1 {
					t.Fatalf("row %d mix %d component %d: got %d, want about %d", y, value, c, got, want)
				}
			}
		}
	}
	p.Close()
	dst.Clear()
	dst.DrawImage(control, nil)
	dst.ReadPixels(pixels)
	if pixels[3] != 255 || pixels[4] != 64 {
		t.Fatal("closing palette material deallocated the borrowed control image")
	}
}
