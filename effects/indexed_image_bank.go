package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// IndexedImagePalette associates copied colors with an optional lookup name.
// Every palette in a bank has the same 1..256 entry count. Names, when supplied,
// are unique; unnamed palettes remain accessible through their slice index.
type IndexedImagePalette struct {
	Name   string
	Colors []color.NRGBA
}

// IndexedImageSlot selects borrowed artwork and a palette, then places its
// selected crop. Crop uses coordinates relative to the artwork's Bounds().Min;
// an empty crop selects the complete image. Hidden skips the slot independently
// of its indices. Options controls placement, tint and blending; Images and
// Uniforms are reserved for the bank's palette shader and ignored.
type IndexedImageSlot struct {
	Image, Palette int
	Crop           image.Rectangle
	Options        ebiten.DrawRectShaderOptions
	Hidden         bool
}

// IndexedImageBankConfig separates artwork, colors and retained output slots.
// Select optionally writes a fixed slot window once per Update; it receives
// reusable storage containing the current selections. It must not retain that
// slice. SetSlots replaces and copies the window explicitly. Neither path scans
// artwork, and Draw never advances the selector's clock. MaxSlots defaults to
// 1024 and cannot exceed 65536.
type IndexedImageBankConfig struct {
	Images        []*ebiten.Image
	Palettes      []IndexedImagePalette
	Slots         []IndexedImageSlot
	Channel       composite.BitplaneChannel
	Scale, Offset float32
	SourceAlpha   bool
	MaxSlots      int
	Select        func(kit.Frame, []IndexedImageSlot) error
}

// IndexedImageBank owns one palette shader and small selection/color banks.
// Each visible slot submits one direct shader draw; there are no intermediate
// conversion surfaces or precolored copies. Images and their cropped views stay
// borrowed. Ordinary Update, Draw and unchanged-window SetSlots reuse storage.
type IndexedImageBank struct {
	images   []*ebiten.Image
	palettes []IndexedImagePalette
	names    map[string]int
	slots    []IndexedImageSlot
	sources  []*ebiten.Image
	scratch  []IndexedImageSlot
	lookup   *composite.IndexedPalette
	selectAt func(kit.Frame, []IndexedImageSlot) error
	maxSlots int
	active   int
	err      error
	closed   bool
}

func NewIndexedImageBank(c IndexedImageBankConfig) (*IndexedImageBank, error) {
	if len(c.Images) < 1 || len(c.Images) > 65536 || len(c.Palettes) < 1 || len(c.Palettes) > 256 {
		return nil, fmt.Errorf("effects: invalid indexed artwork or palette count")
	}
	if c.MaxSlots == 0 {
		c.MaxSlots = 1024
	}
	if c.MaxSlots < 1 || c.MaxSlots > 65536 || len(c.Slots) > c.MaxSlots {
		return nil, fmt.Errorf("effects: invalid indexed slot budget")
	}
	for _, source := range c.Images {
		if source == nil {
			return nil, fmt.Errorf("effects: missing indexed artwork")
		}
		bounds := source.Bounds()
		if bounds.Dx() < 1 || bounds.Dy() < 1 || bounds.Dx() > 8192 || bounds.Dy() > 8192 || int64(bounds.Dx())*int64(bounds.Dy()) > 16*1024*1024 {
			return nil, fmt.Errorf("effects: indexed artwork exceeds its dimension budget")
		}
	}
	count := len(c.Palettes[0].Colors)
	if count < 1 || count > 256 {
		return nil, fmt.Errorf("effects: invalid indexed palette entry count")
	}
	bank := &IndexedImageBank{images: append([]*ebiten.Image(nil), c.Images...),
		palettes: make([]IndexedImagePalette, len(c.Palettes)), names: make(map[string]int),
		selectAt: c.Select, maxSlots: c.MaxSlots, active: -1}
	for i, palette := range c.Palettes {
		if len(palette.Colors) != count || len(palette.Name) > 1024 {
			return nil, fmt.Errorf("effects: inconsistent indexed palette %d", i)
		}
		if palette.Name != "" {
			if _, exists := bank.names[palette.Name]; exists {
				return nil, fmt.Errorf("effects: duplicate indexed palette %q", palette.Name)
			}
			bank.names[palette.Name] = i
		}
		bank.palettes[i] = IndexedImagePalette{Name: palette.Name, Colors: append([]color.NRGBA(nil), palette.Colors...)}
	}
	var err error
	bank.lookup, err = composite.NewIndexedPalette(composite.IndexedPaletteConfig{Palette: bank.palettes[0].Colors,
		Channel: c.Channel, Scale: c.Scale, Offset: c.Offset, SourceAlpha: c.SourceAlpha})
	if err != nil {
		return nil, err
	}
	if err := bank.SetSlots(c.Slots); err != nil {
		bank.Close()
		return nil, err
	}
	return bank, nil
}

// PaletteIndex resolves a configured name without allocating. Slice indices
// remain stable for the lifetime of the bank, including after color changes.
func (b *IndexedImageBank) PaletteIndex(name string) (int, bool) {
	if b == nil || b.closed {
		return 0, false
	}
	index, exists := b.names[name]
	return index, exists
}

// SetPalette copies live colors while preserving its name and current slots.
func (b *IndexedImageBank) SetPalette(index int, colors []color.NRGBA) error {
	if b == nil || b.closed || index < 0 || index >= len(b.palettes) || len(colors) != len(b.palettes[index].Colors) {
		return fmt.Errorf("effects: invalid indexed bank palette update")
	}
	copy(b.palettes[index].Colors, colors)
	if index == b.active {
		return b.lookup.SetPalette(b.palettes[index].Colors)
	}
	return nil
}

