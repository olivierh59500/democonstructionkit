//go:build dck_bitmap_rendercheck

package scrolling

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	g := &bitmapBandsRenderCheck{}
	ebiten.SetWindowSize(256, 192)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if g.err != nil {
		fmt.Fprintln(os.Stderr, g.err)
		os.Exit(1)
	}
}

type bitmapBandsRenderCheck struct {
	done bool
	err  error
}

func (*bitmapBandsRenderCheck) Layout(int, int) (int, int) { return 64, 24 }
func (g *bitmapBandsRenderCheck) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *bitmapBandsRenderCheck) Draw(*ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	g.err = checkBitmapBandsPixels()
}
func checkBitmapBandsPixels() error {
	surface := func(w, h int) *ebiten.Image {
		return ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 24, 4))
	for y := range 4 {
		for x := range 24 {
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 9), G: uint8(y * 60), B: uint8(x * y * 2), A: uint8(60 + (x*13+y*9)%190)})
		}
	}
	atlas := ebiten.NewImageFromImage(pixels)
	defer atlas.Deallocate()
	font, _ := (BitmapSpec{Width: 8, Height: 4, Order: "ABC"}).Grid(atlas, ebiten.FilterLinear)
	text := strings.Repeat("ABC", 12)
	cfg := BitmapBandsConfig{Font: font, Text: text, Width: 64, Height: 24, Entrance: 64, Repeat: true, UseTicks: true, Lanes: []BitmapLane{{Speed: 3}, {Y: 8, Speed: 5}, {Y: 16, Speed: 7}}}
	bands, err := NewBitmapBands(cfg)
	if err != nil {
		return err
	}
	defer bands.Close()
	period := len(text) * 8
	original := surface(period+64, 4)
	defer original.Deallocate()
	head := surface(64, 4)
	defer head.Deallocate()
	font.Print(original, text, 0, 0, 1, 1)
	composite.DrawRegion(head, original, composite.Region{Width: 64, Height: 4}, &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear})
	op := ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(float64(period), 0)
	original.DrawImage(head, &op)
	expected, actual := surface(64, 24), surface(64, 24)
	defer expected.Deallocate()
	defer actual.Deallocate()
	want, got := make([]byte, 64*24*4), make([]byte, 64*24*4)
	ticks := []uint64{0, 1, 8, 9, 12, 13, 21, 22, 100, 101, 1000, 100000}
	for _, tick := range ticks {
		actual.Clear()
		expected.Clear()
		if err := bands.Update(kit.Frame{Tick: tick}); err != nil {
			return err
		}
		bands.Draw(actual)
		for _, lane := range cfg.Lanes {
			travel := lane.Speed * float64(tick)
			x, source := 0.0, 0.0
			if travel < 64 {
				x = 64 - travel
			} else {
				source = math.Mod(travel-64, float64(period))
			}
			op := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
			op.GeoM.Translate(x, lane.Y)
			composite.DrawRegion(expected, original, composite.Region{X: source, Width: 64, Height: 4}, &op)
		}
		actual.ReadPixels(got)
		expected.ReadPixels(want)
		if !bytes.Equal(got, want) {
			return fmt.Errorf("bounded text differs from complete strip at tick %d", tick)
		}
	}
	// A 960,000-pixel logical message must render through the same 64×24 target.
	cfg.Text = strings.Repeat("ABC", 40000)
	long, err := NewBitmapBands(cfg)
	if err != nil {
		return err
	}
	defer long.Close()
	for _, tick := range []uint64{0, 1, 100, 137143, 320001, 10000000} {
		actual.Clear()
		expected.Clear()
		_ = bands.Update(kit.Frame{Tick: tick})
		bands.Draw(expected)
		_ = long.Update(kit.Frame{Tick: tick})
		long.Draw(actual)
		actual.ReadPixels(got)
		expected.ReadPixels(want)
		if !bytes.Equal(got, want) {
			return fmt.Errorf("long repeated message differs at tick %d", tick)
		}
	}
	window, err := NewBitmapText(font, text, 0)
	if err != nil {
		return err
	}
	view := image.Rect(13, 3, 61, 12)
	for _, x := range []float64{-400, -21.5, 0, 13, 50} {
		actual.Clear()
		expected.Clear()
		if err := window.DrawWindow(actual.SubImage(view).(*ebiten.Image), x, 4, float64(view.Dx())); err != nil {
			return err
		}
		font.Print(expected.SubImage(view).(*ebiten.Image), text, x, 4, 1, 1)
		actual.ReadPixels(got)
		expected.ReadPixels(want)
		if !bytes.Equal(got, want) {
			return fmt.Errorf("nonzero-origin text window differs at X=%g", x)
		}
	}
	fmt.Println("bitmap bands: 12 exact full-strip comparisons 6 exact 120,000-character comparisons, and 5 nonzero-origin windows")
	return nil
}
