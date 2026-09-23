package presets

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// BitmapRecipe returns a fresh configurable recipe for the Cuddly/Union atlases.
// Dimensions stay independent from the text, transport and image-effect layers.
func BitmapRecipe(id string) (scrolling.BitmapSpec, error) {
	spec := scrolling.BitmapSpec{First: 32, FractionalColumns: true}
	switch id {
	case "cuddly-colorshock":
		spec.Width, spec.Height = 70, 52
	case "cuddly-megascroller":
		spec.Width, spec.Height, spec.First = 240, 240, 31
	case "cuddly-bigsprite":
		spec.Width, spec.Height = 83.25, 41
	case "cuddly-fullscreen":
		spec.Width, spec.Height = 84, 80
	case "cuddly-knucklebuster", "cuddly-reset", "union-menu", "union-delta":
		spec.Width, spec.Height = 64, 34
	case "cuddly-doc":
		spec.Width, spec.Height = 62, 50
	case "cuddly-ehh-main", "union-replicants":
		spec.Width, spec.Height = 64, 64
	case "cuddly-ehh-middle", "union-level16":
		spec.Width, spec.Height = 32, 32
	case "cuddly-ehh-small":
		spec.Width, spec.Height = 16, 10
	case "cuddly-dna":
		spec.Width, spec.Height = 32, 25
	case "cuddly-dna-sine":
		spec.Width, spec.Height = 32, 16
	case "cuddly-digi":
		spec.Width, spec.Height = 64, 54
	case "cuddly-led":
		spec.Width, spec.Height = 128, 108
	case "cuddly-megaball":
		spec.Width, spec.Height = 320, 288
	case "cuddly-values":
		spec.Width, spec.Height = 8, 8
	case "cuddly-starwars-scroll":
		spec.Width, spec.Height = 32, 26
	case "cuddly-starwars-crawl":
		spec.Width, spec.Height = 17, 11
	case "cuddly-spreadpoint":
		spec.Width, spec.Height = 8, 6
	case "cuddly-reset-letters":
		spec.Width, spec.Height = 44, 44
	case "cuddly-loader", "union-loader":
		spec.Width, spec.Height = 16, 16
		spec.FractionalColumns = false
	case "union-beatdis":
		spec.Width, spec.Height = 96, 100
	case "union-wow":
		spec.Width, spec.Height = 192, 190
	case "union-starballs":
		spec.Width, spec.Height = 15, 8
	case "union-tnt2":
		spec.Width, spec.Height = 64, 40
	case "union-tnt3":
		spec.Width, spec.Height = 16, 18
	case "union-diskcopier":
		spec.Width, spec.Height = 16, 14
	default:
		return scrolling.BitmapSpec{}, fmt.Errorf("presets: unknown bitmap font %q", id)
	}
	return spec, nil
}
func BitmapFont(id string, atlas *ebiten.Image, filter ebiten.Filter) (scrolling.BitmapGrid, error) {
	spec, err := BitmapRecipe(id)
	if err != nil {
		return scrolling.BitmapGrid{}, err
	}
	return spec.Grid(atlas, filter)
}

// UnionMountainBands preserves the independent upper/lower mountain velocities.
// Change bands, motion scales or copy offsets before constructing a renderer.
func UnionMountainBands() composite.BandsConfig {
	c := composite.BandsConfig{CopyOffsets: [][2]float64{{0, 0}, {640, 0}}}
	for i := 0; i < 32; i++ {
		rank := i
		if rank >= 16 {
			rank = 31 - rank
		}
		y := i * 10
		if i >= 16 {
			y += 84
		}
		c.Bands = append(c.Bands, composite.MovingBand{Source: image.Rect(0, i*10, 1024, i*10+10), Y: float64(y), VelocityX: -(8 - float64(rank)*.5), WrapX: 256, MotionScaleX: 2})
	}
	return c
}

