package recipes

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/assets"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/outline"
	"github.com/olivierh59500/democonstructionkit/presets"
)

const Width, Height = 640, 400

// Build creates an independent effect tree borrowing textures from store.
// Recipe scripts use a common canvas; original logical resolutions are in the audit.
func Build(name string, store *assets.Store) (kit.Effect, error) {
	if store == nil {
		return nil, fmt.Errorf("recipes: nil asset store")
	}
	if _, ok := Find(name); !ok {
		return nil, fmt.Errorf("recipes: unknown recipe %q", name)
	}
	b := &builder{store: store}
	group := kit.Group{&effects.Solid{Color: color.NRGBA{6, 7, 18, 255}}}
	add := func(e kit.Effect) {
		if e != nil {
			group = append(group, e)
		}
	}
	switch name {
	case "bilizir-demo":
		add(b.bars())
		add(b.logo("bilizir-demo/assets/logo.png", 320, 70, .6))
		add(b.cubes(12, false, false, nil))
		add(b.scroll(name, 300, 1.4, "strip"))
	case "dma-3d":
		add(b.stars())
		add(b.cubes(1, true, false, nil))
		add(b.scroll(name, 300, 1, "strip"))
		add(b.logo("dma-3d/assets/bandeau.png", 320, 32, .65))
	case "dma-is-back":
		add(b.tiles("dma-is-back/assets/small-dma-jelly.png", false))
		add(b.cubes(1, false, true, nil))
		add(b.scroll(name, 285, 1, "strip"))
	case "go-cocoisthebest":
		add(b.tiles("go-cocoisthebest/assets/coco.png", true))
		add(b.bars())
		add(b.cubes(8, false, false, nil))
		add(b.scroll(name, 300, 1, "twist"))
		add(b.logo("go-cocoisthebest/assets/dma-70.png", 320, 60, 1))
	case "viva_tcb":
		add(b.tiles("viva_tcb/assets/tcb_tile.png", true))
		add(b.logo("viva_tcb/assets/logo.png", 320, 70, .55))
		add(b.scroll(name, 180, .8, "sine"))
		add(b.scroll(name, 300, .9, "sine"))
	case "megatwist":
		add(b.tiles("megatwist/assets/back.png", false))
		add(b.logo("megatwist/assets/logo.png", 320, 70, .7))
		add(b.scroll(name, 210, 1.2, "twist"))
	case "phenomena-dna-scroll-intro":
		add(b.logo("phenomena-dna-scroll-intro/assets/logo.png", 320, 65, 1))
		add(b.bars())
		add(b.scroll(name, 210, 2, "dna"))
		add(b.logo("phenomena-dna-scroll-intro/assets/photon.png", 320, 355, 1))
	case "tcb-multi-plane-3d-scroller":
		add(b.logo("tcb-multi-plane-3d-scroller/assets/mountains.png", 320, 280, 1.5))
		for i := 0; i < 3; i++ {
			add(b.scroll(name, float64(110+i*85), 1, "plane"))
		}
		add(b.logo("tcb-multi-plane-3d-scroller/assets/logo.png", 320, 40, .7))
	case "tcb-replicants-demo":
		add(b.stars())
		add(b.logo("tcb-replicants-demo/assets/tcb_logo.png", 200, 120, .65))
		add(b.logo("tcb-replicants-demo/assets/rep_logo.png", 440, 160, .65))
		add(b.scroll(name, 300, 1, "strip"))
	case "teamg1-demo":
		add(b.plasma())
		texture := b.texture("teamg1-demo/assets/texture.png")
		add(b.cubes(1, false, false, texture))
		for i := 0; i < 6; i++ {
			add(b.logo("teamg1-demo/assets/gameone_logo.png", 320+float64(i-3)*70, 70, .6))
		}
		add(b.scroll(name, 315, 1, "strip"))
	case "go-dom-intro":
		add(b.stars())
		add(b.bars())
		for i := 0; i < 4; i++ {
			add(b.scroll(name, float64(50+i*90), .5+float64(i)*.35, "sine"))
		}
		add(b.logo("go-dom-intro/assets/rep_ik+_logo.png", 320, 30, .45))
	case "nonameno-demo":
		add(b.stars())
		add(b.letters())
		add(b.scroll("nonameno-small", 345, 2, "sine"))
	case "grodan-kvack-kvack-demo":
		add(b.tiles("grodan-kvack-kvack-demo/assets/Grodan_green.png", false))
		add(b.spriteTrain())
		add(b.scroll(name, 120, 1.5, "strip"))
		add(b.scroll("grodan-up", 0, .8, "vertical"))
		add(b.scroll("grodan-small", 360, 2, "sine"))
	case "go-vectorballs":
		add(b.vectorCaption())
		add(b.balls(true))
	case "3d_doc":
		add(b.logo("3d_doc/assets/mountains.png", 320, 130, 1))
		add(b.checkerboard())
		add(b.balls(false))
		add(b.scroll(name, 325, .8, "strip"))
	case "go-cuddlymenu":
		add(b.tileWorld())
		add(b.dude())
		add(b.scroll(name, 310, .8, "sine"))
	case "go-fr010":
		add(b.plasma())
		add(b.cubes(3, false, false, nil))
		add(b.vectorText())
	case "go-secondreality":
		add(b.secondReality())
	case "go-multiscreen":
		for i, childName := range []string{"go-cocoisthebest", "viva_tcb", "phenomena-dna-scroll-intro", "tcb-multi-plane-3d-scroller"} {
			child, err := Build(childName, store)
			if err != nil {
				b.fail(err)
				break
			}
			v, err := kit.NewViewport(child, Width, Height, image.Rect(i%2*320, i/2*200, (i%2+1)*320, (i/2+1)*200))
			if err != nil {
				kit.Close(child)
				b.fail(err)
				break
			}
			add(v)
		}
	}
	if b.err != nil {
		_ = group.Close()
		return nil, fmt.Errorf("recipe %s: %w", name, b.err)
	}
	if name == "dma-is-back" {
		crt, err := effects.NewCRT(group, Width, Height, effects.CRTConfig{Curvature: .025, Scanlines: .15, Chromatic: .7, Vignette: .2})
		if err != nil {
			group.Close()
			return nil, err
		}
		return crt, nil
	}
	return group, nil
}

