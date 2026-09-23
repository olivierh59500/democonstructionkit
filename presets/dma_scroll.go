package presets

import (
	"image/color"

	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// DMAScanlineScroll is an editable recipe for the proportional blue-background
// ribbon. Supply any atlas with suitable glyph metrics and any literal message.
func DMAScanlineScroll(font *scrolling.Atlas, text string) (scrolling.ScanlineConfig, error) {
	curves, err := RibbonCurves(1)
	if err != nil {
		return scrolling.ScanlineConfig{}, err
	}
	wave, err := composite.JoinDeltaCurves(curves[:8], []int{
		1, 1, 4, 1, 1, 2, 3, 2, 1, 5, 2, 1, 7,
	})
	if err != nil {
		return scrolling.ScanlineConfig{}, err
	}
	return scrolling.ScanlineConfig{
		Font: font, Text: text, Wave: wave,
		ViewportWidth: 640, ViewportHeight: 400, SurfaceWidth: 1280,
		SourceRows: 36, RowHeight: 3, StripHeight: 3,
		Scale: 3, WaveStep: 10, CursorRows: 36,
		BounceAmplitude: 18, BounceRate: .1,
		ClampCursor: true, Wrap: scrolling.ScanlineSplit,
		Background: color.RGBA{0x00, 0x00, 0x60, 0xff},
	}, nil
}

// DMAIntroFeed compiles the same finite text trail for any compatible atlas.
func DMAIntroFeed(font *scrolling.Atlas, text string) scrolling.FeedConfig {
	return scrolling.FeedConfig{Font: font, Text: text, Width: 640, Height: 72, Margin: 96, Scale: 2, Speed: 8}
}

// CocoScanlineScroll applies the same displacement program with Coco's
// one-pixel strips, repeated GPU addressing and variable-speed scene clock.
func CocoScanlineScroll(font *scrolling.Atlas, text string) (scrolling.ScanlineConfig, error) {
	c, err := DMAScanlineScroll(font, text)
	if err != nil {
		return c, err
	}
	c.ViewportWidth = 800
	c.ViewportHeight = 600 - 72
	c.SurfaceWidth = 1600
	c.DestinationY = 72
	c.StripHeight = 1
	c.WaveStep = 15
	c.MissingAdvance = 32
	c.ClampCursor = false
	c.Wrap = scrolling.ScanlineAddressRepeat
	c.AlternateDiagonal = true
	c.Background = nil
	c.UseTime = true
	c.SurfaceUnmanaged = true
	return c, nil
}

// CocoIntroFeed uses the same persistent-glyph transport at a wider viewport.
func CocoIntroFeed(font *scrolling.Atlas, text string) scrolling.FeedConfig {
	return scrolling.FeedConfig{Font: font, Text: text, Width: 800, Height: 72, Margin: 96, Scale: 2, Speed: 8}
}

// TeamG1IntroFeed retains its tighter history surface and uppercase lookup.
func TeamG1IntroFeed(font *scrolling.Atlas, text string) scrolling.FeedConfig {
	return scrolling.FeedConfig{Font: font, Text: text, Width: 768, Height: 72, InsertX: 640, Scale: 2, Speed: 6, UppercaseASCII: true}
}

// MegaTwistIntroFeed keeps its unscaled 36-pixel ribbon and 48-pixel tail.
func MegaTwistIntroFeed(font *scrolling.Atlas, text string) scrolling.FeedConfig {
	return scrolling.FeedConfig{Font: font, Text: text, Width: 416, Height: 36, Margin: 48, Scale: 1, Speed: 4}
}

// MegaTwistPrograms compiles the independent foreground and background wave
// timelines, including their one-time introductions and continuing loops.
func MegaTwistPrograms(rate float64) (front, back *composite.DisplacementProgram, err error) {
	curves, err := RibbonCurves(rate)
	if err != nil {
		return nil, nil, err
	}
	join := func(order ...int) ([]int, error) { return composite.JoinDeltaCurves(curves, order) }
	frontIntro, err := join(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 2, 1, 7)
	if err != nil {
		return nil, nil, err
	}
	frontMain, err := join(1, 1, 4, 1, 1, 2, 3, 2, 1, 5, 2, 1, 7)
	if err != nil {
		return nil, nil, err
	}
	backIntro, err := join(0, 0, 0, 0, 0)
	if err != nil {
		return nil, nil, err
	}
	backMain, err := join(8, 8, 9, 9, 10, 10, 8, 8, 9, 9, 10, 10, 8, 8, 9, 9, 10, 10, 7)
	if err != nil {
		return nil, nil, err
	}
	front, err = composite.NewDisplacementProgram(frontIntro, frontMain)
	if err != nil {
		return nil, nil, err
	}
	back, err = composite.NewDisplacementProgram(backIntro, backMain)
	return front, back, err
}

// MegaTwistScanlineScroll uses the foreground wave and its strict X-window
// rejection rule. Font metrics and messages remain supplied by the production.
func MegaTwistScanlineScroll(font *scrolling.Atlas, text string, program *composite.DisplacementProgram) scrolling.ScanlineConfig {
	return scrolling.ScanlineConfig{
		Font: font, Text: text, Program: program,
		ViewportWidth: 416, ViewportHeight: 276, SurfaceWidth: (416*8 + 4) / 5,
		SourceRows: 36, RowHeight: 1, StripHeight: 1,
		Scale: 1, WaveStep: 10, CursorRows: 276,
		BounceAmplitude: 18, BounceRate: .1,
		MissingRune: ' ', ClampCursor: true,
		Wrap: scrolling.ScanlineReject, AlternateDiagonal: true,
	}
}

// DMALogoGrid retains the finite two-by-four tile choreography and overlap.
func DMALogoGrid() composite.ImageGrid {
	return composite.ImageGrid{Columns: 2, Rows: 4, StepX: 640, StepY: 200}
}

// DMACRTOverlay is the editable normalized CRT recipe for the intro ribbon.
func DMACRTOverlay() effects.CRTOverlayConfig {
	return effects.CRTOverlayConfig{
		Curvature: .15, ScanlineFrequency: 800, ScanlineAmplitude: .04,
		ChromaticShift: .002, Vignette: .5,
	}
}
