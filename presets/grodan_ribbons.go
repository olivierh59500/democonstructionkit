package presets

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// GrodanRibbons returns independently editable recipes for the large,
// vertical and two small texts. Text, font order and assets remain in the demo.
func GrodanRibbons(texts [4]string, big, vertical, small *scrolling.Atlas) [4]scrolling.RibbonConfig {
	horizontal := func(text string, font *scrolling.Atlas, speed, scaleX, scaleY, baselineY float64) scrolling.RibbonConfig {
		return scrolling.RibbonConfig{Text: text, Font: font, SkipMissing: true,
			Clock: motion.RibbonClockConfig{Velocity: -speed, Restart: 640,
				Multiplier: 1, Wrap: motion.RibbonWrapBelow},
			ScaleX: scaleX, ScaleY: scaleY, BaselineY: baselineY, CullAdvance: 40}
	}
	return [4]scrolling.RibbonConfig{
		horizontal(texts[0], big, 2, 8, 6, 0),
		{Text: texts[1], Font: vertical, Vertical: true,
			Clock: motion.RibbonClockConfig{Offset: -100, Velocity: 3, Restart: -100,
				Extent: 400, Multiplier: 1, Wrap: motion.RibbonWrapAbove},
			ScaleX: 1, ScaleY: 1, CullAdvance: 40},
		horizontal(texts[2], small, 1, 1, 1, 0),
		horizontal(texts[3], small, 2, 1, 1, 24),
	}
}