type builder struct {
	store *assets.Store
	err   error
}

func (b *builder) fail(err error) {
	if err != nil && b.err == nil {
		b.err = err
	}
}
func (b *builder) texture(path string) *ebiten.Image {
	if b.err != nil {
		return nil
	}
	img, err := b.store.Texture(path)
	b.fail(err)
	return img
}
func (b *builder) font(id string) (*ebiten.Image, *font.Font) {
	spec, ok := presets.FindFont(id)
	if !ok {
		b.fail(fmt.Errorf("unknown font %s", id))
		return nil, nil
	}
	atlas := b.texture(spec.Path)
	if atlas == nil {
		return nil, nil
	}
	f, err := spec.Build(atlas.Bounds())
	b.fail(err)
	return atlas, f
}
func (b *builder) stars() kit.Effect {
	e, err := effects.NewStarfield(Width, Height, 280, 42)
	b.fail(err)
	if err != nil {
		return nil
	}
	return e
}
func (b *builder) bars() kit.Effect {
	var bars []effects.RasterBar
	for i := 0; i < 6; i++ {
		var colors []color.NRGBA
		for j := 0; j < 16; j++ {
			v := uint8(255 * math.Sin(math.Pi*float64(j)/16))
			colors = append(colors, color.NRGBA{R: v, G: uint8(int(v) * i / 6), B: 255 - v/2, A: 170})
		}
		bars = append(bars, effects.RasterBar{Y: 160, Height: 32, Colors: colors, Motion: motion.Waves{{Amplitude: 130, Speed: 1.1, Phase: float64(i) * .6}, {Amplitude: 20, Speed: 2.3, Phase: float64(i)}}})
	}
	return effects.NewRasterBars(bars)
}
func (b *builder) logo(path string, x, y, scale float64) kit.Effect {
	texture := b.texture(path)
	if texture == nil {
		return nil
	}
	return &effects.Sprite{Frames: []*ebiten.Image{texture}, Path: func(t float64) effects.Pose {
		p := effects.IdentityPose(x+25*math.Sin(t+x*.01), y+8*math.Cos(t*1.3))
		p.ScaleX = scale
		p.ScaleY = scale
		return p
	}}
}
func (b *builder) tiles(path string, rotate bool) kit.Effect {
	texture := b.texture(path)
	if texture == nil {
		return nil
	}
	return &effects.Tiles{Texture: texture, Animate: func(t float64) effects.TileTransform {
		angle := 0.0
		scale := 1.0
		if rotate {
			angle = t * .2
			scale = .75 + .3*math.Sin(t*.4)
		}
		return effects.TileTransform{Center: geometry.Vec2{X: 320, Y: 200}, Scale: scale, Angle: angle, PhaseX: t * 25, PhaseY: 20 * math.Sin(t)}
	}}
}

