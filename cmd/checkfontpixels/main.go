// Command checkfontpixels compares every cached preset glyph with its source pixels.
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

type check struct {
	root string
	done bool
	err  error
}

func main() {
	root := flag.String("demos", "../../demos", "local demo asset directory")
	flag.Parse()
	g := &check{root: *root}
	ebiten.SetWindowSize(320, 160)
	ebiten.SetWindowTitle("DCK atlas pixel checks")
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if g.err != nil {
		fmt.Fprintln(os.Stderr, g.err)
		os.Exit(1)
	}
}

func (*check) Layout(int, int) (int, int) { return 320, 160 }
func (g *check) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *check) Draw(*ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	for _, spec := range presets.Fonts() {
		if g.err = checkAtlas(g.root, spec); g.err != nil {
			return
		}
	}
}

func checkAtlas(root string, spec presets.Font) error {
	f, err := os.Open(filepath.Join(root, filepath.FromSlash(spec.Path)))
	if err != nil {
		return err
	}
	source, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return err
	}
	metrics, err := spec.Build(source.Bounds())
	if err != nil {
		return err
	}
	img := ebiten.NewImageFromImage(source)
	defer img.Deallocate()
	atlas, err := scrolling.NewAtlas(img, metrics)
	if err != nil {
		return err
	}
	hash := sha256.New()
	count := 0
	for _, r := range metrics.Characters() {
		view, m, _ := atlas.ExactGlyph(r)
		if m.Rect.Empty() {
			if view != nil {
				return fmt.Errorf("%s %q: blank has pixels", spec.ID, r)
			}
			continue
		}
		want := image.NewRGBA(image.Rect(0, 0, m.Rect.Dx(), m.Rect.Dy()))
		draw.Draw(want, want.Bounds(), source, m.Rect.Min, draw.Src)
		got := make([]byte, len(want.Pix))
		view.ReadPixels(got)
		if !bytes.Equal(got, want.Pix) {
			return fmt.Errorf("%s %q: cached pixels differ from atlas crop", spec.ID, r)
		}
		fmt.Fprintf(hash, "%d:%v:%g\n", r, m.Rect, m.Advance)
		_, _ = hash.Write(got)
		count++
	}
	fmt.Printf("%s: %d exact glyphs sha256=%x\n", spec.ID, count, hash.Sum(nil))
	return nil
}
