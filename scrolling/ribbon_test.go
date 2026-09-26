package scrolling

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestRibbonFacadeUsesAtlasAdvanceAndIndependentCullWidth(t *testing.T) {
	atlasImage := ebiten.NewImage(2, 1)
	defer atlasImage.Deallocate()
	atlasImage.Set(0, 0, color.RGBA{R: 255, A: 255})
	atlasImage.Set(1, 0, color.RGBA{B: 255, A: 255})
	metrics, err := font.NewGrid(font.Grid{Bounds: atlasImage.Bounds(), Cell: image.Pt(1, 1),
		Columns: 2, Order: "AB"})
	if err != nil {
		t.Fatal(err)
	}
	atlas, err := NewAtlas(atlasImage, metrics)
	if err != nil {
		t.Fatal(err)
	}
	config := RibbonConfig{Text: "AB", Font: atlas, CullAdvance: 40,
		Clock: motion.RibbonClockConfig{Offset: 2, Velocity: -1, Restart: 4,
			Multiplier: 1, Wrap: motion.RibbonWrapBelow}}
	scroll, err := New(Config{Ribbon: &config})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	ribbon := scroll.RibbonController()
	if ribbon == nil || ribbon.Length() != 2 {
		t.Fatal("ribbon did not retain the actual two-pixel font advance")
	}
	if err := scroll.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(4, 1)
	defer dst.Deallocate()
	scroll.Draw(dst)
	if got := color.RGBAModel.Convert(dst.At(1, 0)).(color.RGBA); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("first ribbon glyph %+v", got)
	}
	if got := color.RGBAModel.Convert(dst.At(2, 0)).(color.RGBA); got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("second ribbon glyph %+v", got)
	}
}
