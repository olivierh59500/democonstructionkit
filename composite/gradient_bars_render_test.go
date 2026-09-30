//go:build dck_gpu_rendercheck

package composite

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/render"
)

func TestGradientBarsNativeFilledAndOutlineGPU(t *testing.T) {
	testutil.RequireGPU(t)
	white, reference, actual := ebiten.NewImage(1, 1), ebiten.NewImage(640, 480), ebiten.NewImage(640, 480)
	defer white.Deallocate()
	defer reference.Deallocate()
	defer actual.Deallocate()
	white.Fill(color.White)
	levels := [80]float64{.5, 1, 2, 2.001, 3, 4.5, 21.2, 95.301, 98.301, 70, 60.66}
	bars, err := NewGradientBars(GradientBarsConfig{
		Columns: 80, Baseline: 479, Step: 8, Width: 7, Texture: white,
		Level: func(column int) float64 { return levels[column] },
		Top:   color.NRGBA{R: 255, G: 255, A: 255}, Bottom: color.NRGBA{R: 255, A: 255},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer bars.Close()
	batch := render.NewBatch(20000)
	for _, outlined := range []bool{false, true} {
		reference.Clear()
		actual.Clear()
		batch.Begin(reference, white)
		for column, height := range levels {
			if height <= 0 {
				continue
			}
			x, top := float64(column*8), 479-height
			if outlined {
				if height <= 2 {
					continue
				}
				source := image.Rect(0, 0, 1, 1)
				batch.Rect(x, top, 7, 1, source, color.White)
				batch.Rect(x, top+height-1, 7, 1, source, color.White)
				batch.Rect(x, top+1, 1, height-2, source, color.White)
				batch.Rect(x+6, top+1, 1, height-2, source, color.White)
			} else {
				yellow, red := color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 0, 255}
				batch.Quad([4]ebiten.Vertex{
					render.Vertex(x, top, 0, 0, yellow), render.Vertex(x+7, top, 1, 0, yellow),
					render.Vertex(x+7, 479, 1, 1, red), render.Vertex(x, 479, 0, 1, red),
				})
			}
		}
		batch.Flush()
		if outlined {
			bars.DrawOutline(actual)
		} else {
			bars.Draw(actual)
		}
		first, second := make([]byte, 640*480*4), make([]byte, 640*480*4)
		reference.ReadPixels(first)
		actual.ReadPixels(second)
		if !bytes.Equal(first, second) {
			t.Fatalf("native meter geometry changed (outlined=%v)", outlined)
		}
	}
}

func TestGradientBarsOutlineCornersAndInvalidHeightGPU(t *testing.T) {
	testutil.RequireGPU(t)
	dst := ebiten.NewImage(32, 20)
	defer dst.Deallocate()
	outline := color.NRGBA{R: 255, A: 128}
	heights := []float64{8, 2, math.NaN(), math.Inf(1)}
	bars, err := NewGradientBars(GradientBarsConfig{
		Columns: 4, X: 1, Baseline: 12, Step: 8, Width: 6,
		Level: func(column int) float64 { return heights[column] }, OutlineColor: &outline,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer bars.Close()
	bars.DrawOutline(dst)
	pixels := make([]byte, 32*20*4)
	dst.ReadPixels(pixels)
	for _, point := range [][2]int{{1, 4}, {6, 4}, {1, 11}, {6, 11}, {1, 7}} {
		if alpha := pixels[(point[1]*32+point[0])*4+3]; alpha != 128 {
			t.Fatalf("corner/edge %v alpha %d; overlapping strips would brighten it", point, alpha)
		}
	}
	if pixels[(7*32+3)*4+3] != 0 {
		t.Fatal("outline filled the bar interior")
	}
	for y := 0; y < 20; y++ {
		for x := 8; x < 32; x++ {
			if pixels[(y*32+x)*4+3] != 0 {
				t.Fatal("tiny or invalid callback height produced geometry")
			}
		}
	}
}
