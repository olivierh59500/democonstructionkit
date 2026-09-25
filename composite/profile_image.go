package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// ProfileImageConfig samples an arbitrary horizontal displacement table per
// source row. PhaseStep advances once per Update; RowStep spaces lookup indices.
// MotionSpan is the peak-to-peak center movement, with a sine amplitude of half
// that span. A positive WrapWidth draws one opposite-side copy when needed.
// PhaseWrap resets the phase after a strict upper boundary. ScaleX/Y and
// OutputX/Y map native row coordinates into a larger parent viewport.
type ProfileImageConfig struct {
	Offsets                              []float64
	Phase, PhaseStep, RowStep, PhaseWrap int
	Gain, BaseX, BaseY                   float64
	MotionRate, MotionSpan, WrapWidth    float64
	ScaleX, ScaleY, OutputX, OutputY     float64
	Filter                               ebiten.Filter
	Batch                                bool // Submit all sampled rows in a bounded triangle batch.
}

// ProfileImage owns no GPU image. It caches borrowed source rows once, keeps an
// independent phase and preserves source row order when copies overlap.
type ProfileImage struct {
	config ProfileImageConfig
	rows   []*ebiten.Image
	source *ebiten.Image
	batch  *QuadBatch
	width  float64
	phase  int
}

func NewProfileImage(source *ebiten.Image, c ProfileImageConfig) (*ProfileImage, error) {
	if source == nil || source.Bounds().Empty() || len(c.Offsets) == 0 || len(c.Offsets) > 1<<20 {
		return nil, fmt.Errorf("composite: invalid profile image dimensions or phase")
	}
	for _, value := range []float64{c.Gain, c.BaseX, c.BaseY, c.MotionRate, c.MotionSpan, c.WrapWidth, c.ScaleX, c.ScaleY, c.OutputX, c.OutputY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite profile image parameter")
		}
	}
	if c.WrapWidth < 0 || c.PhaseWrap < 0 || c.ScaleX < 0 || c.ScaleY < 0 {
		return nil, fmt.Errorf("composite: invalid profile wrap or scale")
	}
	if c.ScaleX == 0 {
		c.ScaleX = 1
	}
	if c.ScaleY == 0 {
		c.ScaleY = 1
	}
	for _, value := range c.Offsets {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite profile offset")
		}
	}
	c.Offsets = append([]float64(nil), c.Offsets...)
	b := source.Bounds()
	p := &ProfileImage{config: c, rows: make([]*ebiten.Image, b.Dy()), source: source, width: float64(b.Dx()), phase: c.Phase}
	if c.Batch {
		p.batch = NewQuadBatch(min(16383, b.Dy()*2))
		p.batch.Options.Filter = c.Filter
	}
	for row := range p.rows {
		r := b
		r.Min.Y = b.Min.Y + row
		r.Max.Y = r.Min.Y + 1
		p.rows[row] = source.SubImage(r).(*ebiten.Image)
	}
	return p, nil
}

// Update advances the authored phase without drawing or allocating.
func (p *ProfileImage) Update(kit.Frame) error {
	p.Advance()
	return nil
}
func (p *ProfileImage) Advance() {
	p.phase += p.config.PhaseStep
	if p.config.PhaseWrap > 0 && p.phase > p.config.PhaseWrap {
		p.phase = 0
	}
}
func (p *ProfileImage) Phase() int         { return p.phase }
func (p *ProfileImage) SetPhase(phase int) { p.phase = phase }

func (p *ProfileImage) Draw(dst *ebiten.Image) {
	if p != nil {
		p.DrawAt(dst, p.config.BaseX, p.config.BaseY)
	}
}
func (p *ProfileImage) DrawAt(dst *ebiten.Image, baseX, baseY float64) {
	if p == nil || dst == nil || len(p.rows) == 0 {
		return
	}
	c := p.config
	movement := math.Sin(float64(p.phase)*c.MotionRate) * c.MotionSpan / 2
	if p.batch != nil {
		p.batch.Begin(dst, p.source)
	}
	for row, img := range p.rows {
		index := (p.phase + row*c.RowStep) % len(c.Offsets)
		if index < 0 {
			index += len(c.Offsets)
		}
		x := baseX + movement + c.Offsets[index]*c.Gain - p.width/2
		y := baseY + float64(row)
		if x > -p.width && (c.WrapWidth == 0 || x < c.WrapWidth) {
			p.drawRow(dst, img, x, y)
		}
		if c.WrapWidth > 0 {
			if x < 0 {
				p.drawRow(dst, img, c.WrapWidth+x, y)
			} else if x+p.width > c.WrapWidth {
				p.drawRow(dst, img, x-c.WrapWidth, y)
			}
		}
	}
	if p.batch != nil {
		p.batch.Flush()
	}
}

func (p *ProfileImage) drawRow(dst, row *ebiten.Image, x, y float64) {
	c := p.config
	width, height := p.width, 1.0
	if c.ScaleX != 1 || c.ScaleY != 1 || c.OutputX != 0 || c.OutputY != 0 {
		x = c.OutputX + x*c.ScaleX
		y = c.OutputY + y*c.ScaleY
		width *= c.ScaleX
		height *= c.ScaleY
	}
	if p.batch != nil {
		p.batch.Rect(row.Bounds(), float32(x), float32(y), float32(width), float32(height))
		return
	}
	op := ebiten.DrawImageOptions{Filter: c.Filter}
	if c.ScaleX != 1 || c.ScaleY != 1 {
		op.GeoM.Scale(c.ScaleX, c.ScaleY)
	}
	op.GeoM.Translate(x, y)
	dst.DrawImage(row, &op)
}

func (p *ProfileImage) Close() error {
	if p != nil {
		p.rows = nil
		p.source = nil
		p.batch = nil
	}
	return nil
}