// CuddlyCrawlProjection returns the original perspective coverage profile.
// The compiled RowProjection accepts any live image, independently of its font.
func CuddlyCrawlProjection() []composite.Row {
	var rows []composite.Row
	previousY := 0.0
	for n := -160; n < 200; n++ {
		i := n + 160
		perspective := 250 / (250 + float64(n))
		y := math.Floor((float64(n)+350)*perspective + .5)
		if y != previousY {
			h := 1 - float64(i)/360
			srcY := float64(310 - i)
			if srcY >= 0 {
				scale := perspective * .8
				rows = append(rows, composite.Row{Source: composite.Region{Y: srcY, Width: 320, Height: h}, X: 160 - 160*scale, Y: y - h/2, Width: 320 * scale, Height: h, Filter: ebiten.FilterLinear})
			}
		}
		previousY = y
	}
	return rows
}

// CuddlyStarwarsCrawl is a complete bounded paragraph/perspective preset.
// The caller may change any paragraph, transport, projection or placement field.
func CuddlyStarwarsCrawl(font scrolling.BitmapGrid, lines []string) (scrolling.CrawlConfig, error) {
	projection, err := composite.NewRowProjection(CuddlyCrawlProjection())
	if err != nil {
		return scrolling.CrawlConfig{}, err
	}
	return scrolling.CrawlConfig{Paragraph: scrolling.BitmapParagraphConfig{Font: font, Lines: lines, Width: 300, LineAdvance: 17, GlyphAdvance: 20, AlignmentAdvance: 16, OffsetX: -8, Align: scrolling.AlignCenter, MaxCharacters: 12}, Width: 320, Height: 420, ProjectionWidth: 320, ProjectionHeight: 400, VisibleLines: 30, PixelsPerUpdate: 1, Projection: projection, Output: composite.Region{Y: 300, Width: 320, Height: 100}, Y: 100, Filter: ebiten.FilterLinear}, nil
}

// UnionCreditsReveal supplies the column-first entrance used between screens.
// Its clock uses the original units; no assumptions about the host's TPS apply.
func UnionCreditsReveal(font scrolling.BitmapGrid, lines []string) scrolling.RevealConfig {
	columns := 20
	if len(lines) > 0 {
		columns = len([]rune(lines[0]))
	}
	start := 500.0
	return scrolling.RevealConfig{Font: font, Lines: lines, Columns: columns, X: 154, Y: 374 - float64(len(lines)-1)*16, AdvanceX: 16, AdvanceY: 16, UniformStartY: &start, OutputY: -8, Delay: 30, Duration: 50, Order: scrolling.RevealColumnsBottomFirst}
}

// CuddlySpreadpointBands keeps the entrance and independent lane speeds while
// retaining only one viewport-sized text surface, regardless of message length.
func CuddlySpreadpointBands(font scrolling.BitmapGrid, text string) scrolling.BitmapBandsConfig {
	c := scrolling.BitmapBandsConfig{Font: font, Text: text, Width: 320, Height: 200, Entrance: 320, Repeat: true, UseTicks: true, Filter: ebiten.FilterNearest}
	speeds := []int{12, 11, 10, 9, 8, 7, 6, 5, 4, 5, 6, 7, 8, 9, 10, 11, 12, 11, 10, 9, 8, 7, 6, 5, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	for i, speed := range speeds {
		c.Lanes = append(c.Lanes, scrolling.BitmapLane{Y: float64(i * 6), Speed: float64(speed)})
	}
	return c
}

// CuddlyResetSlots configures a recycled text train with the historical
// preceding-glyph tangent. Font dimensions and all motion settings stay editable.
func CuddlyResetSlots(font scrolling.BitmapGrid, text string) scrolling.BitmapSlotsConfig {
	return scrolling.BitmapSlotsConfig{Font: font, Text: text, Count: 24, Start: 24 * 44, Advance: 44, Speed: 3, RecycleBelow: -88, Period: 24 * 44, Y: 250, PreviousTangent: true, Waves: []scrolling.RingWave{{Amplitude: 50, LetterStep: -.5, TickStep: .05}}}
}
