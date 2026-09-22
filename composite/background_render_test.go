//go:build dck_background_rendercheck && !dck_water_rendercheck

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

// The opt-in test compares actual GPU pixels against uncropped brute-force image
// placement, including transparent overlaps, subimages and fractional filtering.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	g := &backgroundRenderCheck{}
	ebiten.SetWindowSize(320, 240)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if g.err != nil {
		fmt.Fprintln(os.Stderr, g.err)
		os.Exit(1)
	}
}

type backgroundRenderCheck struct {
	done bool
	err  error
}

func (*backgroundRenderCheck) Layout(int, int) (int, int) { return 96, 80 }
func (g *backgroundRenderCheck) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *backgroundRenderCheck) Draw(*ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	g.err = checkBackgroundPixels()
}

func checkBackgroundPixels() error {
	pixels := image.NewNRGBA(image.Rect(0, 0, 36, 32))
	for y := range 32 {
		for x := range 36 {
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 7), G: uint8(y * 7), B: uint8(x * y), A: uint8(80 + (x+y)%176)})
		}
	}
	source := ebiten.NewImageFromImage(pixels)
	defer source.Deallocate()
	input := source.SubImage(image.Rect(3, 4, 32, 30)).(*ebiten.Image)
	actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 96, 80), &ebiten.NewImageOptions{Unmanaged: true})
	expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 96, 80), &ebiten.NewImageOptions{Unmanaged: true})
	defer actual.Deallocate()
	defer expected.Deallocate()
	got, want := make([]byte, 96*80*4), make([]byte, 96*80*4)
	count := 0
	for _, crop := range []image.Rectangle{{}, image.Rect(7, 9, 20, 22), image.Rect(0, 0, 20, 22)} {
		for _, period := range [][2]float64{{0, 0}, {17, 0}, {0, 19}, {17, 19}, {17, 18}, {9, 11}} {
			for _, filter := range []ebiten.Filter{ebiten.FilterNearest, ebiten.FilterLinear} {
				for _, scale := range []float64{1, 2, .7, 1.3} {
					for _, pose := range []BackgroundPose{{X: -14.25, Y: 11.5, CameraX: 7, CameraY: 11}, {X: -14, Y: 12, CameraX: 8, CameraY: 10}} {
						c := DefaultBackgroundConfig()
						c.Source = crop
						c.PeriodX, c.PeriodY = period[0], period[1]
						c.ScaleX, c.ScaleY = scale, scale
						c.ParallaxX, c.ParallaxY = .5, 1.5
						c.Filter = filter
						c.ColorScale.Scale(.8, .6, .9, .75)
						background, err := NewBackground(c)
						if err != nil {
							return err
						}
						actual.Fill(color.NRGBA{R: 10, G: 20, B: 30, A: 100})
						expected.Fill(color.NRGBA{R: 10, G: 20, B: 30, A: 100})
						view := image.Rect(5, 6, 91, 74)
						background.Draw(actual.SubImage(view).(*ebiten.Image), input, pose)
						tile := input
						if !crop.Empty() {
							tile = input.SubImage(crop.Intersect(input.Bounds())).(*ebiten.Image)
						}
						minX, maxX, minY, maxY := -40, 40, -40, 40
						if c.PeriodX == 0 {
							minX, maxX = 0, 0
						}
						if c.PeriodY == 0 {
							minY, maxY = 0, 0
						}
						for iy := minY; iy <= maxY; iy++ {
							for ix := minX; ix <= maxX; ix++ {
								op := ebiten.DrawImageOptions{Filter: filter, ColorScale: c.ColorScale}
								op.GeoM.Scale(scale, scale)
								originX := pose.X - float64(pose.CameraX*c.ParallaxX*c.ScaleX)
								originY := pose.Y - float64(pose.CameraY*c.ParallaxY*c.ScaleY)
								op.GeoM.Translate(originX+float64(float64(ix)*c.PeriodX*c.ScaleX), originY+float64(float64(iy)*c.PeriodY*c.ScaleY))
								expected.SubImage(view).(*ebiten.Image).DrawImage(tile, &op)
							}
						}
						actual.ReadPixels(got)
						expected.ReadPixels(want)
						if !bytes.Equal(got, want) {
							index := 0
							for got[index] == want[index] {
								index++
							}
							return fmt.Errorf("background GPU mismatch crop=%v period=%v filter=%v scale=%v pixel=(%d,%d) got=%v want=%v", crop, period, filter, scale, (index/4)%96, index/(96*4), got[index/4*4:index/4*4+4], want[index/4*4:index/4*4+4])
						}
						count++
					}
				}
			}
		}
	}
	fmt.Printf("Background GPU output matches brute-force placement in %d cases\n", count)
	return nil
}
