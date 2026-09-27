package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestPaletteAtlasPacksIrregularFramesAndPreservesSourceColors(t *testing.T) {
	source := image.NewRGBA(image.Rect(10, 20, 13, 23))
	source.Set(10, 20, color.RGBA{96, 0, 0, 255})
	source.Set(10, 21, color.RGBA{160, 0, 0, 255})
	source.Set(11, 21, color.RGBA{240, 0, 0, 255})
	source.Set(11, 22, color.RGBA{0, 80, 0, 255})
	palettes := [][]color.RGBA{
		{{11, 12, 13, 255}, {21, 22, 23, 255}, {31, 32, 33, 255}, {41, 42, 43, 255}, {51, 52, 53, 255}},
		{{61, 62, 63, 255}, {71, 72, 73, 255}, {81, 82, 83, 255}, {91, 92, 93, 255}, {101, 102, 103, 255}},
	}
	config := PaletteAtlasConfig{
		Source: source, Frames: []image.Rectangle{
			image.Rect(10, 20, 11, 21), image.Rect(10, 21, 12, 23),
		},
		Palettes: palettes, RowHeight: 2, RepeatLast: true,
		OriginalFrames: []image.Rectangle{image.Rect(10, 20, 12, 22)},
		Index: func(pixel color.Color) int {
			red, _, _, _ := pixel.RGBA()
			return int(uint8(red>>8)>>5) - 3
		},
	}
	atlas, err := BuildPaletteAtlas(config)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := atlas.Pixels.Bounds(), image.Rect(0, 0, 3, 6); got != want {
		t.Fatalf("atlas bounds = %v, want %v", got, want)
	}
	wantRects := []image.Rectangle{
		image.Rect(0, 0, 1, 1), image.Rect(1, 0, 3, 2), image.Rect(1, 0, 3, 2),
		image.Rect(0, 2, 1, 3), image.Rect(1, 2, 3, 4), image.Rect(1, 2, 3, 4),
		image.Rect(0, 4, 2, 6),
	}
	if len(atlas.Rects) != len(wantRects) {
		t.Fatalf("atlas has %d crops, want %d", len(atlas.Rects), len(wantRects))
	}
	for index, want := range wantRects {
		if got := atlas.Rects[index]; got != want {
			t.Fatalf("crop %d = %v, want %v", index, got, want)
		}
	}
	for _, sample := range []struct {
		at   image.Point
		want color.RGBA
	}{
		{image.Pt(0, 0), palettes[0][0]},
		{image.Pt(1, 0), palettes[0][2]},
		{image.Pt(2, 0), palettes[0][4]},
		{image.Pt(0, 2), palettes[1][0]},
		{image.Pt(2, 2), palettes[1][4]},
		{image.Pt(2, 1), color.RGBA{0, 80, 0, 255}},
		{image.Pt(2, 3), color.RGBA{0, 80, 0, 255}},
		{image.Pt(0, 4), color.RGBA{96, 0, 0, 255}},
		{image.Pt(1, 1), color.RGBA{}},
	} {
		if got := atlas.Pixels.RGBAAt(sample.at.X, sample.at.Y); got != sample.want {
			t.Fatalf("pixel %v = %v, want %v", sample.at, got, sample.want)
		}
	}
}

func TestPaletteAtlasRejectsInvalidGeometry(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2, 2))
	config := PaletteAtlasConfig{
		Source: source, Frames: []image.Rectangle{source.Bounds()},
		Palettes: [][]color.RGBA{{{255, 0, 0, 255}}},
		Index:    func(color.Color) int { return 0 },
	}
	if _, err := BuildPaletteAtlas(config); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*PaletteAtlasConfig){
		func(c *PaletteAtlasConfig) { c.Frames[0] = image.Rect(0, 0, 3, 2) },
		func(c *PaletteAtlasConfig) { c.RowHeight = 1 },
		func(c *PaletteAtlasConfig) { c.Index = nil },
		func(c *PaletteAtlasConfig) { c.OriginalFrames = []image.Rectangle{image.Rect(-1, 0, 1, 1)} },
	} {
		invalid := config
		invalid.Frames = append([]image.Rectangle(nil), config.Frames...)
		change(&invalid)
		if _, err := BuildPaletteAtlas(invalid); err == nil {
			t.Fatalf("accepted invalid palette atlas: %+v", invalid)
		}
	}
	oversized := config
	oversized.Source = image.NewUniform(color.White)
	oversized.Frames = []image.Rectangle{image.Rect(0, 0, 8192, 8192)}
	if _, err := BuildPaletteAtlas(oversized); err == nil {
		t.Fatal("accepted an atlas above the mobile pixel budget")
	}
}
