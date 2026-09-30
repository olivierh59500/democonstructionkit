package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/palette"
	"github.com/olivierh59500/democonstructionkit/render"
)

// IndexedImageConfig describes a borrowed encoded image with a packed live
// palette. Format's zero value is RGB12. FPS defaults to 60. Crop selects a
// rectangle after palette conversion; Options then places/scales the result.
// Scale/Offset/Channel describe the index encoding, independently of the word
// format. A red value of index*17 uses Channel=red and Scale=15.
type IndexedImageConfig struct {
	Image         *ebiten.Image
	Palette       []uint32
	Format        palette.PackedRGB
	FPS           float64
	Channel       composite.BitplaneChannel
	Scale, Offset float32
	SourceAlpha   bool
	Crop          image.Rectangle
	Options       ebiten.DrawImageOptions
}

// IndexedImage owns palette transitions, their absolute clock, color conversion
// and placement. It retains one RGBA conversion surface, one shader and small
// CPU banks. Artwork stays borrowed and is never scanned during update/draw.
type IndexedImage struct {
	config          IndexedImageConfig
	lookup          *composite.IndexedPalette
	canvas, output  *ebiten.Image
	words, from, to []uint32
	colors          []color.NRGBA
	tick            int
	start, ticks    int
	fading, closed  bool
}

// Keep tick arithmetic portable on both 32-bit and 64-bit targets.
const maxIndexedImageTick = 1<<31 - 1

func NewIndexedImage(c IndexedImageConfig) (*IndexedImage, error) {
	if c.Image == nil || len(c.Palette) < 1 || len(c.Palette) > 256 || c.FPS < 0 || math.IsNaN(c.FPS) || math.IsInf(c.FPS, 0) {
		return nil, fmt.Errorf("effects: invalid indexed image or clock")
	}
	if c.FPS == 0 {
		c.FPS = 60
	}
	b := c.Image.Bounds()
	if b.Dx() < 1 || b.Dy() < 1 || b.Dx() > 8192 || b.Dy() > 8192 || int64(b.Dx())*int64(b.Dy()) > 16*1024*1024 {
		return nil, fmt.Errorf("effects: indexed image exceeds surface budget")
	}
	stage := image.Rect(0, 0, b.Dx(), b.Dy())
	if c.Crop.Empty() {
		c.Crop = stage
	}
	if !c.Crop.In(stage) {
		return nil, fmt.Errorf("effects: indexed output crop is outside the image")
	}
	c.Palette = append([]uint32(nil), c.Palette...)
	e := &IndexedImage{config: c, words: append([]uint32(nil), c.Palette...), colors: make([]color.NRGBA, len(c.Palette))}
	for i, word := range e.words {
		e.colors[i] = c.Format.Color(word)
	}
	var err error
	e.lookup, err = composite.NewIndexedPalette(composite.IndexedPaletteConfig{Palette: e.colors, Channel: c.Channel,
		Scale: c.Scale, Offset: c.Offset, SourceAlpha: c.SourceAlpha, Blend: ebiten.BlendCopy})
	if err != nil {
		return nil, err
	}
	e.canvas = render.NewSurface(b.Dx(), b.Dy())
	e.output = e.canvas.SubImage(c.Crop).(*ebiten.Image)
	return e, nil
}

// Fade copies its endpoints and schedules a packed-color transition. Samples
// before start show from; the last step shows to. A one-tick transition retains
// from, matching a zero-step program. Source time is expressed in configured FPS.
// Start and duration are bounded to signed 32-bit ticks on every platform.
func (e *IndexedImage) Fade(from, to []uint32, start, ticks int) error {
	if e == nil || e.closed || len(from) != len(e.words) || len(to) != len(e.words) || start < 0 || start > maxIndexedImageTick || ticks < 1 || ticks > maxIndexedImageTick {
		return fmt.Errorf("effects: invalid indexed palette fade")
	}
	e.from, e.to = append(e.from[:0], from...), append(e.to[:0], to...)
	e.start, e.ticks, e.fading = start, ticks, true
	return e.sample()
}

// SetPalette cancels a previous transition while preserving the image's pose
// and clock. Words are copied, so the caller can reuse its storage immediately.
func (e *IndexedImage) SetPalette(words []uint32) error {
	if e == nil || e.closed || len(words) != len(e.words) {
		return fmt.Errorf("effects: invalid indexed image palette")
	}
	copy(e.words, words)
	e.fading = false
	return e.sample()
}

// Update samples an absolute clock; negative time selects tick zero. Clocks
// beyond signed 32-bit ticks are rejected without changing the current palette.
func (e *IndexedImage) Update(f kit.Frame) error {
	if e == nil || e.closed || math.IsNaN(f.Time) || math.IsInf(f.Time, 0) {
		return fmt.Errorf("effects: invalid indexed image update")
	}
	tick := math.Round(max(0, f.Time) * e.config.FPS)
	if tick > maxIndexedImageTick || math.IsInf(tick, 0) {
		return fmt.Errorf("effects: indexed image clock exceeds tick range")
	}
	e.tick = int(tick)
	return e.sample()
}

func (e *IndexedImage) sample() error {
	for i := range e.words {
		if e.fading {
			e.words[i] = e.from[i]
			if e.tick >= e.start {
				amount := min(e.ticks-1, e.tick-e.start)
				word, err := e.config.Format.Blend(e.from[i], e.to[i], uint32(amount), uint32(max(1, e.ticks-1)))
				if err != nil {
					return err
				}
				e.words[i] = word
			}
		}
		e.colors[i] = e.config.Format.Color(e.words[i])
	}
	return e.lookup.SetPalette(e.colors)
}

func (e *IndexedImage) Draw(dst *ebiten.Image) {
	if e == nil || e.closed || dst == nil {
		return
	}
	if err := e.lookup.Draw(e.canvas, e.config.Image); err != nil {
		panic(err)
	}
	dst.DrawImage(e.output, &e.config.Options)
}

// Close releases conversion resources and leaves the borrowed image untouched.
func (e *IndexedImage) Close() error {
	if e == nil || e.closed {
		return nil
	}
	e.closed = true
	e.output = nil
	e.canvas.Deallocate()
	return e.lookup.Close()
}

var _ kit.Effect = (*IndexedImage)(nil)
