package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type SparklePoint struct{ X, Y float64 }
type SparkleImage struct {
	Image  *ebiten.Image
	Spin   float64 // Degrees per tick; zero can instead select alternating Angles.
	Angles []float64
}

// SparkleConfig cycles arbitrary sprites over caller-selected positions.
// Images are borrowed; positions/angle lists are copied. ScaleStep can shrink or
// grow the sprite. DrawAt places this overlay on any underlying effect layer.
type SparkleConfig struct {
	Images                          []SparkleImage
	Positions                       []SparklePoint
	StartScale, EndScale, ScaleStep float64
	PauseTicks                      int
	Filter                          ebiten.Filter
}
type Sparkles struct {
	config                 SparkleConfig
	scale, rotation        float64
	image, position, pause int
	variants               []int
	visible                bool
}

func NewSparkles(c SparkleConfig) (*Sparkles, error) {
	if len(c.Images) == 0 || len(c.Positions) == 0 || c.PauseTicks < 0 || c.StartScale < 0 || c.EndScale < 0 || (c.EndScale-c.StartScale)*c.ScaleStep <= 0 {
		return nil, fmt.Errorf("sprites: invalid sparkle configuration")
	}
	for _, v := range []float64{c.StartScale, c.EndScale, c.ScaleStep} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("sprites: nonfinite sparkle scale")
		}
	}
	c.Images = append([]SparkleImage(nil), c.Images...)
	c.Positions = append([]SparklePoint(nil), c.Positions...)
	for i := range c.Images {
		if c.Images[i].Image == nil {
			return nil, fmt.Errorf("sprites: nil sparkle image")
		}
		c.Images[i].Angles = append([]float64(nil), c.Images[i].Angles...)
	}
	return &Sparkles{config: c, scale: c.StartScale, visible: true, variants: make([]int, len(c.Images))}, nil
}

// Advance follows drawing: finishing a sprite consumes the first pause tick.
// Rotation is continuous between sprites unless the next style selects angles.
func (s *Sparkles) Advance() {
	if s.visible {
		style := s.config.Images[s.image]
		if len(style.Angles) > 0 {
			s.rotation = style.Angles[s.variants[s.image]]
		} else {
			s.rotation = math.Mod(s.rotation+style.Spin, 360)
		}
		s.scale += s.config.ScaleStep
		if (s.config.ScaleStep < 0 && s.scale <= s.config.EndScale) || (s.config.ScaleStep > 0 && s.scale >= s.config.EndScale) {
			s.scale = s.config.StartScale
			s.visible = false
			s.pause = s.config.PauseTicks
		}
	}
	if !s.visible {
		s.pause--
		if s.pause <= 0 {
			s.image = (s.image + 1) % len(s.config.Images)
			s.position = (s.position + 1) % len(s.config.Positions)
			s.visible = true
			if n := len(s.config.Images[s.image].Angles); n > 0 {
				s.variants[s.image] = (s.variants[s.image] + 1) % n
			}
		}
	}
}
func (s *Sparkles) DrawAt(dst *ebiten.Image, x, y float64) {
	if !s.visible || dst == nil {
		return
	}
	img, p := s.config.Images[s.image].Image, s.config.Positions[s.position]
	op := ebiten.DrawImageOptions{Filter: s.config.Filter}
	op.GeoM.Translate(-float64(img.Bounds().Dx()/2), -float64(img.Bounds().Dy()/2))
	op.GeoM.Scale(s.scale, s.scale)
	op.GeoM.Rotate(s.rotation * math.Pi / 180)
	op.GeoM.Translate(x+p.X, y+p.Y)
	dst.DrawImage(img, &op)
}
