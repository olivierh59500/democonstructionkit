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
type ProfileImageConfig struct {
	Offsets                           []float64
	Phase, PhaseStep, RowStep         int
	Gain, BaseX, BaseY                float64
	MotionRate, MotionSpan, WrapWidth float64
	Filter                            ebiten.Filter
	Batch                             bool // Submit all sampled rows in a bounded triangle batch.
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
	for _, value := range []float64{c.Gain, c.BaseX, c.BaseY, c.MotionRate, c.MotionSpan, c.WrapWidth} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite profile image parameter")
		}
	}
	if c.WrapWidth < 0 {
		return nil, fmt.Errorf("composite: negative profile wrap width")
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
	p.phase += p.config.PhaseStep
	return nil
}
func (p *ProfileImage) Advance()           { p.phase += p.config.PhaseStep }
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
	if p.batch != nil {
		p.batch.Rect(row.Bounds(), float32(x), float32(y), float32(p.width), 1)
		return
	}
	op := ebiten.DrawImageOptions{Filter: p.config.Filter}
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
