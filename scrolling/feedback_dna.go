package scrolling

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// FeedbackDNAConfig describes the Cuddly/Spreadpoint image-feedback scroller.
// Profile maps output rows into the wrapped buffer; it need not be a sine curve.
// Direction chooses the front (+1) or reversed back (-1) face. Dimensions,
// insertion position and speeds are explicit and independent of font metrics.
type FeedbackDNAConfig struct {
	Width, Height, HorizontalSpeed, VerticalSpeed, ColumnWidth, Direction int
	InsertY                                                               float64
	Profile                                                               []int
	Filter                                                                ebiten.Filter
}
type FeedbackDNA struct {
	config  FeedbackDNAConfig
	history *composite.Feedback
	output  *ebiten.Image
}

func NewFeedbackDNA(c FeedbackDNAConfig) (*FeedbackDNA, error) {
	if c.Width < 1 || c.Height < 1 || c.HorizontalSpeed < 1 || c.HorizontalSpeed > c.Width || c.VerticalSpeed < 0 || c.ColumnWidth < 1 || len(c.Profile) == 0 || (c.Direction != 1 && c.Direction != -1) || !finite(c.InsertY) {
		return nil, fmt.Errorf("scrolling: invalid feedback DNA configuration")
	}
	f, err := composite.NewFeedback(image.Pt(c.Width, c.Height))
	if err != nil {
		return nil, err
	}
	c.Profile = append([]int(nil), c.Profile...)
	return &FeedbackDNA{config: c, history: f, output: ebiten.NewImage(c.Width, len(c.Profile))}, nil
}

// Step advances history and inserts the newest right-hand columns from text.
// Call once per update; DrawAt can be repeated without advancing either phase.
func (d *FeedbackDNA) Step(text *ebiten.Image) {
	c := d.config
	if text == nil || d.output == nil {
		return
	}
	d.history.Shift(-c.HorizontalSpeed, -c.VerticalSpeed*c.Direction, false, true)
	b := text.Bounds()
	height := float64(b.Dy())
	for column, x := 0, c.Width-c.HorizontalSpeed; x < c.Width; column, x = column+1, x+c.ColumnWidth {
		width := min(c.ColumnWidth, c.Width-x)
		r := composite.Region{X: float64(b.Min.X + x), Y: float64(b.Min.Y), Width: float64(width), Height: height}
		for _, wrap := range []int{0, -c.Height} {
			op := ebiten.DrawImageOptions{Filter: c.Filter}
			op.GeoM.Scale(1, float64(c.Direction))
			op.GeoM.Translate(float64(x), c.InsertY+float64(column+wrap)*float64(c.Direction))
			composite.DrawRegion(d.history.Image(), text, r, &op)
		}
	}
}
func (d *FeedbackDNA) DrawAt(dst *ebiten.Image, x, y float64, phase int, gradient *ebiten.Image) {
	if d.output == nil {
		return
	}
	d.output.Clear()
	c := d.config
	for row, source := range c.Profile {
		source = ((source+phase)%c.Height + c.Height) % c.Height
		op := ebiten.DrawImageOptions{Filter: c.Filter}
		op.GeoM.Translate(0, float64(row))
		composite.DrawRegion(d.output, d.history.Image(), composite.Region{Y: float64(source), Width: float64(c.Width), Height: 1}, &op)
	}
	if gradient != nil {
		d.output.DrawImage(gradient, &ebiten.DrawImageOptions{Blend: ebiten.BlendSourceAtop, Filter: c.Filter})
	}
	op := ebiten.DrawImageOptions{Filter: c.Filter}
	op.GeoM.Translate(x, y)
	dst.DrawImage(d.output, &op)
}
func (d *FeedbackDNA) Close() error {
	if d.output != nil {
		d.output.Deallocate()
		d.output = nil
	}
	return d.history.Close()
}
