// Command offsetmaterials maps independently moving textures through a live contour.
package main

import (
	"flag"
	"image"
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
	bank     *composite.ContourBank
	lookup   *composite.BitplanePalette
	material *ebiten.Image
	planes   [3]*ebiten.Image
	offsets  [3][2]float32
	tick     int
}

func newGame() (*game, error) {
	g := &game{}
	var err error
	g.bank, err = composite.NewContourBank(composite.ContourBankConfig{
		Width: width, Height: height, Slots: 1, Layers: 1, FillRule: ebiten.FillRuleEvenOdd,
	})
	if err != nil {
		return nil, err
	}
	// One immutable binary material supplies both independently sampled planes.
	bitmap := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			distance := math.Hypot(float64(x-width/2), float64(y-height/2))
			value := uint8(0)
			if math.Sin(distance*0.16)+math.Cos(float64(x-y)*0.07) > 0 {
				value = 255
			}
			bitmap.SetNRGBA(x, y, color.NRGBA{R: value, A: 255})
		}
	}
	g.material = ebiten.NewImageFromImage(bitmap)
	g.planes = [3]*ebiten.Image{g.bank.Image(0, 0), g.material, g.material}
	colors := [8]color.NRGBA{
		1: {R: 255, G: 98, B: 30, A: 255}, 3: {R: 190, G: 60, B: 240, A: 255},
		5: {R: 30, G: 200, B: 255, A: 255}, 7: {R: 255, G: 245, B: 160, A: 255},
	}
	g.lookup, err = composite.NewBitplanePalette(composite.BitplanePaletteConfig{
		Width: width, Height: height, Planes: 3, Palette: colors[:],
		Channels: []composite.BitplaneChannel{composite.BitplaneAlpha, composite.BitplaneRed, composite.BitplaneRed},
	})
	if err != nil {
		g.Close()
		return nil, err
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	seconds := float64(g.tick) / 50
	g.offsets[1] = [2]float32{float32(35 * math.Sin(seconds)), float32(24 * math.Cos(seconds*1.3))}
	g.offsets[2] = [2]float32{float32(28 * math.Cos(seconds*1.7)), float32(30 * math.Sin(seconds*0.8))}
	return g.bank.Paint(0, 0, true, func(batch *render.Batch) {
		batch.Fan(10, func(i int) ebiten.Vertex {
			angle, radius := float64(i)*math.Pi/5+seconds*0.4, 125.0
			if i%2 == 1 {
				radius = 65
			}
			return render.Vertex(width/2+radius*math.Cos(angle), height/2+radius*math.Sin(angle), 0, 0, color.White)
		})
		// An inner contour demonstrates the even-odd hole and transparent palette.
		batch.Fan(32, func(i int) ebiten.Vertex {
			angle := float64(i) * math.Pi / 16
			return render.Vertex(width/2+25*math.Cos(angle), height/2+25*math.Sin(angle), 0, 0, color.White)
		})
	})
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 18, B: 32, A: 255})
	if err := g.lookup.DrawOffsets(dst, g.planes[:], g.offsets[:]); err != nil {
		panic(err)
	}
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	if g.lookup != nil {
		g.lookup.Close()
	}
	if g.bank != nil {
		g.bank.Close()
	}
	if g.material != nil {
		g.material.Deallocate()
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
		err := capture.Run(capture.Config{Directory: *dir, Frames: []int{*frame}, Width: width, Height: height}, func() (ebiten.Game, error) {
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
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Independent Material Offsets")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
