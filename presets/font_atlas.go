package presets

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// FontAtlas builds and caches an audited font recipe against the supplied image.
// A nil image builds metric-only proportional recipes, useful for layout tools.
func FontAtlas(id string, img *ebiten.Image) (*scrolling.Atlas, error) {
	spec, ok := FindFont(id)
	if !ok {
		return nil, fmt.Errorf("presets: unknown font %q", id)
	}
	bounds := image.Rect(0, 0, spec.Cell.X*spec.Columns, spec.Cell.Y*max(6, (59+spec.Columns-1)/spec.Columns))
	if img != nil {
		bounds = img.Bounds()
	}
	metrics, err := spec.Build(bounds)
	if err != nil {
		return nil, err
	}
	return scrolling.NewAtlas(img, metrics)
}

// TileLookup compiles a preset's integer character indices once. Literal mode
// bypasses case folding and aliases; empty glyphs never refer to a source tile.
func TileLookup(id string, literal bool) (func(rune) (int, bool), error) {
	a, err := FontAtlas(id, nil)
	if err != nil {
		return nil, err
	}
	s, _ := FindFont(id)
	metrics := a.Metrics()
	lookup := metrics.Glyph
	if literal {
		lookup = metrics.ExactGlyph
	}
	return func(r rune) (int, bool) {
		g, ok := lookup(r)
		if !ok || g.Rect.Empty() {
			return 0, false
		}
		p := g.Rect.Min.Sub(metrics.Bounds().Min)
		return p.Y/s.Cell.Y*s.Columns + p.X/s.Cell.X, true
	}, nil
}