func (b *builder) cubes(count int, glenz, jelly bool, texture *ebiten.Image) kit.Effect {
	group := kit.Group{}
	size := 60.0
	if count == 1 {
		size = 140
	}
	uv := geometry.Vec2{X: 1, Y: 1}
	if texture != nil {
		uv = geometry.Vec2{X: float64(texture.Bounds().Dx()), Y: float64(texture.Bounds().Dy())}
	}
	for i := 0; i < count; i++ {
		c := color.NRGBA{255, uint8(80 + i*8), 180, 255}
		if glenz {
			c.A = 95
		}
		mesh, err := effects.NewMesh(effects.Cube(size, uv, c), texture, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 190}, Focal: 360, Near: 20})
		if err != nil {
			b.fail(err)
			break
		}
		mesh.CullBackFaces = !glenz
		mesh.Light = geometry.Vec3{X: -.5, Y: -.5, Z: -1}
		mesh.Ambient = .35
		mesh.Animate = func(t float64) effects.Transform {
			x, y := 0.0, 0.0
			if count > 1 {
				x = 220 * math.Sin(t+float64(i)*.55)
				y = 80 * math.Cos(t*1.3+float64(i))
			}
			return effects.Transform{Position: geometry.Vec3{X: x, Y: y, Z: 400}, Rotation: geometry.Vec3{X: t * (.7 + float64(i)*.03), Y: t, Z: t * .3}, Scale: 1}
		}
		if jelly {
			mesh.Deform = func(i int, p geometry.Vec3, t float64) geometry.Vec3 {
				p.X += 25 * math.Sin(p.Y*.03+t*2)
				p.Y += 15 * math.Cos(p.X*.03+t*3)
				return p
			}
		}
		group = append(group, mesh)
	}
	return group
}

