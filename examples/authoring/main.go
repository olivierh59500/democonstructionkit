// Command authoring saves, reloads and renders a DCK project with procedural assets.
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/authoring"
	"github.com/olivierh59500/democonstructionkit/examples/authoring/scene"
)

type game struct {
	scene   *scene.Scene
	tick    uint64
	frames  int
	capture string
	surface *ebiten.Image
	done    bool
	dirty   bool
	err     error
}

func (g *game) Update() error {
	if g.done {
		return ebiten.Termination
	}
	if g.frames > 0 && g.tick >= uint64(g.frames) {
		return nil
	}
	g.tick++
	g.dirty = true
	return g.scene.Update(kit.Frame{Tick: g.tick, Time: float64(g.tick) / float64(g.scene.Project.Canvas.TPS), Delta: 1 / float64(g.scene.Project.Canvas.TPS)})
}
func (g *game) Draw(dst *ebiten.Image) {
	if g.done {
		dst.DrawImage(g.surface, nil)
		return
	}
	if g.dirty {
		g.scene.Draw(g.surface)
		g.dirty = false
	}
	dst.DrawImage(g.surface, nil)
	if g.frames > 0 && g.tick >= uint64(g.frames) {
		g.done = true
		if g.capture != "" {
			b := g.surface.Bounds()
			pixels := image.NewRGBA(b)
			g.surface.ReadPixels(pixels.Pix)
			file, err := os.Create(g.capture)
			if err != nil {
				g.err = err
				return
			}
			g.err = png.Encode(file, pixels)
			if err := file.Close(); g.err == nil {
				g.err = err
			}
		}
	}
}
func (g *game) Layout(int, int) (int, int) { return g.scene.Width, g.scene.Height }

func main() {
	projectFile := flag.String("project", "", "load a version-1 JSON project using this example's asset IDs")
	save := flag.String("save", "", "save the validated project as JSON before playing")
	frames := flag.Int("frames", 0, "stop after this many ticks; zero plays interactively")
	capture := flag.String("capture", "", "write the final frame as PNG; requires -frames")
	eco := flag.Bool("eco", false, "use the 320x180 default project")
	flag.Parse()
	if *frames < 0 || (*capture != "" && *frames == 0) {
		log.Fatal("capture requires a positive frame count")
	}
	var s *scene.Scene
	var err error
	if *projectFile != "" {
		if *eco {
			log.Fatal("use -eco for the default project or provide a resized project")
		}
		file, err := os.Open(*projectFile)
		if err != nil {
			log.Fatal(err)
		}
		project, decodeErr := authoring.Decode(file)
		closeErr := file.Close()
		if decodeErr != nil {
			log.Fatal(decodeErr)
		}
		if closeErr != nil {
			log.Fatal(closeErr)
		}
		s, err = scene.New(project)
	} else if *eco {
		s, err = scene.NewSize(320, 180)
	} else {
		s, err = scene.New(nil)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if *save != "" {
		file, err := os.Create(*save)
		if err != nil {
			log.Fatal(err)
		}
		saveErr := authoring.Encode(file, s.Project)
		closeErr := file.Close()
		if saveErr != nil {
			log.Fatal(saveErr)
		}
		if closeErr != nil {
			log.Fatal(closeErr)
		}
	}
	surface := ebiten.NewImageWithOptions(image.Rect(0, 0, s.Width, s.Height), &ebiten.NewImageOptions{Unmanaged: true})
	defer surface.Deallocate()
	g := &game{scene: s, frames: *frames, capture: *capture, surface: surface}
	ebiten.SetTPS(s.Project.Canvas.TPS)
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("DCK Authoring / Saved composition")
	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
	if g.err != nil {
		log.Fatal(g.err)
	}
}
