package sprites

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

// ImageSlot selects one borrowed image crop and its independent destination
// rectangle. Empty Source selects the image; zero Width/Height select crop size.
type ImageSlot struct {
	Image               int
	Source              image.Rectangle
	X, Y, Width, Height float64
	Tint                ebiten.ColorScale
	Hidden              bool
}
type ImageSlotsConfig struct {
	Images   []*ebiten.Image
	Slots    []ImageSlot
	MaxSlots int
	Select   func(kit.Frame, []ImageSlot) (int, error)
	Blend    ebiten.Blend
	Filter   ebiten.Filter
}

// ImageSlots owns a copied selection window and one reusable triangle batch.
// Drawing preserves slot order and texture changes; no image surface is created.
type ImageSlots struct {
	config         ImageSlotsConfig
	images         []*ebiten.Image
	slots, scratch []ImageSlot
	count          int
	batch          *render.Batch
	closed         bool
}

func slotFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) < 1<<24 }
func NewImageSlots(c ImageSlotsConfig) (*ImageSlots, error) {
	if c.MaxSlots == 0 {
		c.MaxSlots = 1024
	}
	if len(c.Images) < 1 || len(c.Images) > 65536 || c.MaxSlots < 1 || c.MaxSlots > 65536 || len(c.Slots) > c.MaxSlots {
		return nil, fmt.Errorf("sprites: invalid image slot budget")
	}
	for _, img := range c.Images {
		if img == nil {
			return nil, fmt.Errorf("sprites: absent image slot artwork")
		}
	}
	s := &ImageSlots{config: c, images: append([]*ebiten.Image(nil), c.Images...), slots: make([]ImageSlot, c.MaxSlots), scratch: make([]ImageSlot, c.MaxSlots), batch: render.NewBatch(min(20000, c.MaxSlots*2))}
	s.config.Images = nil
	s.config.Slots = nil
	s.batch.Options.Blend = c.Blend
	s.batch.Options.Filter = c.Filter
	if err := s.SetSlots(c.Slots); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *ImageSlots) validate(slots []ImageSlot) error {
	if len(slots) > len(s.slots) {
		return fmt.Errorf("sprites: image slots exceed budget")
	}
	for _, slot := range slots {
		if slot.Hidden {
			continue
		}
		if slot.Image < 0 || slot.Image >= len(s.images) {
			return fmt.Errorf("sprites: image slot index outside bank")
		}
		rect := slot.Source
		if rect.Empty() {
			rect = s.images[slot.Image].Bounds()
		}
		if !rect.In(s.images[slot.Image].Bounds()) || !slotFinite(slot.X) || !slotFinite(slot.Y) || !slotFinite(slot.Width) || !slotFinite(slot.Height) || slot.Width < 0 || slot.Height < 0 {
			return fmt.Errorf("sprites: invalid image slot crop or pose")
		}
	}
	return nil
}
func (s *ImageSlots) SetSlots(slots []ImageSlot) error {
	if s == nil || s.closed {
		return fmt.Errorf("sprites: closed image slot bank")
	}
	if err := s.validate(slots); err != nil {
		return err
	}
	copy(s.slots, slots)
	s.count = len(slots)
	return nil
}
func (s *ImageSlots) Update(f kit.Frame) error {
	if s == nil || s.closed {
		return fmt.Errorf("sprites: closed image slot update")
	}
	if s.config.Select != nil {
		copy(s.scratch, s.slots)
		count, err := s.config.Select(f, s.scratch)
		if err != nil {
			return err
		}
		if count < 0 || count > len(s.scratch) {
			return fmt.Errorf("sprites: invalid selected slot count")
		}
		return s.SetSlots(s.scratch[:count])
	}
	return nil
}
func (s *ImageSlots) Draw(dst *ebiten.Image) {
	if s == nil || s.closed || dst == nil {
		return
	}
	for _, slot := range s.slots[:s.count] {
		if slot.Hidden {
			continue
		}
		img := s.images[slot.Image]
		rect := slot.Source
		if rect.Empty() {
			rect = img.Bounds()
		}
		w, h := slot.Width, slot.Height
		if w == 0 {
			w = float64(rect.Dx())
		}
		if h == 0 {
			h = float64(rect.Dy())
		}
		v := ebiten.Vertex{DstX: float32(slot.X), DstY: float32(slot.Y), SrcX: float32(rect.Min.X), SrcY: float32(rect.Min.Y), ColorR: slot.Tint.R(), ColorG: slot.Tint.G(), ColorB: slot.Tint.B(), ColorA: slot.Tint.A()}
		q := [4]ebiten.Vertex{v, v, v, v}
		q[1].DstX, q[1].SrcX = float32(slot.X+w), float32(rect.Max.X)
		q[2].DstX, q[2].DstY, q[2].SrcX, q[2].SrcY = float32(slot.X+w), float32(slot.Y+h), float32(rect.Max.X), float32(rect.Max.Y)
		q[3].DstY, q[3].SrcY = float32(slot.Y+h), float32(rect.Max.Y)
		s.batch.Begin(dst, img)
		s.batch.Quad(q)
		s.batch.Flush()
	}
}
func (s *ImageSlots) Close() error {
	if s != nil {
		s.closed = true
		s.images = nil
		s.slots = nil
		s.scratch = nil
		s.batch = nil
	}
	return nil
}

var _ kit.Effect = (*ImageSlots)(nil)