func (b *builder) scroll(id string, y, scale float64, mode string) kit.Effect {
	atlas, f := b.font(id)
	if atlas == nil || f == nil {
		return nil
	}
	message := "  DEMOCONSTRUCTIONKIT  -  " + id + "  -  SHARED EFFECTS IN GO!  "
	c := effects.ScrollConfig{Width: Width, Height: Height, Message: message, Speed: 90, Gap: 80, Scale: scale, Y: y}
	if mode == "vertical" {
		c.Vertical = true
		c.Message = "GRODAN\nKVACK\nKVACK\n\nGREETINGS\nTO ALL\nDEMO\nCODERS"
		c.X = 400
		c.Speed = 35
		c.Gap = 100
	}
	if mode == "sine" {
		c.Wave = motion.Waves{{Amplitude: 22, Spatial: .017, Speed: 2}}
	}
	warped := mode == "strip" || mode == "twist" || mode == "dna" || mode == "plane"
	if warped {
		c.Y = 0
		c.Width = Width + 128
		c.Height = int(math.Ceil(f.LineHeight() * scale))
	}
	s, err := effects.NewScroller(atlas, f, c)
	if err != nil {
		b.fail(err)
		return nil
	}
	if !warped {
		return s
	}
	cols, rows := 1, c.Height
	if mode == "dna" {
		cols, rows = c.Width/2, 1
	}
	if mode == "plane" {
		cols, rows = 64, 4
	}
	w, err := effects.NewWarp(s, c.Width, c.Height, cols, rows)
	if err != nil {
		b.fail(err)
		return nil
	}
	h := float64(c.Height)
	w.Map = func(x, v, t float64) geometry.Vec2 {
		switch mode {
		case "dna":
			angle := x*.018 - t*2
			return geometry.Vec2{X: x - 64, Y: y + 70*math.Sin(angle) + (v-h/2)*math.Cos(angle)}
		case "plane":
			angle := t*.6 + y*.01
			z := 380 + (v-h/2)*math.Sin(angle)
			factor := 380 / z
			return geometry.Vec2{X: 320 + (x-float64(c.Width)/2)*factor, Y: y + (v-h/2)*math.Cos(angle)*factor}
		case "twist":
			return geometry.Vec2{X: x - 64 + 40*math.Sin(v*.08+t*1.4) + 16*math.Cos(v*.3-t*2), Y: y + v + 30*math.Sin(t*1.7)}
		default:
			return geometry.Vec2{X: x - 64 + 20*math.Sin(v*.1+t*2) + 12*math.Cos(v*.3-t), Y: y + v + 16*math.Cos(x*.015+t*2)}
		}
	}
	if mode == "dna" {
		w.Depth = func(x, y, t float64) float64 { return math.Cos(x*.018 - t*2) }
		w.Tint = func(x, y, t float64) color.Color {
			if math.Cos(x*.018-t*2) < 0 {
				return color.NRGBA{R: 255, G: 100, B: 180, A: 255}
			}
			return color.White
		}
	}
	return w
}

func palette() []color.NRGBA {
	p := make([]color.NRGBA, 256)
	for i := range p {
		x := float64(i) * 2 * math.Pi / 256
		p[i] = color.NRGBA{R: uint8(128 + 127*math.Sin(x)), G: uint8(128 + 127*math.Sin(x+2)), B: uint8(128 + 127*math.Sin(x+4)), A: 255}
	}
	return p
}
func (b *builder) plasma() kit.Effect {
	p, err := effects.NewPixels(160, 100, effects.Plasma(palette(), .08, 1))
	if err != nil {
		b.fail(err)
		return nil
	}
	p.Rect = image.Rect(0, 0, Width, Height)
	return p
}

func (b *builder) balls(mirror bool) kit.Effect {
	path := "3d_doc/assets/ball.png"
	if mirror {
		path = "go-vectorballs/assets/AllBalls.png"
	}
	texture := b.texture(path)
	if texture == nil {
		return nil
	}
	if mirror {
		texture = texture.SubImage(image.Rect(0, 12, 16, 28)).(*ebiten.Image)
	}
	from, to := make([]geometry.Vec3, 96), make([]geometry.Vec3, 96)
	for i := range from {
		a := float64(i) * 2 * math.Pi / float64(len(from))
		from[i] = geometry.Vec3{X: 125 * math.Cos(a), Y: 90 * math.Sin(a), Z: 30 * math.Sin(3*a)}
		to[i] = geometry.Vec3{X: float64(i%8-4) * 28, Y: float64(i/8-6) * 19, Z: 40 * math.Sin(a*4)}
	}
	p, err := effects.NewPointCloud(from, texture, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 175}, Focal: 360, Near: 10}, 18)
	if err != nil {
		b.fail(err)
		return nil
	}
	p.Shape = func(dst []geometry.Vec3, t float64) { geometry.Morph(dst, from, to, (1+math.Sin(t*.45))/2) }
	p.Animate = func(t float64) effects.Transform {
		return effects.Transform{Position: geometry.Vec3{Z: 420}, Rotation: geometry.Vec3{X: t * .2, Y: t * .45, Z: t * .1}, Scale: 1}
	}
	if mirror {
		r, err := effects.NewReflection(p, Width, Height, 270)
		if err != nil {
			b.fail(err)
			return nil
		}
		r.Scale = .6
		return r
	}
	return p
}

