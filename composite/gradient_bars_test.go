package composite

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

func TestGradientBarsValidationAndBorrowedLifetime(t *testing.T) {
	valid := GradientBarsConfig{Columns: 2, Baseline: 20, Step: 8, Width: 7, Level: func(int) float64 { return 10 }}
	for _, mutate := range []func(*GradientBarsConfig){
		func(c *GradientBarsConfig) { c.Columns = 0 },
		func(c *GradientBarsConfig) { c.Columns = 65537 },
		func(c *GradientBarsConfig) { c.Level = nil },
		func(c *GradientBarsConfig) { c.Step = 0 },
		func(c *GradientBarsConfig) { c.Width = -1 },
		func(c *GradientBarsConfig) { c.X = math.NaN() },
		func(c *GradientBarsConfig) { c.Baseline = math.Inf(1) },
		func(c *GradientBarsConfig) { c.Step = math.MaxFloat64 },
		func(c *GradientBarsConfig) { c.OutlineWidth = -1 },
		func(c *GradientBarsConfig) { c.Source = image.Rect(0, 0, 2, 1) },
	} {
		config := valid
		mutate(&config)
		if bars, err := NewGradientBars(config); err == nil {
			bars.Close()
			t.Fatalf("accepted invalid config: %+v", config)
		}
	}
	white := ebiten.NewImage(1, 1)
	defer white.Deallocate()
	outline := color.NRGBA{R: 128, A: 128}
	valid.Texture, valid.OutlineColor = white, &outline
	valid.Top, valid.Bottom = color.NRGBA{G: 255, A: 128}, color.NRGBA{B: 255, A: 255}
	bars, err := NewGradientBars(valid)
	if err != nil {
		t.Fatal(err)
	}
	outline = color.NRGBA{}
	if bars.stroke[0].ColorA != float32(128)/255 || bars.fill[0].ColorA != float32(128)/255 {
		t.Fatal("palette was not premultiplied and copied")
	}
	if bars.fill[1].SrcX != 1 || bars.fill[2].SrcY != 1 || bars.config.OutlineWidth != 1 {
		t.Fatal("default source coordinates or outline width are incorrect")
	}
	bars.Close()
	bars.Close()
	if white.Bounds().Dx() != 1 || bars.texture != nil || bars.batch != nil || bars.config.Level != nil {
		t.Fatal("closed a borrowed material or retained renderer resources")
	}
	var absent *GradientBars
	absent.Draw(nil)
	absent.DrawOutline(nil)
	absent.Close()
}

func TestGradientBarQuadPreservesVertexColorAndUV(t *testing.T) {
	top, bottom := color.NRGBA{R: 255, G: 255, A: 255}, color.NRGBA{R: 255, A: 255}
	template := [4]ebiten.Vertex{
		render.Vertex(0, 0, 2, 3, top), render.Vertex(0, 0, 7, 3, top),
		render.Vertex(0, 0, 7, 9, bottom), render.Vertex(0, 0, 2, 9, bottom),
	}
	quad := barQuad(template, 8, 14.5, 7, 8.5)
	want := [4][2]float32{{8, 14.5}, {15, 14.5}, {15, 23}, {8, 23}}
	for i, vertex := range quad {
		if vertex.DstX != want[i][0] || vertex.DstY != want[i][1] || vertex.SrcX != template[i].SrcX ||
			vertex.SrcY != template[i].SrcY || vertex.ColorG != template[i].ColorG || vertex.ColorA != template[i].ColorA {
			t.Fatalf("vertex %d: %+v", i, vertex)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { barQuad(template, 8, 14.5, 7, 8.5) }); allocations != 0 {
		t.Fatalf("quad allocated %v times", allocations)
	}
}
