package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// CyclicWindowConfig selects a bounded part of repeated glyph positions.
// Minimum and Maximum are destination pen coordinates; Copies controls the
// maximum number of virtual text cycles considered. Scale applies on the text
// transport axis; the other image scale can be edited on the returned state.
type CyclicWindowConfig struct {
	Scale, Minimum, Maximum float64
	Copies                  int
}

// CyclicWindow borrows one immutable Scrolling layout and prepares a visible
// DrawState in O(log glyph count), without a message-width image or allocation.
type CyclicWindow struct {
	scroll               *Scrolling
	config               CyclicWindowConfig
	count                int
	minOffset, maxOffset float64
	clip                 Mapper
}

func NewCyclicWindow(scroll *Scrolling, config CyclicWindowConfig) (*CyclicWindow, error) {
	if scroll == nil || scroll.backend != nil || len(scroll.glyphs) == 0 ||
		config.Copies < 1 || config.Copies > (1<<20)/len(scroll.glyphs) ||
		!finite(config.Scale) || config.Scale <= 0 ||
		!finite(config.Minimum) || !finite(config.Maximum) || config.Minimum >= config.Maximum {
		return nil, fmt.Errorf("scrolling: invalid cyclic window layout or bounds")
	}
	window := &CyclicWindow{scroll: scroll, config: config, count: len(scroll.glyphs) * config.Copies}
	window.minOffset, window.maxOffset = math.Inf(1), math.Inf(-1)
	for _, glyph := range scroll.glyphs {
		offset := glyph.X
		if scroll.config.Vertical {
			offset = glyph.Y
		}
		window.minOffset = math.Min(window.minOffset, offset)
		window.maxOffset = math.Max(window.maxOffset, offset)
	}
	if !finite(window.minOffset*config.Scale) || !finite(window.maxOffset*config.Scale) {
		return nil, fmt.Errorf("scrolling: cyclic glyph offset overflow")
	}
	if scroll.config.Vertical {
		window.clip = func(sample Sample, _ *ebiten.DrawImageOptions) bool {
			return sample.Y > config.Minimum && sample.Y < config.Maximum
		}
	} else {
		window.clip = func(sample Sample, _ *ebiten.DrawImageOptions) bool {
			return sample.X > config.Minimum && sample.X < config.Maximum
		}
	}
	return window, nil
}

// At returns only virtual glyphs whose pen origins can enter the viewport,
// including the largest authored glyph offset in either direction. Nonfinite
// origins produce an empty safe state. DrawAt does not advance this window.
func (window *CyclicWindow) At(origin float64) DrawState {
	state := IdentityState()
	state.Cycle, state.bounded = true, true
	if window == nil || !finite(origin) {
		state.End = 0
		return state
	}
	config := window.config
	if window.scroll.config.Vertical {
		state.Y, state.ScaleY = origin, config.Scale
	} else {
		state.X, state.ScaleX = origin, config.Scale
	}
	minimum := config.Minimum - window.maxOffset*config.Scale
	maximum := config.Maximum - window.minOffset*config.Scale
	state.First, state.End = cyclicWindowIndices(window.scroll, window.count, origin, config.Scale, minimum, maximum)
	state.Map = window.clip
	return state
}
