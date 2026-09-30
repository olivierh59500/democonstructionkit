// Command palettefades compares reusable integer color passes on one artwork.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/olivierh59500/democonstructionkit/composite"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/render"
)

const width, height = 640, 360

type game struct {
	color  *composite.QuantizedColor
	art    *ebiten.Image
	panels [4]*ebiten.Image
	states [4]composite.QuantizedColorState
	tick   int
}

func newGame() (*game, error) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 320, 160))
	for y := 0; y < 160; y++ {
		for x := 0; x < 320; x++ {
			c := color.NRGBA{R: uint8(x * 255 / 319), G: uint8(y * 255 / 159), B: uint8((x + y) * 255 / 478), A: 255}
			if (x/24+y/16)%2 == 0 {
				c.B = 255 - c.B
			}
			pixels.SetNRGBA(x, y, c)
		}
	}
	g := &game{art: ebiten.NewImageFromImage(pixels)}
	var err error
	g.color, err = composite.NewQuantizedColor(composite.QuantizedColorConfig{})
	if err != nil {
		g.Close()
		return nil, err
	}
	for i := range g.panels {
		g.panels[i] = render.NewSurface(320, 160)
	}
	g.states = [4]composite.QuantizedColorState{
		{Mode: composite.QuantizedPassthrough},
		{Mode: composite.QuantizedKeep},
		{Mode: composite.QuantizedScale, Denominator: 16},
		{Mode: composite.QuantizedFromTarget, Target: [3]uint16{15, 15, 15}, Denominator: 32},
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	phase := g.tick % 128
	strength := min(phase, 128-phase)
	g.states[2].Numerator = uint32(strength / 4)
	g.states[3].Numerator = uint32(strength / 2)
	return nil
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 16, B: 24, A: 255})
	labels := [4]string{"Original", "RGB12 color grid", "16-step scaling", "32-step fade from white"}
	for i, panel := range g.panels {
		panel.Clear()
		if err := g.color.Draw(panel, g.art, g.states[i]); err != nil {
			panic(err)
		}
		x, y := (i%2)*320, (i/2)*180
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(panel, &op)
		ebitenutil.DebugPrintAt(dst, labels[i], x+4, y+163)
	}
}

func (*game) Layout(int, int) (int, int) { return width, height }

func (g *game) Close() {
	if g.color != nil {
		g.color.Close()
	}
	if g.art != nil {
		g.art.Deallocate()
	}
	for _, panel := range g.panels {
		if panel != nil {
			panel.Deallocate()
		}
	}
}

func main() {
	directory := flag.String("capture", "", "write a deterministic native PNG")
	frame := flag.Int("frame", 32, "update tick to capture")
	flag.Parse()
	if *directory != "" {
		if *frame < 0 {
			log.Fatal("-frame must be nonnegative")
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *directory, Frames: []int{*frame}, Width: width, Height: height},
			func() (ebiten.Game, error) { var err error; g, err = newGame(); return g, err })
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
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Palette Fades")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
