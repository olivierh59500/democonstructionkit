package composite

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// TableWarpLogoConfig borrows one image and binds a compiled row curve to an
// independent bounce clock. The same prepared bounce can position other art.
type TableWarpLogoConfig struct {
	Image          *ebiten.Image
	Motion         motion.TableWarpRowsConfig
	SourceX, Width float64
	Filter         ebiten.Filter
	Blend          ebiten.Blend
}

// TableWarpLogo reuses SampledRows without another GPU surface. Update samples
// the current pose before moving its clocks, matching draw-before-step sources.
type TableWarpLogo struct {
	motion         *motion.TableWarpRows
	rows           *SampledRows
	preparedBounce float64
}

func NewTableWarpLogo(c TableWarpLogoConfig) (*TableWarpLogo, error) {
	if c.Image == nil || c.Width <= 0 {
		return nil, fmt.Errorf("composite: invalid table-warped logo image or crop")
	}
	program, err := motion.NewTableWarpRows(c.Motion)
	if err != nil {
		return nil, err
	}
	rows, err := NewSampledRows(SampledRowsConfig{
		Image: c.Image, Program: program, Rows: c.Motion.Rows, Copies: 1,
		SourceX: c.SourceX, SourceWidth: c.Width,
		Filter: c.Filter, Blend: c.Blend,
	})
	if err != nil {
		return nil, err
	}
	return &TableWarpLogo{motion: program, rows: rows,
		preparedBounce: program.Bounce()}, nil
}

func (l *TableWarpLogo) Update(frame kit.Frame) error {
	l.preparedBounce = l.motion.Bounce()
	if err := l.rows.Update(frame); err != nil {
		return err
	}
	l.motion.Advance()
	return nil
}

func (l *TableWarpLogo) Draw(dst *ebiten.Image)        { l.rows.Draw(dst) }
func (l *TableWarpLogo) Bounce() float64               { return l.preparedBounce }
func (l *TableWarpLogo) Motion() *motion.TableWarpRows { return l.motion }
func (l *TableWarpLogo) Rows() *SampledRows            { return l.rows }
