package presets

import (
	"image"
	"image/color"

	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// UnionTNTCaption leaves the messages and font art with the production while
// keeping the entrance/hold/exit clock and small black banner editable.
func UnionTNTCaption(lines []string, font scrolling.BitmapGrid) scrolling.CaptionCarouselConfig {
	return scrolling.CaptionCarouselConfig{
		Lines: lines, Font: font,
		Motion: motion.CaptionCycleConfig{Count: len(lines),
			Top: -18, Bottom: 0, StartY: -18, Speed: 2,
			InitialWait: 200, HoldWait: 100},
		Background: image.Rect(0, 0, 640, 18), BackgroundColor: color.Black,
		CenterX: 320, CenterStep: 8, MeasureBytes: true, ScaleX: 1, ScaleY: 1,
	}
}
