package font

import (
	"fmt"
	"image"
	"image/color"
	"unicode/utf8"
)

// Cell is one lit bitmap pixel in glyph-local coordinates. Colors are straight
// alpha; a missing color reader assigns opaque white.
type Cell struct {
	X, Y  int
	Color color.NRGBA
}

// CellBankConfig prepares a bitmap for geometric pixel-cell rendering. Pixel
// and Color run only during construction. Key optionally shares identical
// bitmap/color data between aliases; characters with the same key must have
// the same artwork. Dimensions and character order are caller supplied.
type CellBankConfig struct {
	Width, Height int
	Characters    string
	Pixel         func(character rune, x, y int) bool
	Color         func(character rune, x, y int) color.NRGBA
	Key           func(character rune) int
}

// CellBank owns immutable, row-major lit-cell lists. Glyph returns a borrowed
// read-only list; no pixels are scanned or allocated during rendering.
type CellBank struct {
	width, height int
	glyphs        map[rune][]Cell
}

func NewCellBank(c CellBankConfig) (*CellBank, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 4096 || c.Height > 4096 ||
		int64(c.Width)*int64(c.Height) > 1<<20 || c.Characters == "" ||
		!utf8.ValidString(c.Characters) || utf8.RuneCountInString(c.Characters) > 65536 || c.Pixel == nil {
		return nil, fmt.Errorf("font: invalid geometric cell font")
	}
	b := &CellBank{width: c.Width, height: c.Height, glyphs: make(map[rune][]Cell)}
	cache := make(map[int][]Cell)
	work, count := int64(0), 0
	for _, character := range c.Characters {
		if _, exists := b.glyphs[character]; exists {
			return nil, fmt.Errorf("font: duplicate cell character %q", character)
		}
		key := int(character)
		if c.Key != nil {
			key = c.Key(character)
		}
		cells, exists := cache[key]
		if !exists {
			work += int64(c.Width) * int64(c.Height)
			if work > 1<<24 {
				return nil, fmt.Errorf("font: geometric font exceeds its sampling budget")
			}
			for y := 0; y < c.Height; y++ {
				for x := 0; x < c.Width; x++ {
					if !c.Pixel(character, x, y) {
						continue
					}
					paint := color.NRGBA{255, 255, 255, 255}
					if c.Color != nil {
						paint = c.Color(character, x, y)
					}
					count++
					if count > 1<<20 {
						return nil, fmt.Errorf("font: too many geometric font cells")
					}
					cells = append(cells, Cell{X: x, Y: y, Color: paint})
				}
			}
			cache[key] = cells
		}
		b.glyphs[character] = cells
	}
	return b, nil
}

func (b *CellBank) Width() int  { return b.width }
func (b *CellBank) Height() int { return b.height }
func (b *CellBank) Glyph(character rune) []Cell {
	if b == nil {
		return nil
	}
	return b.glyphs[character]
}

// NewImageCellBank prepares a decoded CPU atlas using independent font metrics,
// aliases, proportional advances and the requested character set. A zero alpha
// threshold selects 32768. Source colors are retained in the cached cells.
func NewImageCellBank(source image.Image, metrics *Font, characters string, threshold uint16) (*CellBank, error) {
	if source == nil || metrics == nil || !metrics.Bounds().In(source.Bounds()) {
		return nil, fmt.Errorf("font: invalid geometric font atlas or metrics")
	}
	if threshold == 0 {
		threshold = 32768
	}
	rects := make(map[rune]image.Rectangle)
	keys := make(map[image.Rectangle]int)
	width, height := 1, 1
	for _, r := range characters {
		g, _ := metrics.Glyph(r)
		rects[r] = g.Rect
		width, height = max(width, g.Rect.Dx()), max(height, g.Rect.Dy())
		if _, exists := keys[g.Rect]; !exists {
			keys[g.Rect] = len(keys)
		}
	}
	return NewCellBank(CellBankConfig{
		Width: width, Height: height, Characters: characters,
		Key: func(r rune) int { return keys[rects[r]] },
		Pixel: func(r rune, x, y int) bool {
			bounds := rects[r]
			if x >= bounds.Dx() || y >= bounds.Dy() || bounds.Empty() {
				return false
			}
			_, _, _, alpha := source.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			return alpha >= uint32(threshold)
		},
		Color: func(r rune, x, y int) color.NRGBA {
			p := rects[r].Min.Add(image.Pt(x, y))
			return color.NRGBAModel.Convert(source.At(p.X, p.Y)).(color.NRGBA)
		},
	})
}
