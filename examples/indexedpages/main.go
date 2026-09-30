// Command indexedpages shows packed palette fades over one shared indexed image.
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
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/palette"
)

const (
	width, height = 800, 320
	fps           = 50
	cycleTicks    = 300
	fadeTicks     = 101 // Inclusive endpoints: tick 0 through tick 100, two seconds.
)

type game struct {
	art        *ebiten.Image
	pages      [2]*effects.IndexedImage
	warm, cool [2][16]uint32
	black      [16]uint32
	tick       int
}

func newGame() (_ *game, err error) {
	g := &game{art: ebiten.NewImageFromImage(indexedArtwork())}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	rgb565, err := palette.NewPackedRGB(palette.PackedRGBConfig{
		Bits: [3]uint8{5, 6, 5}, Shift: [3]uint8{11, 5, 0},
	})
	if err != nil {
		return nil, err
	}
	warm := [16]color.NRGBA{
		{A: 255}, {R: 18, G: 12, B: 38, A: 255}, {R: 36, G: 18, B: 70, A: 255},
		{R: 70, G: 28, B: 100, A: 255}, {R: 108, G: 39, B: 125, A: 255}, {R: 153, G: 57, B: 143, A: 255},
		{R: 30, G: 16, B: 58, A: 255}, {R: 50, G: 29, B: 83, A: 255},
		{R: 255, G: 224, B: 123, A: 255}, {R: 255, G: 180, B: 87, A: 255},
		{R: 255, G: 132, B: 78, A: 255}, {R: 245, G: 84, B: 111, A: 255},
		{R: 224, G: 64, B: 161, A: 255}, {R: 246, G: 123, B: 218, A: 255},
		{R: 21, G: 13, B: 39, A: 255}, {R: 50, G: 22, B: 67, A: 255},
	}
	for i, c := range warm {
		// The second bank changes the mood without changing a single artwork index.
		cool := color.NRGBA{R: c.B, G: c.R, B: c.G, A: 255}
		g.warm[0][i], g.cool[0][i] = rgb12Word(c), rgb12Word(cool)
		g.warm[1][i], g.cool[1][i] = rgb565Word(c), rgb565Word(cool)
	}
	for panel := range g.pages {
		c := effects.IndexedImageConfig{Image: g.art, Palette: g.black[:],
			FPS: fps, Channel: composite.BitplaneRed, Scale: 15}
		if panel == 0 {
			c.Crop = image.Rect(16, 16, 240, 160)
			c.Options.GeoM.Scale(1.65, 1.65)
			c.Options.GeoM.Translate(16, 44)
		} else {
			c.Format = rgb565
			c.Options.GeoM.Scale(1.4, 1.4)
			c.Options.GeoM.Translate(420, 42)
		}
		g.pages[panel], err = effects.NewIndexedImage(c)
		if err != nil {
			return nil, err
		}
		if err = g.pages[panel].Fade(g.black[:], g.warm[panel][:], 0, fadeTicks); err != nil {
			return nil, err
		}
	}
	return g, nil
}

// indexedArtwork stores index*17 in red. Indices and word precision are separate:
// the same 16-color image can use RGB12, RGB565 or another packed color format.
func indexedArtwork() *image.NRGBA {
	pixels := image.NewNRGBA(image.Rect(0, 0, 256, 176))
	for y := 0; y < 176; y++ {
		for x := 0; x < 256; x++ {
			index := 2 + y/24
			dx, dy := x-128, y-55
			if dx*dx+dy*dy <= 31*31 {
				index = 8 + min(3, max(0, (y-27)/15))
				if y > 54 && y%11 > 7 {
					index = 4
				}
			}
			ridge := 88 + 9*math.Sin(float64(x)*.055) + 5*math.Cos(float64(x)*.11)
			if float64(y) >= ridge {
				index = 6
				if y > 98+int(5*math.Sin(float64(x)*.08)) {
					index = 7
				}
			}
			if y >= 106 {
				// A static perspective floor; moving color comes entirely from palettes.
				column := float64(x-128) * 6 / float64(y-97)
				row := 140 / float64(y-97)
				index = 14 + ((int(math.Floor(column)) + int(math.Floor(row))) & 1)
				if column-math.Floor(column) < .08 || row-math.Floor(row) < .09 {
					index = 13
				}
			}
			if x < 16 || x >= 240 || y < 16 || y >= 160 {
				index = 0
			}
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(index * 17), A: 255})
		}
	}
	return pixels
}

func rgb12Word(c color.NRGBA) uint32 {
	return uint32(c.R>>4)<<8 | uint32(c.G>>4)<<4 | uint32(c.B>>4)
}

func rgb565Word(c color.NRGBA) uint32 {
	return uint32(c.R>>3)<<11 | uint32(c.G>>2)<<5 | uint32(c.B>>3)
}

func (g *game) Update() error {
	g.tick++
	if g.tick%cycleTicks == 0 {
		for panel, page := range g.pages {
			from, to := g.warm[panel][:], g.cool[panel][:]
			if g.tick/cycleTicks%2 == 0 {
				from, to = to, from
			}
			if err := page.Fade(from, to, g.tick, fadeTicks); err != nil {
				return err
			}
		}
	}
	// An absolute source clock makes seeking and deterministic captures repeatable.
	frame := kit.Frame{Time: float64(g.tick) / fps}
	for _, page := range g.pages {
		if err := page.Update(frame); err != nil {
			return err
		}
	}
	return nil
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 14, B: 24, A: 255})
	for _, page := range g.pages {
		page.Draw(dst)
	}
	ebitenutil.DebugPrintAt(dst, "RGB12 | CROPPED PAGE + PLACEMENT", 16, 12)
	ebitenutil.DebugPrintAt(dst, "RGB565 | COMPLETE PAGE + PLACEMENT", 420, 12)
	ebitenutil.DebugPrintAt(dst, "One shared index image | 50 Hz source clock | 2-second palette fades", 16, 302)
}

func (*game) Layout(int, int) (int, int) { return width, height }

func (g *game) Close() {
	for _, page := range g.pages {
		if page != nil {
			_ = page.Close()
		}
	}
	if g.art != nil {
		g.art.Deallocate()
	}
}

func main() {
	directory := flag.String("capture", "", "write a deterministic native PNG")
	frame := flag.Int("frame", 200, "update tick to capture")
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
	ebiten.SetTPS(fps)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Indexed Palette Pages")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
