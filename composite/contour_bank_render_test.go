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

func contourPointInside(points []image.Point, x, y float64) bool {
	inside := false
	for i, a := range points {
		d := points[(i+1)%len(points)]
		if (float64(a.Y) > y) != (float64(d.Y) > y) {
			cross := float64(a.X) + (y-float64(a.Y))*float64(d.X-a.X)/float64(d.Y-a.Y)
			if x < cross {
				inside = !inside
			}
		}
	}
	return inside
}

func TestContourBankGpuParityAndRetainedLayers(t *testing.T) {
	testutil.RequireGPU(t)
	b, err := NewContourBank(ContourBankConfig{Width: 32, Height: 32, Slots: 3, Layers: 2, FillRule: ebiten.FillRuleEvenOdd, Blend: ebiten.BlendXor})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	outer := []image.Point{{2, 2}, {26, 2}, {26, 10}, {14, 10}, {14, 26}, {2, 26}}
	hole := []image.Point{{4, 4}, {10, 4}, {10, 8}, {4, 8}}
	draw := func(batch *render.Batch) {
		for _, points := range [][]image.Point{outer, hole} {
			batch.Fan(len(points), func(i int) ebiten.Vertex {
				p := points[i]
				return render.Vertex(float64(p.X), float64(p.Y), 0, 0, color.White)
			})
		}
	}
	if err := b.Paint(0, 1, true, draw); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, 32*32*4)
	b.Image(0, 1).ReadPixels(got)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			want := contourPointInside(outer, float64(x)+.5, float64(y)+.5) != contourPointInside(hole, float64(x)+.5, float64(y)+.5)
			if (got[(y*32+x)*4+3] != 0) != want {
				t.Fatalf("fan parity differs from ray test at %d,%d", x, y)
			}
		}
	}
	b.Paint(0, 1, false, draw)
	b.Image(0, 1).ReadPixels(got)
	for _, value := range got {
		if value != 0 {
			t.Fatal("XOR write did not remove retained contours")
		}
	}
	b.Paint(2, 0, true, draw)
	b.ClearSlot(0)
	b.Image(2, 0).ReadPixels(got)
	if got[(3*32+3)*4+3] != 255 {
		t.Fatal("clearing one slot erased another retained slot")
	}
	b.Clear()
	b.Image(2, 0).ReadPixels(got)
	for _, value := range got {
		if value != 0 {
			t.Fatal("bank reset left old geometry")
		}
	}
}

func TestContourBankGpuModifiedEdgesAndStroke(t *testing.T) {
	testutil.RequireGPU(t)
	b, err := NewContourBank(ContourBankConfig{Width: 32, Height: 32, Slots: 1, Layers: 2, FillRule: ebiten.FillRuleEvenOdd})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	points := []image.Point{{8, 4}, {25, 8}, {18, 24}, {4, 20}}
	at := func(i int) ebiten.Vertex {
		p := points[i]
		return render.Vertex(float64(p.X), float64(p.Y), 0, 0, color.White)
	}
	for _, swap := range []bool{false, true} {
		b.Paint(0, 0, true, func(batch *render.Batch) {
			batch.ParityContour(len(points), at, render.ParityContour{RayX: -1, SwapY: swap})
		})
		got := make([]byte, 32*32*4)
		b.Image(0, 0).ReadPixels(got)
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				parity := false
				py, px := float64(y)+.5, float64(x)+.5
				for i, a := range points {
					d := points[(i+1)%len(points)]
					if swap {
						a.Y, d.Y = d.Y, a.Y
					}
					if (float64(a.Y) > py) != (float64(d.Y) > py) {
						cross := float64(a.X) + (py-float64(a.Y))*float64(d.X-a.X)/float64(d.Y-a.Y)
						if px < cross {
							parity = !parity
						}
					}
				}
				if (got[(y*32+x)*4+3] != 0) != parity {
					t.Fatalf("edge parity swap=%t differs at %d,%d", swap, x, y)
				}
			}
		}
	}
	box := []image.Point{{5, 5}, {25, 5}, {25, 25}, {5, 25}}
	b.Paint(0, 1, true, func(batch *render.Batch) {
		batch.StrokeContour(len(box), func(i int) ebiten.Vertex {
			p := box[i]
			return render.Vertex(float64(p.X), float64(p.Y), 0, 0, color.White)
		}, render.ContourStroke{Width: 2, SkipHorizontal: true})
	})
	got := make([]byte, 32*32*4)
	b.Image(0, 1).ReadPixels(got)
	if got[(15*32+5)*4+3] == 0 || got[(5*32+15)*4+3] != 0 || got[(15*32+15)*4+3] != 0 {
		t.Fatal("selective outline drew horizontal edges or filled its interior")
	}
}
