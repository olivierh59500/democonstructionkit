// Package scene builds a serializable demo entirely from procedural Go assets.
// Desktop and mobile hosts share this effect and can supply their own clocks.
package scene

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/authoring"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type Scene struct {
	*authoring.Compiled
	Project       authoring.Project
	Width, Height int
	images        []*ebiten.Image
}

// NewSize selects a real logical resolution for profiling, including font and
// sprite sizes. It does not render a full-size hidden surface in economy mode.
func NewSize(width, height int) (*Scene, error) {
	if width < 1 || height < 1 {
		return nil, fmt.Errorf("authoring scene: invalid resolution")
	}
	p := DefaultProject()
	p.Canvas.Width, p.Canvas.Height = width, height
	sx, sy := float64(width)/640, float64(height)/360
	xy := func(v authoring.Point) authoring.Point { return authoring.Point{X: v.X * sx, Y: v.Y * sy} }
	for i := range p.Layers {
		l := &p.Layers[i]
		if b := l.Background; b != nil {
			b.Scale = xy(b.Scale)
			b.Velocity = xy(b.Velocity)
		}
		if g := l.Sprites; g != nil {
			g.Scale = authoring.Point{X: sx, Y: sy}
			g.Weave.Center = xy(g.Weave.Center)
			g.Weave.HorizontalAmplitude *= sx
			g.Weave.VerticalAmplitude *= sy
			g.Weave.VerticalSecondAmplitude *= sy
		}
		if c := l.Scroll; c != nil {
			c.Origin = xy(c.Origin)
			c.Speed *= sx
			c.Gap *= sx
			for name, m := range c.Modes {
				if m.Wave != nil {
					m.Wave.Amplitude *= sy
					m.Wave.Spatial /= sx
				}
				if m.Zoom != nil {
					m.Zoom.Pivot = xy(m.Zoom.Pivot)
				}
				if v := m.Perspective; v != nil {
					v.Focal *= sx
					v.Near *= sx
					v.Depth *= sx
					v.Center = xy(v.Center)
					v.DepthWave.Amplitude *= sx
					v.DepthWave.Spatial /= sx
					v.VerticalWave.Amplitude *= sy
					v.VerticalWave.Spatial /= sx
				}
				if m.Path != nil {
					for j := range m.Path.Points {
						m.Path.Points[j] = xy(m.Path.Points[j])
					}
				}
				c.Modes[name] = m
			}
		}
	}
	sway := p.Signals["sway"]
	sway.Oscillators[0].Amplitude *= sy
	p.Signals["sway"] = sway
	return New(&p)
}

// SurfaceBytes estimates logical RGBA storage, excluding driver/atlas overhead.
func (s *Scene) SurfaceBytes() int64 {
	bytes := int64(s.Width) * int64(s.Height) * 4
	for _, img := range s.images {
		bytes += int64(img.Bounds().Dx()*img.Bounds().Dy()) * 4
	}
	return bytes
}

// New accepts nil for the default project. The default takes a JSON round trip
// before compilation, so the visual demonstration exercises the persisted data.
func New(project *authoring.Project) (*Scene, error) {
	p := DefaultProject()
	if project != nil {
		p = *project
	}
	var saved bytes.Buffer
	if err := authoring.Encode(&saved, p); err != nil {
		return nil, err
	}
	decoded, err := authoring.Decode(&saved)
	if err != nil {
		return nil, err
	}
	s := &Scene{Project: *decoded, Width: p.Canvas.Width, Height: p.Canvas.Height}
	assets, err := s.assets()
	if err != nil {
		s.Close()
		return nil, err
	}
	s.Compiled, err = authoring.Compile(*decoded, assets, authoring.Options{})
	if err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func (s *Scene) Close() error {
	var err error
	if s.Compiled != nil {
		err = s.Compiled.Close()
		s.Compiled = nil
	}
	for _, img := range s.images {
		img.Deallocate()
	}
	s.images = nil
	return err
}

func (s *Scene) texture(pixels image.Image) *ebiten.Image {
	v := ebiten.NewImageFromImage(pixels)
	s.images = append(s.images, v)
	return v
}
func (s *Scene) assets() (authoring.Assets, error) {
	assets := authoring.Assets{Images: map[string]*ebiten.Image{}, Fonts: map[string]scrolling.Face{}}
	tile := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := range 32 {
		for x := range 32 {
			c := color.RGBA{R: 9, G: 15, B: 34, A: 255}
			if (x/8+y/8)%2 == 0 {
				c = color.RGBA{R: 13, G: 24, B: 47, A: 255}
			}
			tile.SetRGBA(x, y, c)
		}
	}
	assets.Images["checker"] = s.texture(tile)
	for i, name := range []string{"orb-cyan", "orb-pink"} {
		pixels := image.NewNRGBA(image.Rect(0, 0, 24, 24))
		for y := range 24 {
			for x := range 24 {
				r := math.Hypot(float64(x)-11.5, float64(y)-11.5)
				if r > 11 {
					continue
				}
				shade := uint8(100 + 155*(1-r/11))
				c := color.NRGBA{R: 25, G: shade, B: 255, A: 255}
				if i == 1 {
					c = color.NRGBA{R: 255, G: 50, B: shade, A: 255}
				}
				pixels.SetNRGBA(x, y, c)
			}
		}
		assets.Images[name] = s.texture(pixels)
	}
	order := ""
	for ch := 32; ch < 127; ch++ {
		order += string(rune(ch))
	}
	for _, name := range []string{"small", "bright"} {
		pixels := image.NewRGBA(image.Rect(0, 0, 16*8, 6*16))
		ink := color.RGBA{R: 120, G: 225, B: 255, A: 255}
		if name == "bright" {
			ink = color.RGBA{R: 255, G: 200, B: 100, A: 255}
		}
		drawer := font.Drawer{Dst: pixels, Src: image.NewUniform(ink), Face: basicfont.Face7x13}
		for i, ch := range order {
			drawer.Dot = fixed.P(i%16*8, i/16*16+12)
			drawer.DrawString(string(ch))
		}
		atlas := s.texture(pixels)
		metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(8, 16), Columns: 16, Order: order, Fallback: '?'})
		if err != nil {
			return assets, err
		}
		scale := 2.0
		if name == "bright" {
			scale = 2.5
		}
		assets.Fonts[name] = scrolling.Face{Atlas: atlas, Metrics: metrics, ScaleX: scale * float64(s.Width) / 640, ScaleY: scale * float64(s.Height) / 360}
	}
	return assets, nil
}

