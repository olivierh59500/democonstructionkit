package presets

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// NonamenoScrollOptions controls the original bottom ribbon without tying it
// to one screen size or font. Gap is the blank distance between consecutive
// messages; zero gives a continuous loop. The source used Width+1 pixels.
type NonamenoScrollOptions struct {
	Width, BaselineY, TicksPerSecond, PixelsPerTick, Gap float64
}

// NonamenoBottomWaves returns two independently editable glyph-space sine
// waves. Spatial is radians per pixel, and Speed is radians per second.
func NonamenoBottomWaves(advance, ticksPerSecond float64) []motion.Wave {
	return []motion.Wave{
		{Amplitude: 15, Spatial: .06 / advance, Speed: .3 * ticksPerSecond},
		{Amplitude: 15, Spatial: -.04 / advance, Speed: .2 * ticksPerSecond},
	}
}

// NonamenoBottomScroll compiles the ribbon into the same scrolling.Config
// accepted by scrolling.New. Motion is based on simulation ticks, not the
// display refresh rate. A gap of Width+1 retains the source's strict reset;
// using zero lets the next copy follow the last glyph without a blank interval.
func NonamenoBottomScroll(text string, face scrolling.Face, c NonamenoScrollOptions) (scrolling.Config, error) {
	if text == "" || face.Atlas == nil || face.Metrics == nil ||
		!finiteNonamenoScroll(c.Width) || c.Width <= 0 || !finiteNonamenoScroll(c.BaselineY) ||
		!finiteNonamenoScroll(c.TicksPerSecond) || c.TicksPerSecond <= 0 ||
		!finiteNonamenoScroll(c.PixelsPerTick) || c.PixelsPerTick <= 0 ||
		!finiteNonamenoScroll(c.Gap) || c.Gap < 0 {
		return scrolling.Config{}, fmt.Errorf("presets: invalid Nonameno bottom scroll")
	}
	advance := face.Metrics.LineHeight()
	length := float64(utf8.RuneCountInString(text)) * advance
	if !finiteNonamenoScroll(advance) || advance <= 0 || !finiteNonamenoScroll(length) ||
		!finiteNonamenoScroll(c.Width+length+c.Gap) || !finiteNonamenoScroll(c.PixelsPerTick*c.TicksPerSecond) {
		return scrolling.Config{}, fmt.Errorf("presets: invalid Nonameno font advance or length")
	}
	waves, err := scrolling.HarmonicSineWith(scrolling.HarmonicSineConfig{
		Direction: scrolling.WaveVertical, SpatialPeriod: length + c.Gap,
		Waves: NonamenoBottomWaves(advance, c.TicksPerSecond),
	})
	if err != nil {
		return scrolling.Config{}, err
	}
	return scrolling.Config{
		Text: text, Fonts: map[string]scrolling.Face{"default": face},
		Advance: advance, Speed: c.PixelsPerTick * c.TicksPerSecond,
		Gap: c.Gap, X: c.Width, Y: c.BaselineY, Repeat: true,
		Modes: map[string]scrolling.Mode{"two-wave": waves}, Shape: "two-wave",
	}, nil
}

func finiteNonamenoScroll(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
