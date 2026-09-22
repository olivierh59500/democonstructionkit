package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// checkRepeatRendering compares automatic repetition with an independent CPU
// ribbon. It runs inside RunGame so ReadPixels observes actual GPU output.
func checkRepeatRendering() error {
	const width, height = 19, 13
	red := ebiten.NewImage(3, 3)
	defer red.Deallocate()
	red.Fill(color.RGBA{R: 255, A: 255})
	blue := ebiten.NewImage(6, 2)
	defer blue.Deallocate()
	blue.Fill(color.RGBA{B: 255, A: 255})
	dst := ebiten.NewImage(width, height)
	defer dst.Deallocate()
	pixels := make([]byte, width*height*4)
	expected := make([]byte, len(pixels))
	cases := []struct {
		name   string
		config scrolling.Config
		times  []float64
	}{
		{
			name: "horizontal overhang and gap",
			config: scrolling.Config{
				Glyphs: []scrolling.Glyph{{Image: red, Advance: 4, X: -1}, {Image: blue, Advance: 3}},
				X:      21, Y: 3, Speed: 20, Gap: 2, Repeat: true,
			},
			times: []float64{0, .1, .4, .449999, .45, .450001, .9, 2.25, 1000},
		},
		{
			name: "short fractional-duration horizontal ribbon",
			config: scrolling.Config{
				Glyphs: []scrolling.Glyph{{Image: red, Advance: 3}, {Image: blue, Advance: 4}},
				X:      21, Y: 4, Speed: 80, Gap: 1, Repeat: true,
			},
			times: []float64{0, .1, .4875, .5, .5125, 10, 1000},
		},
		{
			name: "vertical mixed glyphs",
			config: scrolling.Config{
				Glyphs: []scrolling.Glyph{{Image: red, Advance: 3, Y: -1}, {Image: blue, Advance: 4, X: 1}},
				X:      5, Y: 17, Speed: 80, Gap: 1, Vertical: true, Repeat: true,
			},
			times: []float64{0, .1, .4875, .5, .5125, 10, 1000},
		},
	}
	for _, tc := range cases {
		scroll, err := scrolling.New(tc.config)
		if err != nil {
			return err
		}
		for _, seconds := range tc.times {
			dst.Clear()
			if err := scroll.Update(kit.Frame{Time: seconds}); err != nil {
				return err
			}
			scroll.Draw(dst)
			dst.ReadPixels(pixels)
			clear(expected)
			paintExpectedRibbon(expected, width, height, tc.config, seconds)
			for i, want := range expected {
				if pixels[i] != want {
					return fmt.Errorf("automatic repeat %s at %.9g seconds: pixel (%d,%d), channel %d = %d, want %d", tc.name, seconds, i/4%width, i/4/width, i%4, pixels[i], want)
				}
			}
		}
	}
	fmt.Println("OK automatic repeat: entry, seams, short mixed glyphs, overhang, vertical and later loops")
	return nil
}

// paintExpectedRibbon never wraps time: every copy starts after the preceding
// message plus its gap, and all copies travel at the configured constant speed.
func paintExpectedRibbon(pixels []byte, width, height int, c scrolling.Config, seconds float64) {
	period := c.Gap
	for _, glyph := range c.Glyphs {
		period += glyph.Advance
	}
	distance := seconds * c.Speed
	origin, extent := c.X, float64(width)
	if c.Vertical {
		origin, extent = c.Y, float64(height)
	}
	// These fixtures have bearings and overhangs smaller than one period. Start
	// one copy early; clipping below independently discards its offscreen pixels.
	first := max(0, int(math.Floor((distance-origin)/period))-1)
	last := int(math.Ceil((distance-origin+extent)/period)) + 1
	for copy := first; copy <= last; copy++ {
		offset := float64(copy) * period
		for index, glyph := range c.Glyphs {
			x, y := c.X+glyph.X, c.Y+glyph.Y
			if c.Vertical {
				y += offset - distance
			} else {
				x += offset - distance
			}
			rgba := [4]byte{255, 0, 0, 255}
			if index == 1 {
				rgba = [4]byte{0, 0, 255, 255}
			}
			x0 := max(0, int(math.Ceil(x-.5)))
			y0 := max(0, int(math.Ceil(y-.5)))
			x1 := min(width, int(math.Ceil(x+float64(glyph.Image.Bounds().Dx())-.5)))
			y1 := min(height, int(math.Ceil(y+float64(glyph.Image.Bounds().Dy())-.5)))
			for py := y0; py < y1; py++ {
				for px := x0; px < x1; px++ {
					for channel, value := range rgba {
						pixels[(py*width+px)*4+channel] = value
					}
				}
			}
			offset += glyph.Advance
		}
	}
}