func DefaultProject() authoring.Project {
	return authoring.Project{
		Version: authoring.Version, Units: authoring.DefaultUnits(), Canvas: authoring.Canvas{Width: 640, Height: 360, TPS: 60, Background: [4]uint8{5, 8, 20, 255}}, BPM: 120,
		Assets: map[string]string{"checker": "image", "orb-cyan": "image", "orb-pink": "image", "small": "font", "bright": "font"},
		Signals: map[string]modulation.Spec{
			"pulse": {Base: 1, TimeBase: modulation.Beats, Oscillators: []modulation.Oscillator{{Shape: modulation.Cosine, Amplitude: .12, Frequency: 1}}},
			"sway":  {Oscillators: []modulation.Oscillator{{Shape: modulation.Sine, Amplitude: 8, Frequency: .25}}},
		},
		Layers: []authoring.Layer{
			{ID: "moving-background", Kind: "background", Background: &authoring.Background{Image: "checker", Period: authoring.Point{X: 32, Y: 32}, Scale: authoring.Point{X: 2, Y: 2}, Velocity: authoring.Point{X: -12, Y: 6}}},
			{ID: "orb-formation", Kind: "sprites", Window: authoring.Window{FadeIn: 1}, Sprites: &authoring.SpriteGroup{Images: []string{"orb-cyan", "orb-pink"}, Count: 24, FrameStride: 1, Speed: 25, Delay: .06, Anchor: authoring.Point{X: .5, Y: .5}, Weave: &authoring.Weave{Center: authoring.Point{X: 320, Y: 125}, HorizontalAmplitude: 250, VerticalAmplitude: 30, VerticalSecondAmplitude: 22, HorizontalPeriod: 25, EnvelopePeriod: 300, VerticalPeriod: 37, VerticalSecondPeriod: 17, Spacing: 7}, Signals: authoring.Bindings{"scaleX": "pulse", "scaleY": "pulse", "y": "sway"}}},
			{ID: "mixed-scroll", Kind: "scroll", Window: authoring.Window{Start: 1, FadeIn: .5}, LocalTime: true, Scroll: &authoring.Scroll{Text: "DCK: ONE SAVED PROJECT, SHARED EFFECTS.  {font:bright}MIXED FONTS{font:small}, TIMED MODES, SPRITE FORMATIONS AND BACKGROUNDS.   ", Fonts: map[string]string{"small": "small", "bright": "bright"}, Font: "small", Controls: "braces", Origin: authoring.Point{X: 640, Y: 244}, Speed: 100, Gap: 140, Repeat: true, Modes: map[string]authoring.Mode{
				"plain": {Kind: "normal"}, "sine": {Kind: "sine", Wave: &authoring.Wave{Amplitude: 18, Spatial: .02, Speed: 2}}, "bounce": {Kind: "bounce", Wave: &authoring.Wave{Amplitude: 22, Speed: 3}}, "zoom": {Kind: "zoom", Zoom: &authoring.Zoom{Base: authoring.Point{X: 1, Y: 1}, Pivot: authoring.Point{X: 320, Y: 260}, Wave: authoring.Wave{Amplitude: .2, Speed: 2}}}, "perspective": {Kind: "perspective", Perspective: &authoring.Perspective{Focal: 400, Near: 10, Depth: 50, Center: authoring.Point{X: 320, Y: 260}, DepthWave: authoring.Wave{Amplitude: 110, Spatial: .012, Speed: 1.5}, VerticalWave: authoring.Wave{Amplitude: 15, Spatial: .012, Speed: 1.5}}}, "path": {Kind: "path", Path: &authoring.Path{Points: []authoring.Point{{X: -200, Y: 245}, {X: 80, Y: 220}, {X: 320, Y: 290}, {X: 560, Y: 220}, {X: 850, Y: 245}}, SplineSamples: 24, Orient: true, Clip: true}}}, Sequence: []authoring.Cue{{At: 0, Mode: "plain"}, {At: 3, Mode: "sine"}, {At: 6, Mode: "bounce"}, {At: 9, Mode: "zoom"}, {At: 12, Mode: "perspective"}, {At: 15, Mode: "path"}}, SequencePeriod: 18}},
			{ID: "title", Kind: "scroll", Scroll: &authoring.Scroll{Text: "DCK / SERIALIZABLE COMPOSITION", Fonts: map[string]string{"small": "small"}, Font: "small", Origin: authoring.Point{X: 72, Y: 20}}},
		},
	}
}
