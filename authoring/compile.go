package authoring

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// Resolver returns borrowed assets. Compile validates returned images and font
// metrics but never destroys them, including on partial compilation failure.
type Resolver interface {
	Image(id string) (*ebiten.Image, error)
	Font(id string) (scrolling.Face, error)
}
type Assets struct {
	Images map[string]*ebiten.Image
	Fonts  map[string]scrolling.Face
}

func (a Assets) Image(id string) (*ebiten.Image, error) {
	v := a.Images[id]
	if v == nil {
		return nil, fmt.Errorf("image %q unavailable", id)
	}
	return v, nil
}
func (a Assets) Font(id string) (scrolling.Face, error) {
	v, ok := a.Fonts[id]
	if !ok {
		return v, fmt.Errorf("font %q unavailable", id)
	}
	return v, nil
}

// Options.Context can supply audible music time, beats and named signal inputs.
// Without it the compiler uses layer-local visual time and the project's BPM.
type Options struct {
	Context func(kit.Frame) modulation.Context
}

type Compiled struct {
	canvas Canvas
	layers *kit.Layers
	closed bool
}

func (c *Compiled) Canvas() Canvas { return c.canvas }
func (c *Compiled) Update(f kit.Frame) error {
	if c.closed {
		return nil
	}
	return c.layers.Update(f)
}
func (c *Compiled) Draw(dst *ebiten.Image) {
	if c.closed || dst == nil {
		return
	}
	v := c.canvas.Background
	dst.Fill(color.NRGBA{R: v[0], G: v[1], B: v[2], A: v[3]})
	c.layers.Draw(dst)
}
func (c *Compiled) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.layers.Close()
}

// Compile creates existing DCK effects from validated data. No rendering or
// animation equations are reimplemented here; named configurations only select
// and connect library primitives. Close releases working surfaces, never assets.
func Compile(p Project, r Resolver, o Options) (*Compiled, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if r == nil {
		return nil, fmt.Errorf("authoring: nil asset resolver")
	}
	b := compiler{resolver: r, signals: map[string]*modulation.Signal{}, images: map[string]*ebiten.Image{}, fonts: map[string]scrolling.Face{}, context: o.Context, canvas: p.Canvas}
	if b.context == nil {
		bpm := p.BPM
		b.context = func(f kit.Frame) modulation.Context { return modulation.MusicContext(f.Time, bpm, 0, nil) }
	}
	for id, spec := range p.Signals {
		signal, err := modulation.New(spec)
		if err != nil {
			return nil, err
		}
		b.signals[id] = signal
	}
	layers := make([]kit.TimedLayer, 0, len(p.Layers))
	cleanup := func() {
		for _, l := range layers {
			_ = kit.Close(l.Effect)
		}
	}
	for _, l := range p.Layers {
		effect, err := b.layer(l)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("authoring: layer %q: %w", l.ID, err)
		}
		blendMode, _ := blend(l.Blend)
		layers = append(layers, kit.TimedLayer{Effect: effect, Window: window(l.Window), LocalTime: l.LocalTime, Blend: blendMode})
	}
	root, err := kit.NewLayers(p.Canvas.Width, p.Canvas.Height, layers...)
	if err != nil {
		cleanup()
		return nil, err
	}
	return &Compiled{canvas: p.Canvas, layers: root}, nil
}

type compiler struct {
	canvas   Canvas
	resolver Resolver
	images   map[string]*ebiten.Image
	fonts    map[string]scrolling.Face
	signals  map[string]*modulation.Signal
	context  func(kit.Frame) modulation.Context
}

func (b *compiler) image(id string) (*ebiten.Image, error) {
	if v := b.images[id]; v != nil {
		return v, nil
	}
	v, err := b.resolver.Image(id)
	if err != nil {
		return nil, err
	}
	if v == nil || v.Bounds().Empty() {
		return nil, fmt.Errorf("empty image %q", id)
	}
	b.images[id] = v
	return v, nil
}
func (b *compiler) face(id string) (scrolling.Face, error) {
	if v, ok := b.fonts[id]; ok {
		return v, nil
	}
	v, err := b.resolver.Font(id)
	if err != nil {
		return v, err
	}
	if v.Atlas == nil || v.Metrics == nil || !v.Metrics.Bounds().In(v.Atlas.Bounds()) {
		return v, fmt.Errorf("invalid font %q", id)
	}
	b.fonts[id] = v
	return v, nil
}
func (b *compiler) layer(l Layer) (kit.Effect, error) {
	switch l.Kind {
	case "jelly_cube":
		c, err := jellyConfig(*l.JellyCube)
		if err != nil {
			return nil, err
		}
		return effects.NewJellyCube(c)
	case "scroll":
		return b.scroll(*l.Scroll)
	case "sprites":
		return b.sprites(*l.Sprites)
	case "background":
		return b.background(*l.Background)
	case "rotozoom":
		return b.rotozoom(*l.Rotozoom)
	}
	return nil, fmt.Errorf("unknown effect %q", l.Kind)
}

