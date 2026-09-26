//go:build dck_copper_title_rendercheck && !dck_window_bank_rendercheck && !dck_background_rendercheck && !dck_water_rendercheck

package composite

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// Opt-in GPU parity check for the previous isolated and direct title pipelines.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &copperTitleRenderCheck{}
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

type copperTitleRenderCheck struct {
	done bool
	err  error
}

func (*copperTitleRenderCheck) Layout(int, int) (int, int) { return 800, 100 }
func (game *copperTitleRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *copperTitleRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = compareCopperTitlePixels()
}

func compareCopperTitlePixels() error {
	titlePixels := image.NewNRGBA(image.Rect(0, 0, 120, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 120; x++ {
			alpha := uint8(255)
			if (x/9+y/5)%4 == 0 {
				alpha = 0
			}
			titlePixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 2), G: uint8(y * 9), B: 240, A: alpha})
		}
	}
	barPixels := image.NewNRGBA(image.Rect(0, 0, 46, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 46; x++ {
			barPixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 5), G: uint8(y * 12), B: 80, A: 255})
		}
	}
	title, bars := ebiten.NewImageFromImage(titlePixels), ebiten.NewImageFromImage(barPixels)
	defer title.Deallocate()
	defer bars.Deallocate()
	offsets := make([]int, 16)
	for i := range offsets {
		offsets[i] = (i*11)%67 - 33
	}
	for _, mode := range []CopperTitleMode{CopperTitleSurface, CopperTitleDirect} {
		clockMode := SingleWrapClock
		if mode == CopperTitleDirect {
			clockMode = MaskedClock
		}
		copperConfig := CopperBarsConfig{
			Image: bars, Offsets: offsets, Height: 72, Count: 36, RowStep: 2,
			SourceStep: 2, SourcePeriod: 20, BaseX: 60, XShift: 1,
			VelocityA: 3, VelocityB: -5, IndexStepA: 7, IndexStepB: 10,
			Clock: clockMode, DrawMode: CopperImages,
		}
		motionConfig := motion.WaveClockConfig{
			Wave:  motion.Wave{Amplitude: 800, Offset: 64, Speed: 1, Cos: true},
			Start: .5, Step: .0125,
		}
		band, err := NewCopperTitleBand(CopperTitleBandConfig{
			Title: title, Copper: &copperConfig, TitleMotion: motionConfig,
			Width: 800, Height: 72, Mode: mode, Background: color.Black,
		})
		if err != nil {
			return err
		}
		copper, err := NewCopperBars(copperConfig)
		if err != nil {
			return err
		}
		clock, err := motion.NewWaveClock(motionConfig)
		if err != nil {
			return err
		}
		var oldLayer *SurfaceLayer
		if mode == CopperTitleSurface {
			oldLayer, err = NewSurfaceLayer(SurfaceLayerConfig{
				Width: 800, Height: 72, Background: color.Black,
				Sources: []kit.Effect{copper},
				Passes:  []SurfaceImagePass{{Image: title, X: clock.At(0), ScaleY: 3}},
				Outputs: []SurfaceOutput{{}},
			})
			if err != nil {
				return err
			}
		}
		actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 100), &ebiten.NewImageOptions{Unmanaged: true})
		expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 100), &ebiten.NewImageOptions{Unmanaged: true})
		gotPixels, wantPixels := make([]byte, 800*100*4), make([]byte, 800*100*4)
		for tick := 0; tick <= 400; tick++ {
			if tick > 0 {
				speed := 1.0
				if tick >= 200 && tick < 300 {
					speed = 1.5
				} else if tick >= 300 {
					speed = .5
				}
				if err := band.Advance(speed); err != nil {
					return err
				}
				if err := copper.SetSpeed(speed); err != nil {
					return err
				}
				if oldLayer != nil {
					if err := oldLayer.Update(kit.Frame{}); err != nil {
						return err
					}
				} else if err := copper.Update(kit.Frame{}); err != nil {
					return err
				}
				if err := clock.SetStep(.0125 * speed); err != nil {
					return err
				}
				clock.Step()
				if oldLayer != nil {
					if err := oldLayer.SetPassPosition(0, clock.At(0), 0); err != nil {
						return err
					}
				}
			}
			if tick != 0 && tick != 1 && tick != 17 && tick != 199 && tick != 200 && tick != 299 && tick != 300 && tick != 400 {
				continue
			}
			actual.Fill(color.NRGBA{R: 13, G: 27, B: 60, A: 255})
			expected.Fill(color.NRGBA{R: 13, G: 27, B: 60, A: 255})
			band.Draw(actual)
			if oldLayer != nil {
				oldLayer.Draw(expected)
			} else {
				vector.DrawFilledRect(expected, 0, 0, 800, 72, color.Black, false)
				copper.Draw(expected)
				var op ebiten.DrawImageOptions
				op.GeoM.Scale(1, 3)
				op.GeoM.Translate(clock.At(0), 0)
				expected.DrawImage(title, &op)
			}
			actual.ReadPixels(gotPixels)
			expected.ReadPixels(wantPixels)
			if !bytes.Equal(gotPixels, wantPixels) {
				pixel := 0
				for gotPixels[pixel] == wantPixels[pixel] {
					pixel++
				}
				return fmt.Errorf("title band mode %d differs at tick %d pixel (%d,%d): got %v want %v",
					mode, tick, (pixel/4)%800, pixel/(4*800), gotPixels[pixel/4*4:pixel/4*4+4], wantPixels[pixel/4*4:pixel/4*4+4])
			}
		}
		actual.Deallocate()
		expected.Deallocate()
		if oldLayer != nil {
			oldLayer.Close()
		}
		band.Close()
	}
	fmt.Println("Copper title band matches surface and direct reference pixels at eight phases each")
	return nil
}
