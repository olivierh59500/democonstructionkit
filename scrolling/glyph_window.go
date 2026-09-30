package scrolling

import "fmt"

// GlyphWindowConfig supplies a fixed visible glyph window from an authored
// transport. Glyph runs once per slot during drawing, returning the current
// character, font, pose or image; DCK owns layout storage and rendering. Advance
// is the fixed unscaled pen step. This adapter never advances the supplied clock.
type GlyphWindowConfig struct {
	Count   int
	Advance float64
	Glyph   func(slot int) Glyph
}

func prepareGlyphWindow(c *Config) error {
	w := c.GlyphWindow
	if w == nil {
		return nil
	}
	if c.Insertion != nil || c.Recycled != nil || c.RingLanes != nil || c.DualProfiled != nil || c.Caption != nil || c.Reveal != nil || c.Projected != nil || c.Pseudo3D != nil || c.Sliced != nil || c.CuedSlices != nil || c.Crawl != nil || c.Bands != nil || c.Slots != nil || c.Feed != nil || c.Scanline != nil || c.Profiled != nil || c.RowColumn != nil || c.RowBands != nil || c.SizeBank != nil || c.Ribbon != nil {
		return fmt.Errorf("scrolling: choose authored glyph window or another transport")
	}
	if w.Count < 1 || w.Count > 65536 || w.Count > c.MaxGlyphsPerDraw || !finite(w.Advance) || w.Advance <= 0 || !finite(float64(w.Count)*w.Advance) || w.Glyph == nil ||
		c.Glyphs != nil || c.Text != "" || len(c.Tokens) > 0 || c.Page != nil {
		return fmt.Errorf("scrolling: invalid authored glyph window")
	}
	copyConfig := *w
	c.GlyphWindow = &copyConfig
	c.Glyphs = make([]Glyph, w.Count)
	for i := range c.Glyphs {
		c.Glyphs[i] = Glyph{Advance: w.Advance}
	}
	return nil
}

func (s *Scrolling) sampleGlyphWindow() bool {
	if w := s.config.GlyphWindow; w != nil {
		for i := range s.glyphs {
			g := w.Glyph(i)
			g.Advance = w.Advance
			g.Offset = float64(i) * w.Advance
			if g.ScaleX == 0 {
				g.ScaleX = 1
			}
			if g.ScaleY == 0 {
				g.ScaleY = 1
			}
			if !finite(g.X) || !finite(g.Y) || !finite(g.ScaleX) || !finite(g.ScaleY) {
				s.drawErr = fmt.Errorf("scrolling: nonfinite authored glyph window pose")
				return false
			}
			s.glyphs[i] = g
		}
	}
	return true
}
