package scrolling

import (
	"fmt"
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
)

// BitmapPageConfig builds one retained image from borrowed glyph images.
// Order may use any alphabet layout and Lines retain explicit source spacing.
// A nil background leaves the page transparent.
type BitmapPageConfig struct {
	Width, Height int
	Order         string
	Glyphs        []*ebiten.Image
	Lines         []BitmapPageLine
	Uppercase     bool
	Background    color.Color
	Filter        ebiten.Filter
	Blend         ebiten.Blend
}

// BitmapPage owns only its completed image. The glyph bank remains caller-owned.
type BitmapPage struct{ image *ebiten.Image }

func NewBitmapPage(config BitmapPageConfig) (*BitmapPage, error) {
	if config.Width < 1 || config.Height < 1 || config.Width > 8192 || config.Height > 8192 ||
		len(config.Glyphs) != utf8.RuneCountInString(config.Order) {
		return nil, fmt.Errorf("scrolling: invalid bitmap page dimensions or glyph bank")
	}
	for _, glyph := range config.Glyphs {
		if glyph == nil {
			return nil, fmt.Errorf("scrolling: nil bitmap page glyph")
		}
	}
	placements, err := CompileBitmapPageLayout(config.Order, config.Lines, config.Uppercase)
	if err != nil {
		return nil, err
	}
	page := &BitmapPage{image: ebiten.NewImage(config.Width, config.Height)}
	if config.Background != nil {
		page.image.Fill(config.Background)
	}
	for _, placement := range placements {
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = config.Filter, config.Blend
		op.GeoM.Scale(placement.ScaleX, placement.ScaleY)
		op.GeoM.Translate(placement.X, placement.Y)
		page.image.DrawImage(config.Glyphs[placement.Index], &op)
	}
	return page, nil
}

// Image is borrowed until the next Close; drawing it never changes the page.
func (page *BitmapPage) Image() *ebiten.Image {
	if page == nil {
		return nil
	}
	return page.image
}

func (page *BitmapPage) Draw(dst *ebiten.Image) {
	if page != nil && page.image != nil && dst != nil {
		dst.DrawImage(page.image, nil)
	}
}

func (page *BitmapPage) Close() error {
	if page != nil && page.image != nil {
		page.image.Deallocate()
		page.image = nil
	}
	return nil
}
