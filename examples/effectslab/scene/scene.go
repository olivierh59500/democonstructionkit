// Package scene composes a self-contained desktop and mobile effects laboratory.
package scene

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/timeline"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var pathNames = [...]string{"SINE CURVE", "COORDINATE SPLINE", "CUSTOM FUNCTION"}

// Scene borrows no external assets. All images are generated once in Go.
type Scene struct {
	Width, Height           int
	Eco                     bool
	Path                    int
	AutoPath                bool
	Water, Lens             bool
	frame                   kit.Frame
	scale                   float64
	background, logo, atlas *ebiten.Image
	labels                  map[string]*ebiten.Image
	scroll                  *scrolling.Scrolling
	paths                   [3]*motion.Path
	warp                    *composite.CellWarp
	water                   *composite.WaterReflection
	lens                    *composite.Magnifier
	pipeline                *kit.Pipeline
	lensWindow              timeline.Window
	closed                  bool
}

func New(eco bool) (_ *Scene, err error) {
	s := &Scene{Width: 640, Height: 360, scale: 1, Eco: eco, Water: true, Lens: true, AutoPath: true,
		labels: make(map[string]*ebiten.Image), lensWindow: timeline.Window{Start: 4, Duration: 6, FadeIn: .5, FadeOut: .8}}
	if eco {
		s.Width, s.Height, s.scale = 320, 180, .5
	}
	defer func() {
		if err != nil {
			_ = s.Close()
		}
	}()
	s.background = makeBackground(s.Width, s.Height)
	s.logo = makeLabel("DCK EFFECTS", 3, color.NRGBA{R: 100, G: 235, B: 255, A: 255})
	if eco {
		// Logo cells keep the same relative size in both render resolutions.
		s.logo.Deallocate()
		s.logo = makeLabel("DCK EFFECTS", 1.5, color.NRGBA{R: 100, G: 235, B: 255, A: 255})
	}
	for _, label := range []string{"DEMOCONSTRUCTIONKIT / LIVE LAYERS", "SINE CURVE", "COORDINATE SPLINE", "CUSTOM FUNCTION", "PATH", "WATER ON", "WATER OFF", "LENS AUTO", "LENS OFF", "HIGH", "ECO", "640x360 / ROW 1", "320x180 / ROW 4", "REFLECTING LOGO + SCROLL", "CPU PROFILE FINISHED"} {
		s.labels[label] = makeLabel(label, 1, color.NRGBA{R: 200, G: 215, B: 240, A: 255})
	}
	s.atlas = makeAtlas()
	order := ""
	for r := rune(32); r < 127; r++ {
		order += string(r)
	}
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: s.atlas.Bounds(), Cell: image.Pt(8, 14), Columns: 16, Order: order})
	if err != nil {
		return nil, err
	}
	s.paths[0], err = motion.SamplePath(motion.SineCurve(motion.Point{X: 16 * s.scale, Y: 202 * s.scale}, 608*s.scale, 18*s.scale, 1.5, 0), 192, false)
	if err != nil {
		return nil, err
	}
	s.paths[1], err = motion.NewSpline([]motion.Point{{X: 16 * s.scale, Y: 208 * s.scale}, {X: 130 * s.scale, Y: 176 * s.scale}, {X: 280 * s.scale, Y: 213 * s.scale}, {X: 440 * s.scale, Y: 185 * s.scale}, {X: 624 * s.scale, Y: 210 * s.scale}}, false, 48)
	if err != nil {
		return nil, err
	}
	s.paths[2], err = motion.SamplePath(func(u float64) motion.Point {
		return motion.Point{X: (16 + 608*u) * s.scale, Y: (200 + 15*math.Sin(5*math.Pi*u) + 7*math.Sin(9*math.Pi*u)) * s.scale}
	}, 192, false)
	if err != nil {
		return nil, err
	}
	modes := make(map[string]scrolling.Mode, 3)
	for i, path := range s.paths {
		mode, e := scrolling.AlongPath(scrolling.PathConfig{Path: path, Orient: true, Clip: true})
		if e != nil {
			return nil, e
		}
		modes[pathNames[i]] = mode
	}
	s.scroll, err = scrolling.New(scrolling.Config{Text: " ONE SCROLL - ANY PATH - ANY FONT - LIVE REFLECTIONS - TEMPORARY LENSES - ",
		Fonts: map[string]scrolling.Face{"default": {Atlas: s.atlas, Metrics: metrics, ScaleX: 2 * s.scale, ScaleY: 2 * s.scale}}, Modes: modes})
	if err != nil {
		return nil, err
	}
	// This deformation only depends on X: one full-height cell per column avoids
	// submitting redundant cells. Economy mode uses half as many columns.
	s.warp, err = composite.NewCellWarp(composite.CellWarpConfig{Cell: image.Pt(8, s.logo.Bounds().Dy()), Sample: func(row, column int, f kit.Frame) composite.CellTransform {
		pose := composite.CellTransform{}
		position := float64(column) / s.scale
		pose.GeoM.Translate(0, 7*s.scale*math.Sin(position*.35+f.Time*2.2))
		pose.ColorScale.Scale(1, float32(.8+.2*math.Sin(f.Time+position*.2)), 1, 1)
		return pose
	}})
	if err != nil {
		return nil, err
	}
	c := composite.DefaultWaterReflectionConfig()
	c.Source = image.Rect(0, int(100*s.scale), s.Width, int(250*s.scale))
	c.Horizon = 258 * s.scale
	c.ScaleY = .48
	c.Alpha = .75
	c.Fade = .8
	c.Wave = composite.WaterWave{Amplitude: 4 * s.scale, Wavelength: 24 * s.scale, Speed: 2.2, Phase: .4}
	if eco {
		c.RowHeight = 4
	}
	c.Tint.Scale(.65, .85, 1, 1)
	s.water, err = composite.NewWaterReflection(c)
	if err != nil {
		return nil, err
	}
	s.lens, err = composite.NewMagnifier()
	if err != nil {
		return nil, err
	}
	layers, err := kit.NewLayers(s.Width, s.Height,
		kit.TimedLayer{Effect: kit.Func{OnDraw: func(dst *ebiten.Image) { dst.DrawImage(s.background, nil) }}},
		kit.TimedLayer{Effect: kit.Func{OnDraw: s.drawLogo}, Window: timeline.Window{FadeIn: 1}},
		kit.TimedLayer{Effect: kit.Func{OnDraw: s.drawScroll}, Window: timeline.Window{Start: .3, FadeIn: .6}},
	)
	if err != nil {
		return nil, err
	}
	s.pipeline, err = kit.NewPipeline(layers, s.Width, s.Height,
		kit.ImagePass{Enabled: func(kit.Frame) bool { return s.Water }, Apply: func(dst, source *ebiten.Image, f kit.Frame) { dst.DrawImage(source, nil); s.water.Draw(dst, source, f) }},
		kit.ImagePass{Enabled: func(f kit.Frame) bool {
			_, alpha, on := s.lensWindow.At(math.Mod(f.Time, 14))
			return s.Lens && on && alpha > 0
		}, Apply: func(dst, source *ebiten.Image, f kit.Frame) {
			dst.DrawImage(source, nil)
			_, alpha, _ := s.lensWindow.At(math.Mod(f.Time, 14))
			op := composite.DefaultMagnifierOptions()
			op.CenterX = (320 + 150*math.Sin(f.Time*.7)) * s.scale
			op.CenterY = (172 + 60*math.Sin(f.Time*.95)) * s.scale
			op.Radius = 55 * s.scale
			op.Zoom = 2.4
			op.Feather = 3 * s.scale
			op.Opacity = alpha
			s.lens.Draw(dst, source, op)
		}},
	)
	if err != nil {
		_ = layers.Close()
		return nil, err
	}
	return s, nil
}

