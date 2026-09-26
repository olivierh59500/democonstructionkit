package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// PhenomenaIntroLine keeps the source's fixed 32-pixel pen, 2x glyph size and
// 48-pixel left anchor while leaving each text and vertical position authored.
func PhenomenaIntroLine(y float64, text string) scrolling.BitmapPageLine {
	return scrolling.BitmapPageLine{Text: text, X: 48, Y: y, Advance: 32, ScaleX: 2, ScaleY: 2}
}

// PhenomenaIntroPage binds either normal or inverted glyph bank to one retained
// 640x480 page. The caller chooses its background and ordered text lines.
func PhenomenaIntroPage(glyphs []*ebiten.Image, lines []scrolling.BitmapPageLine, background color.Color) scrolling.BitmapPageConfig {
	return scrolling.BitmapPageConfig{
		Width: 640, Height: 480, Order: PhenomenaAlphabet,
		Glyphs: glyphs, Lines: lines, Background: background,
	}
}
