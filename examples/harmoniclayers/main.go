// Command harmoniclayers combines independently moving bands and sprite formations.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

const width, height = 640, 360

type game struct {
	bands    *composite.HarmonicBands
	balls    *sprites.HarmonicField
	image    *ebiten.Image
	tick     int
	outlined bool
}

func newGame(outlined bool) (*game, error) {
	g := &game{outlined: outlined}
	colors := make([]color.NRGBA, 12)
	for i := range colors {
		colors[i] = color.NRGBA{R: uint8(90 + i*12), G: uint8(190 - i*8), B: uint8(250 - i*4), A: 255}
	}
	var err error
	g.bands, err = composite.NewHarmonicBands(composite.HarmonicBandsConfig{
		LeftX: 0, RightX: width, Thickness: 8, Colors: colors,
		Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: height / 2, Y: height / 2},
			X: []motion.IndexedHarmonic{{Amplitude: 80, Rate: 1.1, IndexRate: .32}, {Amplitude: 70, Rate: .7, IndexRate: .53}},
			Y: []motion.IndexedHarmonic{{Amplitude: 75, Rate: .9, IndexRate: .43}, {Amplitude: 75, Rate: 1.4, IndexRate: .21}}},
	})
	if err != nil {
		return nil, err
	}
	bitmap := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			dx, dy := (float64(x)-11.5)/11.5, (float64(y)-11.5)/11.5
			if distance := dx*dx + dy*dy; distance <= 1 {
				light := max(0, -.4*dx-.5*dy+.75*math.Sqrt(1-distance))
				bitmap.SetNRGBA(x, y, color.NRGBA{R: uint8(50 + 180*light), G: uint8(60 + 180*light), B: 255, A: 255})
			}
		}
	}
	g.image = ebiten.NewImageFromImage(bitmap)
	g.balls, err = sprites.NewHarmonicField(sprites.HarmonicFieldConfig{
		Count: 64, Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: width / 2, Y: height / 2},
			X: []motion.IndexedHarmonic{{Amplitude: 155, Rate: 1.2, IndexRate: .15}, {Amplitude: 100, Rate: .7, IndexRate: .29}},
			Y: []motion.IndexedHarmonic{{Amplitude: 70, Rate: .9, IndexRate: .19}, {Amplitude: 75, Rate: 1.1, IndexRate: .31}}},
		Style: sprites.FieldStyle{Image: g.image, Appearance: sprites.FieldAppearance{AnchorX: .5, AnchorY: .5}},
	})
	if err != nil {
		g.Close()
		return nil, err
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.outlined = !g.outlined
	}
	frame := kit.Frame{Time: float64(g.tick) / 50}
	if err := g.bands.Update(frame); err != nil {
		return err
	}
	return g.balls.Update(frame)
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 15, B: 28, A: 255})
	if g.outlined {
		g.bands.DrawOutline(dst, 1.2)
	} else {
		g.bands.Draw(dst)
	}
	g.balls.Draw(dst)
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	if g.balls != nil {
		g.balls.Close()
	}
	if g.bands != nil {
		g.bands.Close()
	}
	if g.image != nil {
		g.image.Deallocate()
	}
}

func main() {
	dir := flag.String("capture", "", "write a deterministic native PNG")
	frame := flag.Int("frame", 200, "update tick to capture")
	outlined := flag.Bool("outline", false, "start with outlined bands; W toggles the material")
	flag.Parse()
	if *dir != "" {
		if *frame < 0 {
			log.Fatal("-frame must be nonnegative")
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *dir, Frames: []int{*frame}, Width: width, Height: height}, func() (ebiten.Game, error) {
			var err error
			g, err = newGame(*outlined)
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
	g, err := newGame(*outlined)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(50)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Harmonic Bands and Sprites")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