func (b *builder) vectorCaption() kit.Effect {
	t := b.texture("go-vectorballs/assets/text.png")
	if t == nil {
		return nil
	}
	row := t.SubImage(image.Rect(0, 0, 640, 14)).(*ebiten.Image)
	return &effects.Sprite{Frames: []*ebiten.Image{row}, Pose: effects.IdentityPose(320, 370)}
}

func (b *builder) letters() kit.Effect {
	atlas, f := b.font("nonameno-demo")
	if atlas == nil || f == nil {
		return nil
	}
	text, err := effects.NewText(atlas, f, "NO NAME\nDEMO CONSTRUCTION\nKIT", 2)
	if err != nil {
		b.fail(err)
		return nil
	}
	text.GlyphPose = func(i int, g font.Placement, t float64) effects.Pose {
		t = motion.Wrap(t, 10)
		progress := motion.ElasticOut(motion.Linear((t - float64(i)*.04) / 2))
		p := effects.IdentityPose(70+g.X, 90+g.Y+(1-progress)*350)
		p.ScaleX = progress
		p.ScaleY = progress
		p.Angle = (1 - progress) * 2
		return p
	}
	return text
}
func (b *builder) checkerboard() kit.Effect {
	m := effects.Mesh{}
	for z := 0; z < 10; z++ {
		for x := -8; x < 8; x++ {
			i := len(m.Points)
			xx, zz := float64(x*60), float64(z*60)
			m.Points = append(m.Points, geometry.Vec3{X: xx, Y: 110, Z: zz}, geometry.Vec3{X: xx + 60, Y: 110, Z: zz}, geometry.Vec3{X: xx + 60, Y: 110, Z: zz + 60}, geometry.Vec3{X: xx, Y: 110, Z: zz + 60})
			c := color.NRGBA{30, 30, 60, 255}
			if (x+z)%2 == 0 {
				c = color.NRGBA{150, 100, 220, 255}
			}
			m.Triangles = append(m.Triangles, effects.Triangle{Indices: [3]int{i, i + 1, i + 2}, Color: c}, effects.Triangle{Indices: [3]int{i, i + 2, i + 3}, Color: c})
		}
	}
	e, err := effects.NewMesh(m, nil, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 140}, Focal: 300, Near: 20})
	if err != nil {
		b.fail(err)
		return nil
	}
	e.Animate = func(t float64) effects.Transform {
		return effects.Transform{Position: geometry.Vec3{X: 30 * math.Sin(t), Z: 80 - motion.Wrap(t*50, 60)}, Scale: 1}
	}
	return e
}

