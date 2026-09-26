package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CocoTitleBand keeps the autonomous and embedded presentations on the same
// title/copper effect. The autonomous surface uses a fractional wrap clock;
// the embedded direct mode retains its integer masked clock and zero extra
// full-width texture.
func CocoTitleBand(title, bars *ebiten.Image, width int, mode composite.CopperTitleMode) composite.CopperTitleBandConfig {
	var copper *composite.CopperBarsConfig
	if bars != nil {
		clock := composite.SingleWrapClock
		if mode == composite.CopperTitleDirect {
			clock = composite.MaskedClock
		}
		config := BilizirCopperBars(bars, 72, composite.CopperImages, clock)
		copper = &config
	}
	return composite.CopperTitleBandConfig{
		Title: title, Copper: copper, TitleMotion: CocoTitleMotion(float64(width)),
		Width: width, Height: 72, Mode: mode, Background: color.Black,
		TitleScaleX: 1, TitleFilter: ebiten.FilterNearest,
	}
}
