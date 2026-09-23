//go:build dck_rows_rendercheck && !dck_background_rendercheck && !dck_water_rendercheck

package composite

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// This check compares cached geometry against independently submitted Row draws
// on a real graphics context, including overlaps and clipped destination images.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &rowProjectionCheck{}
	ebiten.SetWindowSize(256, 192)
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if game.err != nil {
		fmt.Fprintln(os.Stderr, game.err)
		os.Exit(1)
	}
}

type rowProjectionCheck struct {
	done bool
	err  error
}

func (*rowProjectionCheck) Layout(int, int) (int, int) { return 96, 80 }
func (g *rowProjectionCheck) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *rowProjectionCheck) Draw(*ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	g.err = checkProjectedRows()
}
func checkProjectedRows() error {
	pixels := image.NewNRGBA(image.Rect(0, 0, 36, 32))
	for y := range 32 {
		for x := range 36 {
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 7), G: uint8(y * 7), B: uint8(x * y), A: uint8(80 + (x+y)%176)})
		}
	}
	sheet := ebiten.NewImageFromImage(pixels)
	defer sheet.Deallocate()
	source := sheet.SubImage(image.Rect(3, 4, 32, 30)).(*ebiten.Image)
	actual, expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 96, 80), &ebiten.NewImageOptions{Unmanaged: true}), ebiten.NewImageWithOptions(image.Rect(0, 0, 96, 80), &ebiten.NewImageOptions{Unmanaged: true})
	defer actual.Deallocate()
	defer expected.Deallocate()
	got, want := make([]byte, 96*80*4), make([]byte, 96*80*4)
	cases := 0
	for _, filter := range []ebiten.Filter{ebiten.FilterNearest, ebiten.FilterLinear} {
		for _, height := range []float64{.01, .2, .7, 1, 3.25} {
			for _, origin := range []float64{-2.3, 0, 1.9, 4.5, 75.7} {
				rows := []Row{
					{Source: Region{X: 3, Y: 4, Width: 29, Height: 7}, X: 2.25, Y: origin, Width: 80.5, Height: height, Filter: filter},
					{Source: Region{X: 3, Y: 6, Width: 29, Height: 10}, X: -2.25, Y: origin + .25, Width: 90, Height: height * 2, Filter: filter},
					{Source: Region{X: 3, Y: 8, Width: 29, Height: 9}, X: 12.25, Y: origin + .75, Width: 64.5, Height: height, Filter: filter, Blend: ebiten.BlendLighter},
				}
				rows[1].ColorScale.Scale(.8, .6, .9, .75)
				projected, err := NewRowProjection(rows)
				if err != nil {
					return err
				}
				actual.Fill(color.NRGBA{R: 10, G: 20, B: 30, A: 100})
				expected.Fill(color.NRGBA{R: 10, G: 20, B: 30, A: 100})
				view := image.Rect(5, 1, 91, 76)
				projected.Draw(actual.SubImage(view).(*ebiten.Image), source)
				for _, row := range rows {
					row.Draw(expected.SubImage(view).(*ebiten.Image), source)
				}
				actual.ReadPixels(got)
				expected.ReadPixels(want)
				if !bytes.Equal(got, want) {
					return fmt.Errorf("row projection differs: filter %v height %g origin %g", filter, height, origin)
				}
				cases++
			}
		}
	}
	fmt.Printf("row projection: %d exact native pixel comparisons\n", cases)
	return nil
}