func (b *builder) spriteTrain() kit.Effect {
	t := b.texture("grodan-kvack-kvack-demo/assets/sprite.png")
	if t == nil {
		return nil
	}
	frame := t.SubImage(image.Rect(0, 0, 16, 10)).(*ebiten.Image)
	group := kit.Group{}
	for i := 0; i < 12; i++ {
		group = append(group, &effects.Sprite{Frames: []*ebiten.Image{frame}, Path: func(t float64) effects.Pose {
			a := t*1.4 + float64(i)*.25
			p := effects.IdentityPose(320+200*math.Sin(a), 80+50*math.Cos(a*1.3))
			p.ScaleX = 2
			p.ScaleY = 2
			return p
		}})
	}
	return group
}
func (b *builder) tileWorld() kit.Effect {
	atlas := b.texture("go-cuddlymenu/assets/menu/tiles.png")
	if atlas == nil {
		return nil
	}
	var rects []image.Rectangle
	for y := 0; y+32 <= atlas.Bounds().Dy(); y += 32 {
		for x := 0; x+32 <= atlas.Bounds().Dx(); x += 32 {
			rects = append(rects, image.Rect(x, y, x+32, y+32))
		}
	}
	cells := make([]int, 80*9)
	for i := range cells {
		if i/80 < 7 {
			cells[i] = -1
		} else {
			cells[i] = 2
		}
	}
	for i := 0; i < 80; i += 7 {
		cells[5*80+i] = min(60, len(rects)-1)
	}
	m, err := effects.NewTilemap(atlas, rects, cells, 80, 9, image.Pt(32, 32))
	if err != nil {
		b.fail(err)
		return nil
	}
	m.Camera = func(t float64) geometry.Vec2 { return geometry.Vec2{X: motion.Wrap(t*45, 1500)} }
	return m
}
func (b *builder) dude() kit.Effect {
	atlas := b.texture("go-cuddlymenu/assets/menu/dude.png")
	if atlas == nil {
		return nil
	}
	var rects []image.Rectangle
	for i := 0; i < 10; i++ {
		rects = append(rects, image.Rect(i*64, 0, (i+1)*64, 64))
	}
	frames, err := effects.SplitSheet(atlas, rects)
	if err != nil {
		b.fail(err)
		return nil
	}
	return &effects.Sprite{Frames: frames, FPS: 9, Path: func(t float64) effects.Pose { return effects.IdentityPose(320, 185-30*math.Abs(math.Sin(t*2))) }}
}
func (b *builder) vectorText() kit.Effect {
	data, err := fs.ReadFile(b.store.Files, "go-fr010/data/secrcode.ttf")
	if err != nil {
		b.fail(err)
		return nil
	}
	f, err := outline.New(data, 50, 8)
	if err != nil {
		b.fail(err)
		return nil
	}
	points, edges, err := f.Text("FR-010\nGO CONSTRUCTION KIT", 1)
	if err != nil {
		b.fail(err)
		return nil
	}
	for i := range points {
		points[i].X -= 230
		points[i].Y -= 20
	}
	w, err := effects.NewWireframe(points, edges, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 190}, Focal: 400, Near: 20})
	if err != nil {
		b.fail(err)
		return nil
	}
	w.Color = color.NRGBA{255, 240, 190, 255}
	w.Width = 1.5
	track, err := motion.NewTrack([]motion.Key[geometry.Vec3]{{Time: 0, Value: geometry.Vec3{Z: 650}, Ease: motion.Sine}, {Time: 5, Value: geometry.Vec3{Z: 400}, Ease: motion.Sine}, {Time: 10, Value: geometry.Vec3{Z: 650}}}, geometry.Lerp)
	if err != nil {
		b.fail(err)
		w.Close()
		return nil
	}
	w.Animate = func(t float64) effects.Transform {
		return effects.Transform{Position: track.At(motion.Wrap(t, 10)), Rotation: geometry.Vec3{Y: .45 * math.Sin(t*.5)}, Scale: 1}
	}
	return w
}
func (b *builder) secondReality() kit.Effect {
	texture := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			texture.Set(x, y, color.NRGBA{uint8(x * 4), uint8(y * 4), uint8((x ^ y) * 4), 255})
		}
	}
	tunnel, err := effects.NewPixels(160, 100, effects.Tunnel(texture, 160, 100, 25, .08))
	if err != nil {
		b.fail(err)
		return nil
	}
	tunnel.Rect = image.Rect(0, 0, Width, Height)
	scenes := []kit.Effect{kit.Group{b.stars(), b.cubes(1, true, false, nil)}, tunnel, b.plasma(), b.balls(true), b.checkerboard()}
	if b.err != nil {
		kit.Group(scenes).Close()
		return nil
	}
	seq, err := kit.NewSequence(scenes, []time.Duration{6 * time.Second, 6 * time.Second, 6 * time.Second, 6 * time.Second, 6 * time.Second}, true)
	if err != nil {
		b.fail(err)
		kit.Group(scenes).Close()
		return nil
	}
	return seq
}