// SetSlots validates the entire new window before changing a visible selection.
// Crops create only borrowed subimage views, reused when the source/crop remains
// unchanged. Caller slot, transform and color-scale storage is copied.
func (b *IndexedImageBank) SetSlots(slots []IndexedImageSlot) error {
	if b == nil || b.closed || len(slots) > b.maxSlots {
		return fmt.Errorf("effects: invalid indexed slot window")
	}
	for _, slot := range slots {
		if err := b.validateSlot(slot); err != nil {
			return err
		}
	}
	previousSlots, previousSources := b.slots, b.sources
	if cap(b.slots) < len(slots) {
		b.slots = make([]IndexedImageSlot, len(slots))
		b.sources = make([]*ebiten.Image, len(slots))
	} else {
		b.slots, b.sources = b.slots[:len(slots)], b.sources[:len(slots)]
	}
	for i, slot := range slots {
		var source *ebiten.Image
		if !slot.Hidden {
			if i < len(previousSlots) && !previousSlots[i].Hidden && slot.Image == previousSlots[i].Image && slot.Crop == previousSlots[i].Crop {
				source = previousSources[i]
			} else {
				source = b.images[slot.Image]
				if !slot.Crop.Empty() {
					source = source.SubImage(slot.Crop.Add(source.Bounds().Min)).(*ebiten.Image)
				}
			}
		}
		slot.Options.Images, slot.Options.Uniforms = [4]*ebiten.Image{}, nil
		b.slots[i], b.sources[i] = slot, source
	}
	if cap(b.scratch) < len(slots) {
		b.scratch = make([]IndexedImageSlot, len(slots))
	} else {
		b.scratch = b.scratch[:len(slots)]
	}
	return nil
}

func (b *IndexedImageBank) validateSlot(slot IndexedImageSlot) error {
	if slot.Hidden {
		return nil
	}
	if slot.Image < 0 || slot.Image >= len(b.images) || slot.Palette < 0 || slot.Palette >= len(b.palettes) {
		return fmt.Errorf("effects: indexed slot selects unavailable artwork or palette")
	}
	bounds := b.images[slot.Image].Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if !slot.Crop.Empty() {
		if !slot.Crop.In(image.Rect(0, 0, width, height)) {
			return fmt.Errorf("effects: indexed slot crop is outside its artwork")
		}
		width, height = slot.Crop.Dx(), slot.Crop.Dy()
	}
	for _, point := range [4][2]float64{{0, 0}, {float64(width), 0}, {0, float64(height)}, {float64(width), float64(height)}} {
		x, y := slot.Options.GeoM.Apply(point[0], point[1])
		if !indexedBankCoordinate(x) || !indexedBankCoordinate(y) {
			return fmt.Errorf("effects: indexed slot has an invalid transform")
		}
	}
	for _, value := range [...]float32{slot.Options.ColorScale.R(), slot.Options.ColorScale.G(), slot.Options.ColorScale.B(), slot.Options.ColorScale.A()} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return fmt.Errorf("effects: indexed slot has an invalid tint")
		}
	}
	return nil
}

func indexedBankCoordinate(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && math.Abs(value) <= math.MaxFloat32
}

func (b *IndexedImageBank) Update(frame kit.Frame) error {
	if b == nil || b.closed || math.IsNaN(frame.Time) || math.IsInf(frame.Time, 0) || math.IsNaN(frame.Delta) || math.IsInf(frame.Delta, 0) {
		return fmt.Errorf("effects: invalid indexed bank update")
	}
	if b.selectAt == nil {
		return nil
	}
	copy(b.scratch, b.slots)
	if err := b.selectAt(frame, b.scratch); err != nil {
		return err
	}
	return b.SetSlots(b.scratch)
}

func (b *IndexedImageBank) Draw(dst *ebiten.Image) {
	if b == nil || b.closed || dst == nil {
		return
	}
	b.err = nil
	for i, slot := range b.slots {
		if !slot.Hidden && (dst == b.sources[i] || dst == b.images[slot.Image]) {
			b.err = fmt.Errorf("effects: indexed bank destination aliases its artwork")
			return
		}
	}
	for i, slot := range b.slots {
		if slot.Hidden {
			continue
		}
		if b.active != slot.Palette {
			if b.err = b.lookup.SetPalette(b.palettes[slot.Palette].Colors); b.err != nil {
				return
			}
			b.active = slot.Palette
		}
		if b.err = b.lookup.DrawWith(dst, b.sources[i], slot.Options); b.err != nil {
			return
		}
	}
}

// Err reports the most recent draw error; selector and slot errors return from
// Update or SetSlots directly without replacing the last valid selection.
func (b *IndexedImageBank) Err() error {
	if b == nil {
		return nil
	}
	return b.err
}

// Close releases the shader and owned CPU storage, leaving all caller images
// alive. Repeated Close and Draw after Close are harmless.
func (b *IndexedImageBank) Close() error {
	if b == nil || b.closed {
		return nil
	}
	b.closed = true
	b.images, b.sources = nil, nil
	b.palettes, b.slots, b.scratch = nil, nil, nil
	b.names, b.selectAt = nil, nil
	return b.lookup.Close()
}

var _ kit.Effect = (*IndexedImageBank)(nil)