func (s *Scene) Update(f kit.Frame) error {
	if s.closed {
		return nil
	}
	s.frame = f
	if s.AutoPath {
		s.Path = int(f.Time/8) % len(s.paths)
	}
	return s.pipeline.Update(f)
}
func (s *Scene) Draw(dst *ebiten.Image) {
	if s.closed {
		return
	}
	s.pipeline.Draw(dst)
	s.drawHUD(dst)
}
func (s *Scene) drawLogo(dst *ebiten.Image) {
	x := (float64(s.Width)-float64(s.logo.Bounds().Dx()))/2 + 10*s.scale*math.Sin(s.frame.Time*.7)
	s.warp.DrawAt(dst, s.logo, s.frame, x, 112*s.scale)
}
func (s *Scene) drawScroll(dst *ebiten.Image) {
	// Window provides a bounded circular character range. Movement stays continuous
	// across message boundaries without restarting the scroller's time controls.
	position := s.frame.Time * 60 * s.scale
	advance := 16 * s.scale
	first := int(math.Floor(position / advance))
	state := s.scroll.Window(first, s.paths[s.Path].Length()+advance)
	state.X -= math.Mod(position, advance)
	state.Time = s.frame.Time
	state.Shape = pathNames[s.Path]
	s.scroll.DrawAt(dst, state)
}
func (s *Scene) drawHUD(dst *ebiten.Image) {
	s.label(dst, "DEMOCONSTRUCTIONKIT / LIVE LAYERS", 12*s.scale, 12*s.scale)
	pathY := max(38*s.scale, 12*s.scale+15)
	s.label(dst, pathNames[s.Path], 12*s.scale, pathY)
	s.label(dst, "REFLECTING LOGO + SCROLL", 12*s.scale, 244*s.scale)
	water, lens, quality := "WATER OFF", "LENS OFF", "640x360 / ROW 1"
	if s.Water {
		water = "WATER ON"
	}
	if s.Lens {
		lens = "LENS AUTO"
	}
	if s.Eco {
		quality = "320x180 / ROW 4"
	}
	s.label(dst, quality, 12*s.scale, max(58*s.scale, pathY+15))
	quality = "HIGH"
	if s.Eco {
		quality = "ECO"
	}
	labels := [4]string{"PATH", water, lens, quality}
	for i, label := range labels {
		s.label(dst, label, float64(i*s.Width/4+4), float64(s.Height-15))
	}
}
func (s *Scene) label(dst *ebiten.Image, label string, x, y float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	dst.DrawImage(s.labels[label], &op)
}