func (b *compiler) scroll(c Scroll) (kit.Effect, error) {
	config := scrolling.Config{Text: c.Text, Font: c.Font, Fonts: map[string]scrolling.Face{}, Speed: c.Speed, Gap: c.Gap, X: c.Origin.X, Y: c.Origin.Y, Advance: c.Advance, Vertical: c.Vertical, Repeat: c.Repeat, Shape: c.Mode, Modes: map[string]scrolling.Mode{}}
	if c.Controls == "braces" {
		config.Controls = scrolltext.Braces
	}
	if c.RepeatBounds != nil {
		config.RepeatBounds = rectangle(*c.RepeatBounds)
	}
	for name, id := range c.Fonts {
		face, err := b.face(id)
		if err != nil {
			return nil, err
		}
		config.Fonts[name] = face
	}
	if c.Page != nil {
		align, _ := alignment(c.Page.Align)
		config.Page = &scrolling.PageConfig{Width: c.Page.Width, LineHeight: c.Page.LineHeight, Align: align}
	}
	for name, spec := range c.Modes {
		m, err := compileMode(spec, c.Vertical || c.Page != nil)
		if err != nil {
			return nil, err
		}
		config.Modes[name] = m
	}
	if len(c.Sequence) > 0 {
		cues := make([]scrolling.Cue, len(c.Sequence))
		for i, cue := range c.Sequence {
			cues[i] = scrolling.Cue{At: cue.At, Mode: cue.Mode}
		}
		sequence, err := scrolling.NewModeSequence(cues, c.SequencePeriod)
		if err != nil {
			return nil, err
		}
		config.Sequence = sequence
	}
	scroll, err := scrolling.New(config)
	if err != nil {
		return nil, err
	}
	if err = scroll.ValidateRenderBounds(image.Rect(0, 0, b.canvas.Width, b.canvas.Height)); err != nil {
		_ = scroll.Close()
		return nil, err
	}
	return scroll, nil
}

func compileMode(c Mode, vertical bool) (scrolling.Mode, error) {
	payloads := 0
	if c.Wave != nil {
		payloads++
	}
	if c.Zoom != nil {
		payloads++
	}
	if c.Perspective != nil {
		payloads++
	}
	if c.Path != nil {
		payloads++
	}
	if c.Kind == "normal" {
		if payloads != 0 {
			return scrolling.Mode{}, fmt.Errorf("normal mode takes no payload")
		}
		return scrolling.Normal(), nil
	}
	if payloads != 1 {
		return scrolling.Mode{}, fmt.Errorf("mode %q requires exactly one payload", c.Kind)
	}
	switch c.Kind {
	case "sine":
		if c.Wave != nil {
			return scrolling.Sine(wave(*c.Wave)), nil
		}
	case "bounce":
		if c.Wave != nil {
			if c.Wave.Spatial != 0 {
				return scrolling.Mode{}, fmt.Errorf("bounce has no spatial frequency")
			}
			return scrolling.Bounce(wave(*c.Wave)), nil
		}
	case "zoom":
		if c.Zoom != nil {
			v := c.Zoom
			return scrolling.Zoom(scrolling.ZoomConfig{BaseX: v.Base.X, BaseY: v.Base.Y, Wave: wave(v.Wave), PivotX: v.Pivot.X, PivotY: v.Pivot.Y}), nil
		}
	case "perspective":
		if c.Perspective != nil {
			v := c.Perspective
			return scrolling.Perspective(scrolling.PerspectiveConfig{Focal: v.Focal, Near: v.Near, Depth: v.Depth, CenterX: v.Center.X, CenterY: v.Center.Y, DepthWave: wave(v.DepthWave), VerticalWave: wave(v.VerticalWave)})
		}
	case "path":
		if c.Path != nil {
			path, err := compilePath(*c.Path)
			if err != nil {
				return scrolling.Mode{}, err
			}
			v := c.Path
			return scrolling.AlongPath(scrolling.PathConfig{Path: path, Offset: v.Offset, NormalOffset: v.NormalOffset, Rotation: v.Rotation, Orient: v.Orient, Clip: v.Clip, Vertical: vertical})
		}
	}
	return scrolling.Mode{}, fmt.Errorf("unknown or mismatched mode %q", c.Kind)
}

