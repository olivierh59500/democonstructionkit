//go:build dck_window_bank_rendercheck && !dck_background_rendercheck && !dck_water_rendercheck

package composite

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// This opt-in GPU check compares the direct window renderer to the previous
// six-surface pipeline at startup, visible strips, wrap and later playback.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &windowedImageRenderCheck{}
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

type windowedImageRenderCheck struct {
	done bool
	err  error
}

func (*windowedImageRenderCheck) Layout(int, int) (int, int) { return 768, 540 }
func (game *windowedImageRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *windowedImageRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = checkWindowedImagePixels()
}

func checkWindowedImagePixels() error {
	pixels := image.NewNRGBA(image.Rect(0, 0, 640, 1024))
	for y := 0; y < 1024; y++ {
		for x := 0; x < 640; x++ {
			alpha := uint8(255)
			if (x/19+y/13)%7 == 0 {
				alpha = 0
			}
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 3), G: uint8(y), B: uint8(x + y), A: alpha})
		}
	}
	source := ebiten.NewImageFromImage(pixels)
	defer source.Deallocate()
	actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 768, 540), &ebiten.NewImageOptions{Unmanaged: true})
	expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 768, 540), &ebiten.NewImageOptions{Unmanaged: true})
	defer actual.Deallocate()
	defer expected.Deallocate()
	var strips [6]*ebiten.Image
	windows := make([]WindowedImageWindow, len(strips))
	for i := range strips {
		strips[i] = ebiten.NewImageWithOptions(image.Rect(0, 0, 768, 32), &ebiten.NewImageOptions{Unmanaged: true})
		defer strips[i].Deallocate()
		y := 132 + i*34
		windows[i] = WindowedImageWindow{Clip: image.Rect(0, y, 768, y+32), Offset: motion.Point{Y: -float64(i * 5)}}
	}
	bank, err := NewWindowedImageBank(WindowedImageBankConfig{
		Image: source, Windows: windows, ScaleX: 1.3, Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
		MotionY: &motion.WrapBankConfig{
			Start: []float64{174}, Velocity: []float64{-1.5},
			Lower: &motion.WrapLimit{Boundary: -974, Restart: 174, Inclusive: true},
		},
	})
	if err != nil {
		return err
	}
	got, want := make([]byte, 768*540*4), make([]byte, 768*540*4)
	offset := 174.0
	checks := map[int]bool{0: true, 100: true, 116: true, 130: true, 301: true, 765: true, 766: true, 900: true, 1532: true}
	for tick := 0; tick <= 1532; tick++ {
		if checks[tick] {
			background := color.NRGBA{R: 13, G: 27, B: 60, A: 255}
			actual.Fill(background)
			expected.Fill(background)
			bank.Draw(actual)
			for i, strip := range strips {
				strip.Clear()
				op := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver}
				op.GeoM.Scale(1.3, 1)
				op.GeoM.Translate(0, offset-float64(i*5))
				strip.DrawImage(source, &op)
				output := ebiten.DrawImageOptions{Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver}
				output.GeoM.Translate(0, float64(132+i*34))
				expected.DrawImage(strip, &output)
			}
			actual.ReadPixels(got)
			expected.ReadPixels(want)
			if !bytes.Equal(got, want) {
				pixel := 0
				for got[pixel] == want[pixel] {
					pixel++
				}
				return fmt.Errorf("windowed image GPU mismatch at tick %d pixel (%d,%d): got %v, want %v", tick,
					(pixel/4)%768, pixel/(4*768), got[pixel/4*4:pixel/4*4+4], want[pixel/4*4:pixel/4*4+4])
			}
		}
		offset -= 1.5
		if offset <= -974 {
			offset = 174
		}
		bank.Step()
	}
	fmt.Println("Windowed image GPU output matches the six-surface raster in 9 sampled frames")
	return nil
}
