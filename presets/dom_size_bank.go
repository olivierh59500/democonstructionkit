package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

// DOMSizeBank returns the four editable font layers and their synchronized
// reference speed. The same atlas is borrowed by every layer; the fourth is
// deliberately scaled 8x horizontally and 12x vertically.
func DOMSizeBank(text string, font *ebiten.Image) (scrolling.SizeBankConfig, error) {
	lookup, err := TileLookup("go-dom-intro", true)
	if err != nil {
		return scrolling.SizeBankConfig{}, err
	}
	tile := func(r rune) int { index, _ := lookup(r); return index }
	layer := func(name string, sx, sy float64, height, y, step, repeats int) scrolling.SizeBankLayer {
		return scrolling.SizeBankLayer{Name: name, Image: font,
			CellWidth: 40, CellHeight: 32, Columns: 1, ScaleX: sx, ScaleY: sy,
			SurfaceHeight: height, RepeatY: y, RepeatStep: step, Repeats: repeats}
	}
	return scrolling.SizeBankConfig{
		Text: text, Controls: scrolltext.DomSizes, InitialFont: "0",
		ViewportWidth: 640, Lookup: tile,
		Layers: []scrolling.SizeBankLayer{
			layer("0", 1, 1, 32, 2, 36, 11),
			layer("1", 2, 2, 64, 2, 66, 6),
			layer("2", 4, 4, 128, 0, 134, 3),
			layer("3", 8, 12, 384, 4, 0, 1),
		},
		BaseSpeeds: []float64{8, 4, 2, 1}, StartOffset: 640,
		SpeedMultiplier: 1, Lookahead: 1, WrapInclusive: true,
	}, nil
}
