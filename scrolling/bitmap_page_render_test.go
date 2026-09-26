//go:build dck_bitmap_page_rendercheck && !dck_bitmap_rendercheck

package scrolling

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// Opt-in GPU check against the former per-glyph Phenomena intro page builder.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &bitmapPageRenderCheck{}
	ebiten.SetWindowSize(320, 240)
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if game.err != nil {
		fmt.Fprintln(os.Stderr, game.err)
		os.Exit(1)
	}
}

type bitmapPageRenderCheck struct {
	done bool
	err  error
}

func (*bitmapPageRenderCheck) Layout(int, int) (int, int) { return 640, 480 }
func (game *bitmapPageRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *bitmapPageRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = compareBitmapPages()
}

func compareBitmapPages() error {
	const order = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"
	pixels := image.NewNRGBA(image.Rect(0, 0, 16*45, 26))
	for y := 0; y < 26; y++ {
		for x := 0; x < 16*45; x++ {
			alpha := uint8(255)
			if (x+y)%5 == 0 {
				alpha = 0
			}
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 3), G: uint8(y * 9), B: uint8(x + y), A: alpha})
		}
	}
	atlas := ebiten.NewImageFromImage(pixels)
	defer atlas.Deallocate()
	base, err := GridImages(atlas, image.Pt(16, 26), len(order), len(order))
	if err != nil {
		return err
	}
	for _, sample := range []struct {
		name       string
		background color.Color
		lines      []BitmapPageLine
	}{
		{name: "inverted-black", background: color.Black, lines: []BitmapPageLine{
			{Text: "   FOR HOT VHS", X: 48, Y: 18, Advance: 32, ScaleX: 2, ScaleY: 2},
			{Text: "  AND SOFTWARE", X: 48, Y: 75, Advance: 32, ScaleX: 2, ScaleY: 2},
			{Text: "SWAPPING, CONTACT", X: 48, Y: 133, Advance: 32, ScaleX: 2, ScaleY: 2},
		}},
		{name: "normal-transparent", lines: []BitmapPageLine{
			{Text: "    PHENOMENA", X: 48, Y: 78, Advance: 32, ScaleX: 2, ScaleY: 2},
			{Text: "   SKALDEV. 69", X: 48, Y: 158, Advance: 32, ScaleX: 2, ScaleY: 2},
		}},
	} {
		glyphs := base
		if sample.background != nil {
			inverted := image.NewNRGBA(pixels.Bounds())
			for y := 0; y < 26; y++ {
				for x := 0; x < 16*45; x++ {
					c := pixels.NRGBAAt(x, y)
					inverted.SetNRGBA(x, y, color.NRGBA{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A})
				}
			}
			invertedAtlas := ebiten.NewImageFromImage(inverted)
			defer invertedAtlas.Deallocate()
			glyphs, err = GridImages(invertedAtlas, image.Pt(16, 26), len(order), len(order))
			if err != nil {
				return err
			}
		}
		page, err := NewBitmapPage(BitmapPageConfig{
			Width: 640, Height: 480, Order: order, Glyphs: glyphs,
			Lines: sample.lines, Background: sample.background,
		})
		if err != nil {
			return err
		}
		reference := ebiten.NewImage(640, 480)
		if sample.background != nil {
			reference.Fill(sample.background)
		}
		for _, line := range sample.lines {
			x := line.X
			for _, character := range line.Text {
				if index := strings.IndexRune(order, character); index >= 0 {
					var op ebiten.DrawImageOptions
					op.GeoM.Scale(2, 2)
					op.GeoM.Translate(x, line.Y)
					reference.DrawImage(glyphs[index], &op)
				}
				x += 32
			}
		}
		got, want := make([]byte, 640*480*4), make([]byte, 640*480*4)
		page.Image().ReadPixels(got)
		reference.ReadPixels(want)
		if !bytes.Equal(got, want) {
			pixel := 0
			for got[pixel] == want[pixel] {
				pixel++
			}
			return fmt.Errorf("%s bitmap page differs at (%d,%d): got %v, want %v", sample.name,
				(pixel/4)%640, pixel/(4*640), got[pixel/4*4:pixel/4*4+4], want[pixel/4*4:pixel/4*4+4])
		}
		page.Close()
		reference.Deallocate()
	}
	fmt.Println("Bitmap intro pages match the former per-glyph composition")
	return nil
}
