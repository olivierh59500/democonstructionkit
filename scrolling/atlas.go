package scrolling

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
)

// Atlas caches the image views of an immutable bitmap font. It borrows the image;
// callers retain ownership, and independent scrollers may share the same Atlas.
type Atlas struct {
	face   Face
	images map[image.Rectangle]*ebiten.Image
}

// NewAtlas validates the atlas and creates each distinct image view once.
// A nil image is allowed for metric-only initialization and layout inspection.
func NewAtlas(img *ebiten.Image, metrics *font.Font) (*Atlas, error) {
	if metrics == nil || (img != nil && !metrics.Bounds().In(img.Bounds())) {
		return nil, fmt.Errorf("scrolling: atlas metrics must lie inside the image")
	}
	a := &Atlas{face: Face{Atlas: img, Metrics: metrics}, images: make(map[image.Rectangle]*ebiten.Image)}
	if img != nil {
		for _, r := range metrics.Characters() {
			g, _ := metrics.ExactGlyph(r)
			if !g.Rect.Empty() && a.images[g.Rect] == nil {
				a.images[g.Rect] = img.SubImage(g.Rect).(*ebiten.Image)
			}
		}
	}
	return a, nil
}

// Glyph applies the font's aliases, case conversion, fallback and blank rules.
// The image is shared and must not be deallocated independently of its atlas.
func (a *Atlas) Glyph(r rune) (*ebiten.Image, font.Glyph, bool) {
	g, ok := a.face.Metrics.Glyph(r)
	return a.images[g.Rect], g, ok
}

// ExactGlyph selects a literal atlas entry, preserving case-sensitive transports.
// Missing entries return zero metrics, unlike Glyph's configured fallback advance.
func (a *Atlas) ExactGlyph(r rune) (*ebiten.Image, font.Glyph, bool) {
	g, ok := a.face.Metrics.ExactGlyph(r)
	return a.images[g.Rect], g, ok
}

// Metrics returns the immutable font metrics.
func (a *Atlas) Metrics() *font.Font { return a.face.Metrics }

// Face returns a face suitable for Config.Fonts.
func (a *Atlas) Face() Face { return a.face }

// Glyphs compiles a single text ribbon, preserving blanks and unsupported advances.
// Newlines are treated as ordinary characters; use Config.Page for paragraphs.
func (a *Atlas) Glyphs(text string) []Glyph {
	return a.Layout(text, AtlasText{})
}

// AtlasText selects legacy-compatible text layout without redefining a font.
type AtlasText struct {
	Literal     bool // Bypass aliases, case conversion and fallback.
	SkipMissing bool // Omit unsupported characters rather than advancing a blank.
	Vertical    bool // Advance by line height rather than each glyph's width.
}

// Layout compiles a cached-image ribbon with resolved offsets. It performs no
// image allocation and is intended for initialization or text changes.
func (a *Atlas) Layout(text string, options AtlasText) []Glyph {
	glyphs := make([]Glyph, 0, len(text))
	lookup := a.Glyph
	if options.Literal {
		lookup = a.ExactGlyph
	}
	offset := 0.0
	for _, r := range text {
		img, g, ok := lookup(r)
		if !ok && options.SkipMissing {
			continue
		}
		advance := g.Advance
		if options.Vertical {
			advance = a.Metrics().LineHeight()
		}
		glyphs = append(glyphs, Glyph{Image: img, Rune: r, Advance: advance, Offset: offset, X: g.OffsetX, Y: g.OffsetY})
		offset += advance
	}
	return glyphs
}

// GridImages caches rectangular atlas cells in row-major order. Coordinates are
// relative to image.Bounds().Min; count can include deliberately unused cells.
func GridImages(img *ebiten.Image, cell image.Point, columns, count int) ([]*ebiten.Image, error) {
	if img == nil || cell.X <= 0 || cell.Y <= 0 || columns <= 0 || count < 0 {
		return nil, fmt.Errorf("scrolling: invalid image grid")
	}
	if columns > img.Bounds().Dx()/cell.X || count > columns*(img.Bounds().Dy()/cell.Y) {
		return nil, fmt.Errorf("scrolling: image grid exceeds atlas bounds")
	}
	result := make([]*ebiten.Image, count)
	for i := range result {
		p := img.Bounds().Min.Add(image.Pt(i%columns*cell.X, i/columns*cell.Y))
		r := image.Rectangle{Min: p, Max: p.Add(cell)}
		if !r.In(img.Bounds()) {
			return nil, fmt.Errorf("scrolling: grid cell %d is outside the image", i)
		}
		result[i] = img.SubImage(r).(*ebiten.Image)
	}
	return result, nil
}
