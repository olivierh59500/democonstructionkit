package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CrawlConfig combines cached paragraphs, vertical transport and a scanline
// projection. Source size remains bounded by the viewport, never by text length.
// A zero PixelsPerUpdate leaves animation entirely under SetPosition control.
type CrawlConfig struct {
	Paragraph                                        BitmapParagraphConfig
	Width, Height, ProjectionWidth, ProjectionHeight int
	VisibleLines                                     int
	PixelsPerUpdate                                  float64
	Projection                                       *composite.RowProjection
	Output                                           composite.Region
	X, Y                                             float64
	Filter                                           ebiten.Filter
}
type Crawl struct {
	config            CrawlConfig
	paragraph         *BitmapParagraph
	source, projected *ebiten.Image
	line              int
	offset            float64
}

func NewCrawl(c CrawlConfig) (*Crawl, error) {
	if c.Width <= 0 || c.Height <= 0 || c.ProjectionWidth <= 0 || c.ProjectionHeight <= 0 || c.VisibleLines <= 0 || !finite(c.PixelsPerUpdate) || c.PixelsPerUpdate < 0 || c.Projection == nil || !c.Output.Valid() || !finite(c.X) || !finite(c.Y) {
		return nil, fmt.Errorf("scrolling: invalid perspective crawl")
	}
	p, err := NewBitmapParagraph(c.Paragraph)
	if err != nil {
		return nil, err
	}
	if p.Len() < c.VisibleLines {
		return nil, fmt.Errorf("scrolling: paragraph shorter than visible window")
	}
	c.Paragraph.Lines = nil
	return &Crawl{config: c, paragraph: p, source: ebiten.NewImageWithOptions(image.Rect(0, 0, c.Width, c.Height), &ebiten.NewImageOptions{Unmanaged: true}), projected: ebiten.NewImageWithOptions(image.Rect(0, 0, c.ProjectionWidth, c.ProjectionHeight), &ebiten.NewImageOptions{Unmanaged: true})}, nil
}
func (c *Crawl) Update(kit.Frame) error { c.Step(); return nil }
func (c *Crawl) Step() {
	c.offset -= c.config.PixelsPerUpdate
	advance := c.config.Paragraph.LineAdvance
	if c.offset <= -advance {
		lines := math.Floor(-c.offset / advance)
		c.offset = math.Mod(c.offset, advance)
		count := c.paragraph.Len() - c.config.VisibleLines + 1
		c.line = (c.line + int(math.Mod(lines, float64(count)))) % count
	}
}

// SetPosition seeks to a complete paragraph window and an upward pixel offset.
func (c *Crawl) SetPosition(line int, offset float64) error {
	if line < 0 || line > c.paragraph.Len()-c.config.VisibleLines || !finite(offset) || offset > 0 || offset <= -c.config.Paragraph.LineAdvance {
		return fmt.Errorf("scrolling: invalid crawl position")
	}
	c.line, c.offset = line, offset
	return nil
}
func (c *Crawl) Draw(dst *ebiten.Image) {
	if c == nil || c.source == nil || dst == nil {
		return
	}
	c.source.Clear()
	c.projected.Clear()
	c.paragraph.DrawWindow(c.source, c.line, c.config.VisibleLines, 0, c.offset)
	c.config.Projection.Draw(c.projected, c.source)
	op := ebiten.DrawImageOptions{Filter: c.config.Filter}
	op.GeoM.Translate(c.config.X, c.config.Y)
	composite.DrawRegion(dst, c.projected, c.config.Output, &op)
}
func (c *Crawl) Close() error {
	if c.source != nil {
		c.source.Deallocate()
		c.projected.Deallocate()
		c.source = nil
		c.projected = nil
	}
	return nil
}

// SetSpeed changes upward pixels per Update without changing the current line.
// Use zero to pause. Call it from the same thread that updates this effect.
func (c *Crawl) SetSpeed(pixelsPerUpdate float64) error {
	if !finite(pixelsPerUpdate) || pixelsPerUpdate < 0 {
		return fmt.Errorf("scrolling: invalid crawl speed")
	}
	c.config.PixelsPerUpdate = pixelsPerUpdate
	return nil
}
