package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// Pixels lets original software effects write into a persistent RGBA buffer.
// Render must write premultiplied RGBA. No per-frame texture creation is needed.
type Pixels struct {
	State
	Render  func(*image.RGBA, float64)
	Rect    image.Rectangle
	Filter  ebiten.Filter
	pixels  *image.RGBA
	texture *ebiten.Image
}

func NewPixels(width, height int, drawPixels func(*image.RGBA, float64)) (*Pixels, error) {
	if width <= 0 || height <= 0 || drawPixels == nil {
		return nil, fmt.Errorf("effects: invalid pixel surface")
	}
	r := image.Rect(0, 0, width, height)
	return &Pixels{Render: drawPixels, Rect: r, pixels: image.NewRGBA(r), texture: render.NewSurface(width, height)}, nil
}
func (p *Pixels) Update(f kit.Frame) error { p.Frame = f; p.Render(p.pixels, f.Time); return nil }
func (p *Pixels) Draw(dst *ebiten.Image) {
	p.texture.WritePixels(p.pixels.Pix)
	op := ebiten.DrawImageOptions{Filter: p.Filter}
	op.GeoM.Scale(float64(p.Rect.Dx())/float64(p.pixels.Rect.Dx()), float64(p.Rect.Dy())/float64(p.pixels.Rect.Dy()))
	op.GeoM.Translate(float64(p.Rect.Min.X), float64(p.Rect.Min.Y))
	dst.DrawImage(p.texture, &op)
}
func (p *Pixels) Close() error { p.texture.Deallocate(); return nil }

// Plasma is a four-harmonic indexed plasma with a configurable cycling palette.
func Plasma(palette []color.NRGBA, spatial, speed float64) func(*image.RGBA, float64) {
	colors := append([]color.NRGBA(nil), palette...)
	return func(dst *image.RGBA, t float64) {
		if len(colors) == 0 {
			clear(dst.Pix)
			return
		}
		w, h := dst.Rect.Dx(), dst.Rect.Dy()
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				fx, fy := float64(x)*spatial, float64(y)*spatial
				v := math.Sin(fx+t*speed) + math.Sin(fy+t*speed*.7) + math.Sin((fx+fy)*.7-t*speed*.8) + math.Sin(math.Hypot(fx, fy)-t*speed)
				i := int(motion.Wrap((v+4)*float64(len(colors))/8+t*speed*12, float64(len(colors))))
				dst.Set(x, y, colors[i])
			}
		}
	}
}

// SampleMap maps destination coordinates to source pixels. Wrap is optional.
func SampleMap(source image.Image, wrap bool, mapping func(x, y, t float64) (float64, float64)) func(*image.RGBA, float64) {
	return func(dst *image.RGBA, t float64) {
		if source == nil || mapping == nil {
			clear(dst.Pix)
			return
		}
		b := source.Bounds()
		for y := 0; y < dst.Rect.Dy(); y++ {
			for x := 0; x < dst.Rect.Dx(); x++ {
				u, v := mapping(float64(x), float64(y), t)
				if wrap {
					u = motion.Wrap(u, float64(b.Dx()))
					v = motion.Wrap(v, float64(b.Dy()))
				}
				sample := image.Pt(int(math.Floor(u))+b.Min.X, int(math.Floor(v))+b.Min.Y)
				if sample.In(b) {
					dst.Set(x, y, source.At(sample.X, sample.Y))
				} else {
					dst.SetRGBA(x, y, color.RGBA{})
				}
			}
		}
	}
}

// Tunnel uses polar angle and inverse radius to wrap a texture into a tunnel.
func Tunnel(source image.Image, width, height int, speed, rotation float64) func(*image.RGBA, float64) {
	return SampleMap(source, true, func(x, y, t float64) (float64, float64) {
		dx, dy := x-float64(width)/2, y-float64(height)/2
		r := math.Max(1, math.Hypot(dx, dy))
		return (math.Atan2(dy, dx)/(2*math.Pi) + t*rotation) * float64(source.Bounds().Dx()), float64(height)*24/r + t*speed
	})
}

// Ripple displaces a texture with radial water waves around a moving center.
func Ripple(source image.Image, amplitude, frequency, speed float64) func(*image.RGBA, float64) {
	return SampleMap(source, true, func(x, y, t float64) (float64, float64) {
		b := source.Bounds()
		dx, dy := x-float64(b.Dx())/2, y-float64(b.Dy())/2
		r := math.Hypot(dx, dy)
		if r < 1 {
			return x, y
		}
		wave := amplitude * math.Sin(r*frequency-t*speed)
		return x + dx/r*wave, y + dy/r*wave
	})
}

// Lens magnifies a circular region; the mapping is continuous at its boundary.
func Lens(source image.Image, centerX, centerY, radius, strength float64) func(*image.RGBA, float64) {
	return SampleMap(source, false, func(x, y, t float64) (float64, float64) {
		dx, dy := x-centerX, y-centerY
		r := math.Hypot(dx, dy)
		if radius <= 0 || r >= radius {
			return x, y
		}
		scale := 1 + strength*(1-r/radius)*(1-r/radius)
		if scale <= .01 {
			scale = .01
		}
		return centerX + dx/scale, centerY + dy/scale
	})
}

// Indexed expands an indexed frame with a palette offset, preserving transparent
// palette entries. Applications can fill Indices with original VGA/planar decoders.
type Indexed struct {
	Width, Height  int
	Indices        []byte
	Palette        []color.NRGBA
	CyclePerSecond float64
}

func (s *Indexed) Render(dst *image.RGBA, t float64) {
	if len(s.Palette) == 0 {
		clear(dst.Pix)
		return
	}
	offset := int(math.Floor(t * s.CyclePerSecond))
	for y := 0; y < dst.Rect.Dy(); y++ {
		for x := 0; x < dst.Rect.Dx(); x++ {
			i := y*s.Width + x
			if x >= s.Width || y >= s.Height || i >= len(s.Indices) {
				dst.SetRGBA(x, y, color.RGBA{})
				continue
			}
			index := int(motion.Wrap(float64(int(s.Indices[i])+offset), float64(len(s.Palette))))
			dst.Set(x, y, s.Palette[index])
		}
	}
}