// SurfaceInventory reports logical image storage, not physical GPU allocation:
// Ebitengine may use larger atlases and backend command buffers internally.
func (s *Scene) SurfaceInventory() map[string]int64 {
	labels := int64(0)
	for _, img := range s.labels {
		labels += int64(img.Bounds().Dx() * img.Bounds().Dy() * 4)
	}
	return map[string]int64{"pipeline_ping_pong_rgba": int64(s.Width * s.Height * 4 * 2), "layer_fade_surface_rgba": int64(s.Width * s.Height * 4), "background_rgba": int64(s.Width * s.Height * 4), "atlas_rgba": int64(s.atlas.Bounds().Dx() * s.atlas.Bounds().Dy() * 4), "logo_rgba": int64(s.logo.Bounds().Dx() * s.logo.Bounds().Dy() * 4), "labels_rgba": labels}
}
func (s *Scene) Close() error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	if s.pipeline != nil {
		_ = s.pipeline.Close()
	}
	if s.water != nil {
		_ = s.water.Close()
	}
	if s.lens != nil {
		_ = s.lens.Close()
	}
	if s.warp != nil {
		_ = s.warp.Close()
	}
	for _, img := range []*ebiten.Image{s.background, s.logo, s.atlas} {
		if img != nil {
			img.Deallocate()
		}
	}
	for _, img := range s.labels {
		img.Deallocate()
	}
	return nil
}

func makeAtlas() *ebiten.Image {
	pixels := image.NewRGBA(image.Rect(0, 0, 16*8, 6*14))
	drawer := font.Drawer{Dst: pixels, Src: image.NewUniform(color.NRGBA{R: 255, G: 207, B: 118, A: 255}), Face: basicfont.Face7x13}
	for i := 0; i < 95; i++ {
		drawer.Dot = fixed.P((i%16)*8, (i/16)*14+11)
		drawer.DrawString(string(rune(32 + i)))
	}
	return ebiten.NewImageFromImage(pixels)
}
func makeLabel(label string, scale float64, c color.NRGBA) *ebiten.Image {
	base := image.NewRGBA(image.Rect(0, 0, max(1, len(label)*7), 14))
	drawer := font.Drawer{Dst: base, Src: image.NewUniform(c), Face: basicfont.Face7x13, Dot: fixed.P(0, 11)}
	drawer.DrawString(label)
	if scale == 1 {
		return ebiten.NewImageFromImage(base)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, int(math.Ceil(float64(base.Bounds().Dx())*scale)), int(math.Ceil(14*scale))))
	for y := range pixels.Bounds().Dy() {
		for x := range pixels.Bounds().Dx() {
			pixels.SetRGBA(x, y, base.RGBAAt(int(float64(x)/scale), int(float64(y)/scale)))
		}
	}
	return ebiten.NewImageFromImage(pixels)
}
func makeBackground(width, height int) *ebiten.Image {
	pixels := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			v := float64(y) / float64(height)
			c := color.RGBA{R: uint8(7 + 5*v), G: uint8(12 + 10*v), B: uint8(26 + 22*v), A: 255}
			if y > height*258/360 {
				c.B += 10
			}
			pixels.SetRGBA(x, y, c)
		}
	}
	for i := range 70 {
		x := (i*173 + 71) % width
		y := (i*i*31 + 11) % (height * 230 / 360)
		pixels.SetRGBA(x, y, color.RGBA{R: 45, G: 67, B: 105, A: 255})
	}
	return ebiten.NewImageFromImage(pixels)
}

func (s *Scene) String() string { return fmt.Sprintf("%dx%d", s.Width, s.Height) }