func compilePath(c Path) (*motion.Path, error) {
	if len(c.Points) < 2 || len(c.Points) > 4096 || c.SplineSamples < 0 || c.SplineSamples > 4096 || len(c.Points)*max(1, c.SplineSamples) > 65536 {
		return nil, fmt.Errorf("invalid path points or spline resolution")
	}
	points := points(c.Points)
	if c.SplineSamples > 0 {
		return motion.NewSpline(points, c.Closed, c.SplineSamples)
	}
	return motion.NewPolyline(points, c.Closed)
}
func points(values []Point) []motion.Point {
	out := make([]motion.Point, len(values))
	for i, p := range values {
		out[i] = point(p)
	}
	return out
}
func point(p Point) motion.Point { return motion.Point{X: p.X, Y: p.Y} }
func wave(w Wave) motion.Wave {
	return motion.Wave{Amplitude: w.Amplitude, Spatial: w.Spatial, Speed: w.Speed, Phase: w.Phase}
}
func rectangle(r Rect) image.Rectangle { return image.Rect(r.X, r.Y, r.X+r.Width, r.Y+r.Height) }

func (b *compiler) sprites(c SpriteGroup) (kit.Effect, error) {
	f, _ := filter(c.Filter)
	blendMode, _ := blend(c.Blend)
	config := sprites.GroupConfig{Count: c.Count, FPS: c.FPS, FrameStride: c.FrameStride, FrameOffset: c.FrameOffset, Points: points(c.Points), Closed: c.Closed, SplineSamples: c.SplineSamples, Origin: point(c.Origin), Velocity: point(c.Velocity), Spacing: point(c.Spacing), Speed: c.Speed, Phase: c.Phase, PhaseSpacing: c.PhaseSpacing, Delay: c.Delay, Orient: c.Orient, ScaleX: c.Scale.X, ScaleY: c.Scale.Y, Angle: c.Angle, AnchorX: c.Anchor.X, AnchorY: c.Anchor.Y, Filter: f, Blend: blendMode, Reverse: c.Reverse, Signals: b.bindings(c.Signals), Context: b.context}
	for _, id := range c.Images {
		img, err := b.image(id)
		if err != nil {
			return nil, err
		}
		config.Frames = append(config.Frames, img)
	}
	for _, bindings := range c.PerInstance {
		config.PerInstance = append(config.PerInstance, b.bindings(bindings))
	}
	if c.Orbit != nil {
		v := c.Orbit
		config.Orbit = &motion.NestedOrbit{Center: point(v.Center), Radius: point(v.Radius), XRate: v.XRate, YRate: v.YRate, ModulationRate: v.ModulationRate, ModulationPhase: v.ModulationPhase, ModulationDepth: v.ModulationDepth}
	}
	if c.Weave != nil {
		v := c.Weave
		config.Weave = &motion.Weave{Center: point(v.Center), HorizontalAmplitude: v.HorizontalAmplitude, VerticalAmplitude: v.VerticalAmplitude, VerticalSecondAmplitude: v.VerticalSecondAmplitude, HorizontalPeriod: v.HorizontalPeriod, EnvelopePeriod: v.EnvelopePeriod, VerticalPeriod: v.VerticalPeriod, VerticalSecondPeriod: v.VerticalSecondPeriod, Spacing: v.Spacing}
	}
	return sprites.NewGroup(config)
}
func (b *compiler) bindings(values Bindings) sprites.GroupSignals {
	return sprites.GroupSignals{X: b.signals[values["x"]], Y: b.signals[values["y"]], ScaleX: b.signals[values["scaleX"]], ScaleY: b.signals[values["scaleY"]], Angle: b.signals[values["angle"]], Opacity: b.signals[values["opacity"]]}
}

