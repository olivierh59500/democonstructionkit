package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// ProjectedFieldConfig composes a depth field and its painter into one effect.
// Delta and Velocity control one logical Update; View and Style are independent
// so the same particles can be redrawn as pixels, sprites or trails on layers.
type ProjectedFieldConfig struct {
	Field            FieldConfig
	View             FieldView
	Velocity         geometry.Vec3
	Delta            float64
	ViewVelocity     geometry.Vec3 // Camera offset change per Update; zero leaves the authored view unchanged.
	AngleStep        float64       // Radians per Update; zero leaves the authored angle unchanged.
	Style            FieldStyle
	RendererCapacity int
	// ColdStart suppresses history on the first Update, for trails whose
	// first step has no previous projected endpoint.
	ColdStart bool
}

// ProjectedField owns the particle transport and bounded draw storage. It
// samples initial positions before the first Update, preserving frame zero.
type ProjectedField struct {
	field    *Field
	renderer *FieldRenderer
	config   ProjectedFieldConfig
}

func NewProjectedField(c ProjectedFieldConfig) (*ProjectedField, error) {
	if c.RendererCapacity < 0 || c.View.Camera.Focal <= 0 || c.View.Camera.Near <= 0 {
		return nil, fmt.Errorf("sprites: invalid projected field step or capacity")
	}
	for _, v := range []float64{c.Delta, c.Velocity.X, c.Velocity.Y, c.Velocity.Z,
		c.ViewVelocity.X, c.ViewVelocity.Y, c.ViewVelocity.Z, c.AngleStep,
		c.View.Camera.Focal, c.View.Camera.Near, c.View.Camera.Center.X, c.View.Camera.Center.Y,
		c.View.Offset.X, c.View.Offset.Y, c.View.Offset.Z, c.View.Angle} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("sprites: non-finite projected field parameter")
		}
	}
	f, err := NewField(c.Field)
	if err != nil {
		return nil, err
	}
	capacity := c.RendererCapacity
	if capacity == 0 {
		capacity = max(1, len(f.points))
	}
	c.Field.Points = nil
	p := &ProjectedField{field: f, renderer: NewFieldRenderer(capacity), config: c}
	p.field.Sample(c.View)
	if c.ColdStart {
		clear(p.field.history)
	}
	return p, nil
}

func (p *ProjectedField) Update(kit.Frame) error {
	if p.config.ViewVelocity != (geometry.Vec3{}) {
		p.config.View.Offset = p.config.View.Offset.Add(p.config.ViewVelocity)
	}
	if p.config.AngleStep != 0 {
		p.config.View.Angle += p.config.AngleStep
	}
	p.field.Step(p.config.Delta, p.config.Velocity)
	p.field.Sample(p.config.View)
	return nil
}

func (p *ProjectedField) Draw(dst *ebiten.Image) {
	p.renderer.Draw(dst, p.field.Samples(), p.config.Style)
}

// DrawStyle reuses the same sampled positions with a different material.
func (p *ProjectedField) DrawStyle(dst *ebiten.Image, style FieldStyle) {
	p.renderer.Draw(dst, p.field.Samples(), style)
}

// SetView changes the camera for the next update without resetting particles.
func (p *ProjectedField) SetView(view FieldView) { p.config.View = view }

// View returns the prepared camera pose without exposing internal particle data.
func (p *ProjectedField) View() FieldView { return p.config.View }

// ResetCount reinitializes a field that supplies a spawn function.
func (p *ProjectedField) ResetCount(count int) error { return p.field.ResetCount(count) }

// Samples borrows the current projection until the next Update or ResetCount.
func (p *ProjectedField) Samples() []FieldSample { return p.field.Samples() }

func (p *ProjectedField) Close() error { return p.renderer.Close() }
