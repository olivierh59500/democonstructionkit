// Command contourtrails composes retained contours with a live bitplane palette.
package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/render"
)

const width, height = 480, 300

type game struct {
	bank          *composite.ContourBank
	palette       *composite.BitplanePalette
	tick, current int
	colors        [8]color.NRGBA
}

func newGame() (*game, error) {
	g := &game{}
	var err error
	g.bank, err = composite.NewContourBank(composite.ContourBankConfig{Width: width, Height: height, Slots: 6, Layers: 1, FillRule: ebiten.FillRuleEvenOdd})
	if err != nil {
		return nil, err
	}
	for i := range g.colors {
		g.colors[i] = color.NRGBA{R: uint8(i * 35), G: uint8(220 - i*20), B: uint8(70 + i*24), A: 255}
	}
	g.colors[0] = color.NRGBA{R: 8, G: 16, B: 30, A: 255}
	g.palette, err = composite.NewBitplanePalette(composite.BitplanePaletteConfig{Width: width, Height: height, Planes: 3, Palette: g.colors[:]})
	if err != nil {
		g.Close()
		return nil, err
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	// Retain each 25 Hz source pose across two 50 Hz simulation updates.
	if g.tick%2 != 0 {
		return nil
	}
	g.current = (g.current + 1) % 6
	seconds := float64(g.tick) / 50
	return g.bank.Paint(g.current, 0, true, func(batch *render.Batch) {
		for object := 0; object < 3; object++ {
			x := 240 + 125*math.Sin(seconds*1.2+float64(object)*2.1)
			y := 150 + 70*math.Cos(seconds*1.7+float64(object))
			batch.Fan(10, func(i int) ebiten.Vertex {
				angle := float64(i)*math.Pi/5 + seconds
				radius := 40.0
				if i%2 == 1 {
					radius = 18
				}
				return render.Vertex(x+radius*math.Cos(angle), y+radius*math.Sin(angle), 0, 0, color.White)
			})
		}
	})
}

func (g *game) Draw(dst *ebiten.Image) {
	planes := [3]*ebiten.Image{g.bank.Image(g.current, 0), g.bank.Image((g.current+4)%6, 0), g.bank.Image((g.current+2)%6, 0)}
	if err := g.palette.Draw(dst, planes[:]); err != nil {
		panic(err)
	}
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	if g.palette != nil {
		g.palette.Close()
	}
	if g.bank != nil {
		g.bank.Close()
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
		err := capture.Run(capture.Config{Directory: *dir, Frames: []int{*frame}, Width: width, Height: height}, func() (ebiten.Game, error) { var err error; g, err = newGame(); return g, err })
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
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Retained Contour Trails")
	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
