package scrolling

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestSizeBankFacadeDrawsActiveStripAtAuthoredRows(t *testing.T) {
	font := ebiten.NewImage(1, 1)
	defer font.Deallocate()
	font.Fill(color.RGBA{R: 255, A: 255})
	config := SizeBankConfig{
		Text: "A", InitialFont: "0", ViewportWidth: 4,
		Layers: []SizeBankLayer{{Name: "0", Image: font, CellWidth: 1, CellHeight: 1,
			Columns: 1, Count: 1, ScaleX: 1, ScaleY: 1,
			RepeatY: 2, RepeatStep: 2, Repeats: 2}},
		Lookup:     func(r rune) int { return int(r - 'A') },
		BaseSpeeds: []float64{1}, StartOffset: 1, SpeedMultiplier: 1,
	}
	scroll, err := New(Config{SizeBank: &config})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	if scroll.SizeBankController() == nil {
		t.Fatal("size bank controller is not exposed by scrolling.New")
	}
	if err := scroll.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(4, 6)
	defer dst.Deallocate()
	scroll.Draw(dst)
	for _, y := range []int{2, 4} {
		got := color.RGBAModel.Convert(dst.At(0, y)).(color.RGBA)
		if got != (color.RGBA{R: 255, A: 255}) {
			t.Fatalf("row %d pixel %+v, want red", y, got)
		}
	}
	if got := color.RGBAModel.Convert(dst.At(0, 3)).(color.RGBA); got.A != 0 {
		t.Fatalf("gap between repeated rows is not transparent: %+v", got)
	}
}
