package sprites

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// MaskedProjectedFieldConfig draws the same projected population twice: first
// on an ordinary layer, then through an independently painted alpha layer.
// Paint callbacks receive cleared, reusable canvases each Draw. BaseOutput and
// MaskOutput can position, scale, tint and blend the two passes independently.
type MaskedProjectedFieldConfig struct {
	Field                  ProjectedFieldConfig
	Width, Height          int
	Unmanaged              bool
	BaseStyle, MaskStyle   FieldStyle
	PaintBase, PaintMask   func(*ebiten.Image)
	BaseOutput, MaskOutput ebiten.DrawImageOptions
}

// MaskedProjectedField owns the field transport and two bounded canvases.
type MaskedProjectedField struct {
	field  *ProjectedField
	config MaskedProjectedFieldConfig
	base   *ebiten.Image
	mask   *ebiten.Image
}

func NewMaskedProjectedField(c MaskedProjectedFieldConfig) (*MaskedProjectedField, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 {
		return nil, fmt.Errorf("sprites: invalid masked projected field size")
	}
	field, err := NewProjectedField(c.Field)
	if err != nil {
		return nil, err
	}
	newCanvas := func() *ebiten.Image {
		return ebiten.NewImageWithOptions(image.Rect(0, 0, c.Width, c.Height),
			&ebiten.NewImageOptions{Unmanaged: c.Unmanaged})
	}
	return &MaskedProjectedField{field: field, config: c, base: newCanvas(), mask: newCanvas()}, nil
}

func (m *MaskedProjectedField) Update(frame kit.Frame) error { return m.field.Update(frame) }

func (m *MaskedProjectedField) Draw(dst *ebiten.Image) {
	if m == nil || dst == nil {
		return
	}
	m.base.Clear()
	m.mask.Clear()
	if m.config.PaintBase != nil {
		m.config.PaintBase(m.base)
	}
	if m.config.PaintMask != nil {
		m.config.PaintMask(m.mask)
	}
	m.field.DrawStyle(m.base, m.config.BaseStyle)
	m.field.DrawStyle(m.mask, m.config.MaskStyle)
	composite.Instance{Image: m.base, Options: m.config.BaseOutput}.Draw(dst)
	composite.Instance{Image: m.mask, Options: m.config.MaskOutput}.Draw(dst)
}

// Field exposes the shared population for live count or camera controls.
func (m *MaskedProjectedField) Field() *ProjectedField { return m.field }

func (m *MaskedProjectedField) Close() error {
	if m == nil {
		return nil
	}
	m.base.Deallocate()
	m.mask.Deallocate()
	return m.field.Close()
}
