package scrolling

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type CaptionCarouselConfig struct {
	Lines               []string
	Font                BitmapGrid
	Motion              motion.CaptionCycleConfig
	Background          image.Rectangle
	BackgroundColor     color.Color
	CenterX, CenterStep float64
	MeasureBytes        bool
	ScaleX, ScaleY      float64
}

type captionGlyph struct {
	region composite.Region
	index  int
}

type captionLine struct {
	glyphs      []captionGlyph
	centerWidth float64
}

// CaptionCarousel caches font regions once and borrows the background image.
// Update prepares the current caption before advancing the clock, so the
// usual Update-then-Draw order retains the first source pose and hold bounds.
type CaptionCarousel struct {
	config         CaptionCarouselConfig
	motion         *motion.CaptionCycle
	prepared       motion.CaptionPose
	lines          []captionLine
	lastDst        *ebiten.Image
	backgroundView *ebiten.Image
}

func NewCaptionCarousel(c CaptionCarouselConfig) (*CaptionCarousel, error) {
	if len(c.Lines) == 0 || c.Font.Image == nil || c.Background.Empty() ||
		c.BackgroundColor == nil || !finiteCaption(c.CenterX) ||
		!finiteCaption(c.CenterStep) || !finiteCaption(c.ScaleX) ||
		!finiteCaption(c.ScaleY) || c.ScaleX == 0 || c.ScaleY == 0 {
		return nil, fmt.Errorf("scrolling: invalid caption carousel")
	}
	if c.Motion.Count == 0 {
		c.Motion.Count = len(c.Lines)
	}
	if c.Motion.Count != len(c.Lines) {
		return nil, fmt.Errorf("scrolling: caption motion and line count differ")
	}
	clock, err := motion.NewCaptionCycle(c.Motion)
	if err != nil {
		return nil, err
	}
	if c.CenterStep == 0 {
		c.CenterStep = c.Font.Width * c.ScaleX / 2
	}
	c.Lines = append([]string(nil), c.Lines...)
	result := &CaptionCarousel{config: c, motion: clock, prepared: clock.Pose(),
		lines: make([]captionLine, len(c.Lines))}
	for i, line := range c.Lines {
		runes := []rune(line)
		count := len(runes)
		if c.MeasureBytes {
			count = len(line)
		}
		result.lines[i].centerWidth = float64(count) * c.CenterStep
		for index, ch := range runes {
			region, ok := c.Font.Region(ch)
			if ok {
				result.lines[i].glyphs = append(result.lines[i].glyphs,
					captionGlyph{region: region, index: index})
			}
		}
	}
	return result, nil
}

func finiteCaption(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (c *CaptionCarousel) Draw(dst *ebiten.Image) {
	if c == nil || dst == nil {
		return
	}
	if c.lastDst != dst {
		clip := c.config.Background.Intersect(dst.Bounds())
		c.backgroundView = nil
		if !clip.Empty() {
			c.backgroundView = dst.SubImage(clip).(*ebiten.Image)
		}
		c.lastDst = dst
	}
	if c.backgroundView != nil {
		c.backgroundView.Fill(c.config.BackgroundColor)
	}
	pose := c.prepared
	line := c.lines[pose.Index]
	x := c.config.CenterX - line.centerWidth
	for _, glyph := range line.glyphs {
		var options ebiten.DrawImageOptions
		options.Filter = c.config.Font.Filter
		options.GeoM.Scale(c.config.ScaleX, c.config.ScaleY)
		options.GeoM.Translate(x+float64(glyph.index)*c.config.Font.Width*c.config.ScaleX, pose.Y)
		composite.DrawRegion(dst, c.config.Font.Image, glyph.region, &options)
	}
}

func (c *CaptionCarousel) Step() {
	c.prepared = c.motion.Pose()
	c.motion.Step()
}
func (c *CaptionCarousel) Update(kit.Frame) error       { c.Step(); return nil }
func (c *CaptionCarousel) Pose() motion.CaptionPose     { return c.prepared }
func (c *CaptionCarousel) Motion() *motion.CaptionCycle { return c.motion }
