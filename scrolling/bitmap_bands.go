package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// BitmapLane places one independently moving copy of a shared text. Speed is
// pixels/second, or pixels/tick when UseTicks is enabled. Phase is pixels.
type BitmapLane struct{ X, Y, Speed, Phase float64 }

// BitmapBandsConfig renders repeated text lanes through a bounded viewport.
// Entrance is the initial left-edge position, before the first complete copy
// enters. Repeat wraps text without constructing a message-length-wide image.
// Font.Filter controls glyph rasterization; Filter controls the final image.
type BitmapBandsConfig struct {
	Font             BitmapGrid
	Text             string
	Advance          float64
	Width, Height    int
	Lanes            []BitmapLane
	Entrance         float64
	Repeat, UseTicks bool
	Filter           ebiten.Filter
	X, Y             float64
}
type BitmapBands struct {
	config  BitmapBandsConfig
	text    *BitmapText
	surface *ebiten.Image
	batch   *composite.QuadBatch
	frame   kit.Frame
	err     error
}

func NewBitmapBands(c BitmapBandsConfig) (*BitmapBands, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 8192 || c.Height > 8192 || len(c.Lanes) == 0 || len(c.Lanes) > 4096 || c.Text == "" || !finite(c.Entrance) || c.Entrance < 0 || !finite(c.X) || !finite(c.Y) {
		return nil, fmt.Errorf("scrolling: invalid bitmap text bands")
	}
	for _, lane := range c.Lanes {
		if !finite(lane.X) || !finite(lane.Y) || !finite(lane.Speed) || lane.Speed < 0 || !finite(lane.Phase) || lane.Phase < 0 {
			return nil, fmt.Errorf("scrolling: invalid text lane")
		}
	}
	text, err := NewBitmapText(c.Font, c.Text, c.Advance)
	if err != nil {
		return nil, err
	}
	maximum := math.Ceil((float64(c.Width)+c.Font.Width)/text.advance) + 2
	if !finite(text.Width()) || maximum*float64(len(c.Lanes)) > 65536 {
		return nil, fmt.Errorf("scrolling: text lanes exceed glyph budget")
	}
	c.Lanes = append([]BitmapLane(nil), c.Lanes...)
	c.Text = ""
	batch := composite.NewQuadBatch(256)
	batch.Options.Filter = c.Font.Filter
	batch.Options.Address = ebiten.AddressClampToZero
	return &BitmapBands{config: c, text: text, batch: batch, surface: ebiten.NewImageWithOptions(image.Rect(0, 0, c.Width, c.Height), &ebiten.NewImageOptions{Unmanaged: true})}, nil
}
func (b *BitmapBands) Update(frame kit.Frame) error {
	if !finite(frame.Time) || frame.Time < 0 {
		return fmt.Errorf("scrolling: invalid text-band time")
	}
	clock := frame.Time
	if b.config.UseTicks {
		if frame.Tick > 1<<52 {
			return fmt.Errorf("scrolling: text-band tick exceeds exact range")
		}
		clock = float64(frame.Tick)
	}
	for _, lane := range b.config.Lanes {
		if !finite(clock*lane.Speed + lane.Phase) {
			return fmt.Errorf("scrolling: text-band distance overflows")
		}
	}
	b.frame = frame
	return b.err
}
func (b *BitmapBands) Draw(dst *ebiten.Image) {
	if b.surface == nil || dst == nil {
		return
	}
	b.surface.Clear()
	b.err = nil
	b.batch.Begin(b.surface, b.config.Font.Image)
	clock := b.frame.Time
	if b.config.UseTicks {
		clock = float64(b.frame.Tick)
	}
	for _, lane := range b.config.Lanes {
		travel := clock*lane.Speed + lane.Phase
		x := b.config.Entrance - travel
		if b.config.Repeat && travel >= b.config.Entrance {
			x = -math.Mod(travel-b.config.Entrance, b.text.Width())
		}
		if err := b.text.appendWindow(b.batch, lane.X+x, lane.Y, float64(b.config.Width), b.config.Repeat); err != nil {
			b.err = err
			break
		}
	}
	b.batch.Flush()
	op := ebiten.DrawImageOptions{Filter: b.config.Filter}
	op.GeoM.Translate(b.config.X, b.config.Y)
	dst.DrawImage(b.surface, &op)
}
func (b *BitmapBands) Err() error { return b.err }
func (b *BitmapBands) Close() error {
	if b.surface != nil {
		b.surface.Deallocate()
		b.surface = nil
	}
	return nil
}
