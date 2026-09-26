//go:build dck_tiled_wave_rendercheck && !dck_copper_title_rendercheck && !dck_window_bank_rendercheck && !dck_background_rendercheck && !dck_water_rendercheck

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

// Opt-in GPU comparison against the two earlier tiled-wave pipelines.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &tiledWaveRenderCheck{}
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

type tiledWaveRenderCheck struct {
	done bool
	err  error
}

func (*tiledWaveRenderCheck) Layout(int, int) (int, int) { return 640, 400 }
func (game *tiledWaveRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *tiledWaveRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = compareTiledWavePixels()
}

func tileTestImage(width, height int) *ebiten.Image {
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 7), G: uint8(y * 5), B: uint8(x * y), A: 255})
		}
	}
	return ebiten.NewImageFromImage(pixels)
}

func compareTiledWavePixels() error {
	for _, sample := range []struct {
		name                                    string
		tileWidth, tileHeight, periodX, periodY int
		canvas                                  image.Point
		wave                                    WaveStrips
		outputX, outputY                        float64
		motion                                  bool
	}{
		{name: "Mega Scroller", tileWidth: 8, tileHeight: 8, periodX: 8, periodY: 8,
			canvas: image.Pt(500, 240), outputX: -110, outputY: -9,
			wave: WaveStrips{Axis: Rows, Thickness: 1, Filter: ebiten.FilterNearest,
				Waves: []StripWave{{Amplitude: 30, Spatial: .03, Speed: -.05}, {Amplitude: 30, Spatial: .01, Speed: .08}}}},
		{name: "LED Scroller", tileWidth: 32, tileHeight: 33, periodX: 32, periodY: 33,
			canvas: image.Pt(384, 233), outputX: -32, motion: true,
			wave: WaveStrips{Axis: Rows, Thickness: 1, Filter: ebiten.FilterLinear,
				Waves: []StripWave{{Amplitude: 6, Spatial: .08, Speed: .2}}}},
	} {
		tile := tileTestImage(sample.tileWidth, sample.tileHeight)
		tilesConfig := BackgroundConfig{PeriodX: float64(sample.periodX), PeriodY: float64(sample.periodY), Filter: ebiten.FilterLinear}
		config := TiledWaveBackdropConfig{
			Tile: tile, TileCanvas: sample.canvas, Tiles: tilesConfig, Wave: sample.wave,
			OutputX: sample.outputX, OutputY: sample.outputY,
		}
		if sample.motion {
			config.RetainSource = true
			config.SourceFilter, config.SourceBlend = ebiten.FilterLinear, ebiten.BlendSourceOver
			config.MotionY = &motion.WrapBankConfig{
				Start: []float64{0}, Velocity: []float64{-2},
				Lower: &motion.WrapLimit{Boundary: -33, Restart: 0, Inclusive: true},
			}
		}
		backdrop, err := NewTiledWaveBackdrop(config)
		if err != nil {
			return err
		}
		oldTiled := ebiten.NewImageWithOptions(image.Rect(0, 0, sample.canvas.X, sample.canvas.Y), &ebiten.NewImageOptions{Unmanaged: true})
		background, err := NewBackground(tilesConfig)
		if err != nil {
			return err
		}
		background.DrawAt(oldTiled, tile, 0, 0)
		var oldSource *ebiten.Image
		var oldMotion *motion.WrapBank
		if sample.motion {
			oldSource = ebiten.NewImageWithOptions(image.Rect(0, 0, sample.canvas.X, sample.canvas.Y), &ebiten.NewImageOptions{Unmanaged: true})
			oldMotion, err = motion.NewWrapBank(*config.MotionY)
			if err != nil {
				return err
			}
		}
		oldWave := sample.wave
		oldWave.Waves = append([]StripWave(nil), sample.wave.Waves...)
		actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 640, 400), &ebiten.NewImageOptions{Unmanaged: true})
		expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 640, 400), &ebiten.NewImageOptions{Unmanaged: true})
		got, want := make([]byte, 640*400*4), make([]byte, 640*400*4)
		frames := 80
		if !sample.motion {
			frames = 5
		}
		for frame := 0; frame < frames; frame++ {
			actual.Clear()
			expected.Clear()
			source := oldTiled
			if oldSource != nil {
				var op ebiten.DrawImageOptions
				op.Filter, op.Blend = ebiten.FilterLinear, ebiten.BlendSourceOver
				op.GeoM.Translate(0, oldMotion.At(0))
				oldSource.DrawImage(oldTiled, &op)
				source = oldSource
			}
			oldWave.DrawAt(expected, source, sample.outputX, sample.outputY)
			backdrop.Draw(actual)
			if frame == 0 || frame == 1 || frame == 16 || frame == 17 || frame == 33 || frame == 64 || frame == 79 {
				actual.ReadPixels(got)
				expected.ReadPixels(want)
				if !bytes.Equal(got, want) {
					pixel := 0
					for got[pixel] == want[pixel] {
						pixel++
					}
					return fmt.Errorf("%s differs at frame %d pixel (%d,%d): got %v want %v", sample.name, frame,
						(pixel/4)%640, pixel/(4*640), got[pixel/4*4:pixel/4*4+4], want[pixel/4*4:pixel/4*4+4])
				}
			}
			if oldMotion != nil {
				oldMotion.Step()
			}
			oldWave.Advance()
			backdrop.Step()
		}
		if oldSource != nil {
			oldSource.Deallocate()
		}
		oldTiled.Deallocate()
		actual.Deallocate()
		expected.Deallocate()
		backdrop.Close()
		tile.Deallocate()
	}
	fmt.Println("Tiled wave backdrop matches both preceding pipelines at sampled frames")
	return nil
}