func (b *compiler) background(c Background) (kit.Effect, error) {
	img, err := b.image(c.Image)
	if err != nil {
		return nil, err
	}
	f, _ := filter(c.Filter)
	blendMode, _ := blend(c.Blend)
	config := composite.BackgroundConfig{PeriodX: c.Period.X, PeriodY: c.Period.Y, CopiesX: c.CopiesX, CopiesY: c.CopiesY, ScaleX: c.Scale.X, ScaleY: c.Scale.Y, ParallaxX: c.Parallax.X, ParallaxY: c.Parallax.Y, Filter: f, Blend: blendMode}
	if c.Source != nil {
		config.Source = rectangle(*c.Source)
		if !config.Source.In(img.Bounds()) {
			return nil, fmt.Errorf("background crop exceeds image bounds")
		}
	}
	sourceWidth, sourceHeight := img.Bounds().Dx(), img.Bounds().Dy()
	if c.Source != nil {
		sourceWidth, sourceHeight = c.Source.Width, c.Source.Height
	}
	if err := backgroundBudget(c, sourceWidth, sourceHeight, b.canvas); err != nil {
		return nil, err
	}
	renderer, err := composite.NewBackground(config)
	if err != nil {
		return nil, err
	}
	return &composite.BackgroundLayer{Renderer: renderer, Image: img, VelocityX: c.Velocity.X, VelocityY: c.Velocity.Y, Sample: func(f kit.Frame) composite.BackgroundPose {
		return composite.BackgroundPose{X: c.Origin.X, Y: c.Origin.Y, CameraX: c.Camera.X + c.CameraVelocity.X*f.Time, CameraY: c.Camera.Y + c.CameraVelocity.Y*f.Time}
	}}, nil
}

func (b *compiler) rotozoom(c Rotozoom) (kit.Effect, error) {
	img, err := b.image(c.Image)
	if err != nil {
		return nil, err
	}
	f, _ := filter(c.Filter)
	return composite.NewRotozoomBackground(composite.RotozoomBackgroundConfig{
		Image: img,
		Pose: composite.Repetition{CenterX: c.Center.X, CenterY: c.Center.Y, Zoom: c.Zoom, Rotation: c.Rotation,
			PhaseX: c.Phase.X, PhaseY: c.Phase.Y, Filter: f},
		Velocity: composite.RotozoomVelocity{CenterX: c.CenterVelocity.X, CenterY: c.CenterVelocity.Y,
			Zoom: c.ZoomVelocity, Rotation: c.RotationVelocity, PhaseX: c.PhaseVelocity.X, PhaseY: c.PhaseVelocity.Y},
	})
}

// Version 1 limits fallback copies as well as canvas and formation sizes. The
// budget is conservative over all offsets, including transparent overlaps.
func backgroundBudget(c Background, width, height int, canvas Canvas) error {
	sx, sy := c.Scale.X, c.Scale.Y
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	count := func(extent int, period, scale float64, viewport, copies int) float64 {
		if period == 0 {
			return 1
		}
		visible := math.Ceil((float64(viewport)+float64(extent)*scale)/(period*scale)) + 1
		if copies > 0 {
			return math.Min(visible, float64(copies))
		}
		return visible
	}
	x, y := count(width, c.Period.X, sx, canvas.Width, c.CopiesX), count(height, c.Period.Y, sy, canvas.Height, c.CopiesY)
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || x*y > 16384 {
		return fmt.Errorf("background exceeds 16384 visible-copy budget")
	}
	return nil
}

func filter(name string) (ebiten.Filter, error) {
	switch name {
	case "", "nearest":
		return ebiten.FilterNearest, nil
	case "linear":
		return ebiten.FilterLinear, nil
	}
	return 0, fmt.Errorf("unknown filter %q", name)
}
func blend(name string) (ebiten.Blend, error) {
	switch name {
	case "", "source-over":
		return ebiten.BlendSourceOver, nil
	case "add":
		return ebiten.BlendLighter, nil
	case "source-in":
		return ebiten.BlendSourceIn, nil
	case "source-atop":
		return ebiten.BlendSourceAtop, nil
	case "copy":
		return ebiten.BlendCopy, nil
	}
	return ebiten.Blend{}, fmt.Errorf("unknown blend %q", name)
}
func alignment(name string) (scrolling.Alignment, error) {
	switch name {
	case "", "left":
		return scrolling.AlignLeft, nil
	case "center":
		return scrolling.AlignCenter, nil
	case "right":
		return scrolling.AlignRight, nil
	}
	return 0, fmt.Errorf("unknown page alignment %q", name)
}
