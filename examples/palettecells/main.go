// Command palettecells combines a shadowed palette grid and a moving row material.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/olivierh59500/democonstructionkit/composite"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/palette"
	"github.com/olivierh59500/democonstructionkit/render"
)

const width, height = 400, 280

type game struct {
	bank          *composite.ContourBank
	grid, rows    *composite.PaletteGrid
	material      *ebiten.Image
	left, right   *ebiten.Image
	first, second []color.NRGBA
	rowA, rowB    []color.NRGBA
	tick          int
}

func newGame() (_ *game, err error) {
	g := &game{first: make([]color.NRGBA, 10*14), second: make([]color.NRGBA, 10*14),
		rowA: make([]color.NRGBA, height), rowB: make([]color.NRGBA, height)}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	g.bank, err = composite.NewContourBank(composite.ContourBankConfig{Width: width, Height: height, Slots: 1, Layers: 1, FillRule: ebiten.FillRuleEvenOdd})
	if err != nil {
		return nil, err
	}
	bodyColor := color.NRGBA{R: 10, G: 18, B: 30, A: 255}
	g.grid, err = composite.NewPaletteGrid(composite.PaletteGridConfig{Width: width, Height: height, Columns: 10, Rows: 14,
		X: composite.PaletteGridAxis{CellSize: 40}, Y: composite.PaletteGridAxis{CellSize: 20},
		ControlThreshold: .5, BodyColor: &bodyColor})
	if err != nil {
		return nil, err
	}
	g.rows, err = composite.NewPaletteGrid(composite.PaletteGridConfig{Width: width, Height: height,
		Columns: 1, Rows: height, ControlChannel: composite.BitplaneRed})
	if err != nil {
		return nil, err
	}
	bitmap := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := uint8(math.Round((.5 + .5*math.Sin(float64(x)*.05+float64(y)*.04)) * 255))
			bitmap.SetNRGBA(x, y, color.NRGBA{R: value, A: 255})
		}
	}
	g.material = ebiten.NewImageFromImage(bitmap)
	g.left, g.right = ebiten.NewImage(width, height), ebiten.NewImage(width, height)
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	seconds := float64(g.tick) / 50
	for row := 0; row < 14; row++ {
		for col := 0; col < 10; col++ {
			r, green, b := palette.HSLToRGB(math.Mod(float64(row+col)*.04+seconds*.03, 1), .7, .65)
			c := color.NRGBA{R: uint8(r * 255), G: uint8(green * 255), B: uint8(b * 255), A: 255}
			g.first[row*10+col] = c
			g.second[row*10+col] = color.NRGBA{R: c.R / 3, G: c.G / 3, B: c.B / 3, A: 255}
		}
	}
	for y := 0; y < height; y++ {
		wave := .5 + .5*math.Sin(seconds+float64(y)*.035)
		g.rowA[y] = color.NRGBA{R: uint8(40 + wave*110), G: 30, B: 170, A: 255}
		g.rowB[y] = color.NRGBA{R: 240, G: uint8(80 + wave*160), B: 80, A: 255}
	}
	if err := g.grid.SetColors(g.first, g.second); err != nil {
		return err
	}
	if err := g.rows.SetColors(g.rowA, g.rowB); err != nil {
		return err
	}
	return g.bank.Paint(0, 0, true, func(batch *render.Batch) {
		batch.Fan(10, func(i int) ebiten.Vertex {
			angle, radius := float64(i)*math.Pi/5+seconds*.5, 70.0
			if i%2 == 1 {
				radius = 32
			}
			return render.Vertex(200+70*math.Sin(seconds*.8)+radius*math.Cos(angle), 140+45*math.Cos(seconds)+radius*math.Sin(angle), 0, 0, color.White)
		})
	})
}

func (g *game) Draw(dst *ebiten.Image) {
	mask := g.bank.Image(0, 0)
	if err := g.grid.Draw(g.left, mask, mask, composite.PaletteGridState{ControlOffset: [2]float32{-12, -8}}); err != nil {
		panic(err)
	}
	if err := g.rows.Draw(g.right, g.material, nil, composite.PaletteGridState{ControlOffset: [2]float32{float32(18 * math.Sin(float64(g.tick)*.02))}}); err != nil {
		panic(err)
	}
	dst.DrawImage(g.left, nil)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(width, 0)
	dst.DrawImage(g.right, &op)
	ebitenutil.DebugPrintAt(dst, "PALETTE GRID + BODY + OFFSET SHADOW", 8, 8)
	ebitenutil.DebugPrintAt(dst, "ROW COLORS + CONTINUOUS MATERIAL", width+8, 8)
}
func (*game) Layout(int, int) (int, int) { return width * 2, height }
func (g *game) Close() {
	if g.grid != nil {
		g.grid.Close()
	}
	if g.rows != nil {
		g.rows.Close()
	}
	if g.bank != nil {
		g.bank.Close()
	}
	for _, img := range []*ebiten.Image{g.material, g.left, g.right} {
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
	ebiten.SetWindowTitle("DCK Spatial Palette Materials")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
