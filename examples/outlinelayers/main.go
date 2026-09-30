// Command outlinelayers compares filled and outlined skins of one live composition.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const width, height = 480, 320

type game struct {
	cube          *effects.MeshEffect
	logo          *effects.Warp
	field         *sprites.HarmonicField
	white, ball   *ebiten.Image
	letters       *ebiten.Image
	filled, lines *ebiten.Image
	border        sprites.FieldOutline
	tick          int
}

func newGame() (_ *game, err error) {
	g := &game{border: sprites.FieldOutline{Width: 1, Color: color.NRGBA{R: 130, G: 210, B: 255, A: 255}}}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	g.white = ebiten.NewImage(1, 1)
	g.white.Fill(color.White)
	g.filled, g.lines = ebiten.NewImage(width, height), ebiten.NewImage(width, height)
	bitmap := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if (x-8)*(x-8)+(y-8)*(y-8) < 60 {
				bitmap.SetNRGBA(x, y, color.NRGBA{R: uint8(80 + x*8), G: uint8(130 + y*7), B: 255, A: 255})
			}
		}
	}
	g.ball = ebiten.NewImageFromImage(bitmap)
	word := image.NewRGBA(image.Rect(0, 0, 77, 13))
	writer := font.Drawer{Dst: word, Src: image.NewUniform(color.White), Face: basicfont.Face7x13, Dot: fixed.P(0, 12)}
	writer.DrawString("DCK EFFECTS")
	g.letters = ebiten.NewImageFromImage(word)
	g.logo, err = effects.NewWarp(kit.Func{OnDraw: func(dst *ebiten.Image) {
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(192.0/77, 40.0/13)
		dst.DrawImage(g.letters, &op)
	}}, 192, 40, 8, 4)
	if err != nil {
		return nil, err
	}
	g.logo.Map = func(x, y, seconds float64) geometry.Vec2 {
		return geometry.Vec2{X: 144 + x + 12*math.Sin(seconds+y*.08), Y: 26 + y + 4*math.Cos(seconds+x*.015)}
	}
	if err = g.logo.SetOutline(effects.WarpOutlineConfig{Width: 1, White: g.white,
		Color: color.NRGBA{R: 220, G: 150, B: 255, A: 255}}); err != nil {
		return nil, err
	}
	g.cube, err = effects.NewMesh(effects.Cube(70, geometry.Vec2{X: 1, Y: 1}, color.NRGBA{R: 255, G: 160, B: 210, A: 255}),
		nil, geometry.Camera{Center: geometry.Vec2{X: width / 2, Y: 170}, Focal: 250, Near: 1})
	if err != nil {
		return nil, err
	}
	g.cube.CullBackFaces = true
	g.cube.Light = geometry.Vec3{X: -.5, Y: -.8, Z: -1}
	g.cube.Ambient = .35
	g.cube.Animate = func(seconds float64) effects.Transform {
		return effects.Transform{Position: geometry.Vec3{Z: 220}, Rotation: geometry.Vec3{X: seconds * .6, Y: seconds * .9}, Scale: 1}
	}
	if err = g.cube.SetOutline(effects.MeshOutlineConfig{Faces: effects.CubeFaces(), Width: 1.4, White: g.white,
		Color: color.NRGBA{R: 255, G: 170, B: 220, A: 255}}); err != nil {
		return nil, err
	}
	g.field, err = sprites.NewHarmonicField(sprites.HarmonicFieldConfig{
		Count: 40, Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: width / 2, Y: 220},
			X: []motion.IndexedHarmonic{{Amplitude: 150, Rate: 1.1, IndexRate: .32}},
			Y: []motion.IndexedHarmonic{{Amplitude: 55, Rate: .9, IndexRate: .43}}},
		Style: sprites.FieldStyle{Image: g.ball, Appearance: sprites.FieldAppearance{AnchorX: .5, AnchorY: .5}},
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	frame := kit.Frame{Time: float64(g.tick) / 50}
	if err := g.cube.Update(frame); err != nil {
		return err
	}
	if err := g.logo.Update(frame); err != nil {
		return err
	}
	return g.field.Update(frame)
}

func (g *game) Draw(dst *ebiten.Image) {
	background := color.NRGBA{R: 8, G: 15, B: 28, A: 255}
	g.filled.Fill(background)
	g.lines.Fill(background)
	g.field.Draw(g.filled)
	g.cube.Draw(g.filled)
	g.logo.Draw(g.filled)
	style := g.field.Style
	style.Outline = &g.border
	g.field.DrawStyle(g.lines, style)
	g.cube.DrawOutline(g.lines)
	g.logo.DrawOutline(g.lines)
	dst.DrawImage(g.filled, nil)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(width, 0)
	dst.DrawImage(g.lines, &op)
	ebitenutil.DebugPrintAt(dst, "FILLED", 8, 8)
	ebitenutil.DebugPrintAt(dst, "OUTLINE: SAME CLOCKS AND POSES", width+8, 8)
}
func (*game) Layout(int, int) (int, int) { return width * 2, height }
func (g *game) Close() {
	if g.field != nil {
		g.field.Close()
	}
	if g.cube != nil {
		g.cube.Close()
	}
	if g.logo != nil {
		g.logo.Close()
	}
	for _, img := range []*ebiten.Image{g.white, g.ball, g.letters, g.filled, g.lines} {
		if img != nil {
			img.Deallocate()
		}
	}
}

func main() {
	dir := flag.String("capture", "", "write a deterministic native PNG")
	frame := flag.Int("frame", 200, "update tick to capture")
	flag.Parse()
	if *dir != "" {
		if *frame < 0 {
			log.Fatal("-frame must be nonnegative")
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *dir, Frames: []int{*frame}, Width: width * 2, Height: height}, func() (ebiten.Game, error) {
			var err error
			g, err = newGame()
			return g, err
		})
		if g != nil {
			g.Close()
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	g, err := newGame()
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(50)
	ebiten.SetWindowSize(width*2, height)
	ebiten.SetWindowTitle("DCK Filled and Outlined Layers")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
