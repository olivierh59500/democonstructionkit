// Command bitplanes shows how independent animated masks share one live palette.
package main

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/palette"
)

const width, height = 640, 360

type game struct {
	lookup *composite.BitplanePalette
	planes [4]*ebiten.Image
	white  *ebiten.Image
	colors [16]color.NRGBA
	tick   int
}

func newGame() (*game, error) {
	g := &game{white: ebiten.NewImage(1, 1)}
	g.white.Fill(color.White)
	for i := range g.planes {
		g.planes[i] = ebiten.NewImage(width, height)
	}
	g.colors[0] = color.NRGBA{} // Index zero reveals the background layer.
	for i := 1; i < len(g.colors); i++ {
		g.colors[i] = color.NRGBA{R: uint8(80 + i*10), G: uint8(24 + i*8), B: 220, A: 230}
	}
	var err error
	g.lookup, err = composite.NewBitplanePalette(composite.BitplanePaletteConfig{
		Width: width, Height: height, Planes: len(g.planes), Palette: g.colors[:],
	})
	if err != nil {
		g.Close()
		return nil, err
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	for i := 1; i < len(g.colors); i++ {
		hue := math.Mod(float64(g.tick)*0.002+float64(i)*0.065, 1)
		r, green, b := palette.HSLToRGB(hue, 0.8, 0.55)
		g.colors[i] = color.NRGBA{R: uint8(math.Round(r * 255)), G: uint8(math.Round(green * 255)), B: uint8(math.Round(b * 255)), A: 230}
	}
	return g.lookup.SetPalette(g.colors[:])
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 18, B: 32, A: 255})
	for bit, plane := range g.planes {
		plane.Clear()
		for repeat := 0; repeat < 3; repeat++ {
			phase := float64(g.tick)*float64(bit+1)*0.018 + float64(repeat)*2.094
			x := float64(width)/2 + math.Sin(phase)*float64(72+bit*32)
			y := float64(height)/2 + math.Cos(phase*0.75)*float64(35+bit*15)
			var op ebiten.DrawImageOptions
			op.GeoM.Scale(float64(105+bit*18), float64(35+bit*12))
			op.GeoM.Translate(x-float64(105+bit*18)/2, y-float64(35+bit*12)/2)
			plane.DrawImage(g.white, &op)
		}
	}
	if err := g.lookup.Draw(dst, g.planes[:]); err != nil {
		panic(err)
	}
}

func (*game) Layout(int, int) (int, int) { return width, height }

func (g *game) Close() {
	if g.lookup != nil {
		g.lookup.Close()
	}
	for _, plane := range g.planes {
		if plane != nil {
			plane.Deallocate()
		}
	}
	if g.white != nil {
		g.white.Deallocate()
	}
}

func main() {
	g, err := newGame()
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Bitplane Palette")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
