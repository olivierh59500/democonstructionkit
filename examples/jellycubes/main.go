package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/authoring"
	"github.com/olivierh59500/democonstructionkit/examples/jellycubes/scene"
)

type game struct {
	scene      *scene.Scene
	output     *ebiten.Image
	tick, last uint64
	limit      int
	rate       int
	capture    string
	done       bool
}

func (g *game) Update() error {
	if g.done || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.limit > 0 && int(g.tick) >= g.limit {
		return nil
	}
	g.tick++
	return g.scene.Update(kit.Frame{Tick: g.tick, Time: float64(g.tick) / float64(g.rate), Delta: 1.0 / float64(g.rate)})
}
func (g *game) Draw(dst *ebiten.Image) {
	if g.tick != g.last {
		g.scene.Draw(g.output)
		g.last = g.tick
	}
	dst.DrawImage(g.output, nil)
	if !g.done && g.limit > 0 && int(g.tick) == g.limit {
		if g.capture != "" {
			pixels := image.NewRGBA(g.output.Bounds())
			g.output.ReadPixels(pixels.Pix)
			file, err := os.Create(g.capture)
			if err != nil {
				log.Fatal(err)
			}
			if err = png.Encode(file, pixels); err != nil {
				file.Close()
				log.Fatal(err)
			}
			file.Close()
		}
		g.done = true
	}
}
func (g *game) Layout(int, int) (int, int) { return g.scene.Width, g.scene.Height }
func main() {
	eco := flag.Bool("eco", false, "render at 320x180")
	frames := flag.Int("frames", 0, "stop after this many updates; zero runs continuously")
	capture := flag.String("capture", "", "write the final frame as PNG (requires frames)")
	save := flag.String("save", "", "save the configured project as JSON")
	load := flag.String("project", "", "load a cube project JSON file")
	flag.Parse()
	if *frames < 0 || (*capture != "" && *frames == 0) {
		log.Fatal("invalid frame limit or capture configuration")
	}
	w, h := 640, 360
	if *eco {
		w, h = 320, 180
	}
	p := scene.Project(w, h)
	if *load != "" {
		f, err := os.Open(*load)
		if err != nil {
			log.Fatal(err)
		}
		decoded, err := authoring.Decode(f)
		f.Close()
		if err != nil {
			log.Fatal(err)
		}
		p = *decoded
	}
	if *save != "" {
		f, err := os.Create(*save)
		if err != nil {
			log.Fatal(err)
		}
		err = authoring.Encode(f, p)
		f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
	s, err := scene.New(p)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	g := &game{scene: s, output: ebiten.NewImage(s.Width, s.Height), limit: *frames, capture: *capture, rate: p.Canvas.TPS}
	defer g.output.Deallocate()
	ebiten.SetTPS(p.Canvas.TPS)
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("DCK / independent jelly cubes")
	ebiten.SetRunnableOnUnfocused(*frames > 0)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
