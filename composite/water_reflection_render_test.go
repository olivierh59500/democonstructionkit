//go:build dck_water_rendercheck

package composite

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// This optional check needs a native graphics session because it compares actual
// GPU output. Run go test -tags dck_water_rendercheck ./composite.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	g := &waterRenderCheck{}
	ebiten.SetWindowSize(320, 240)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if g.err != nil {
		fmt.Fprintln(os.Stderr, g.err)
		os.Exit(1)
	}
	fmt.Println("Water reflection GPU output matches original flat and constant-wave transforms")
}

type waterRenderCheck struct {
	done bool
	err  error
}

func (g *waterRenderCheck) Layout(int, int) (int, int) { return 640, 480 }
func (g *waterRenderCheck) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return nil
}
func (g *waterRenderCheck) Draw(_ *ebiten.Image) {
	if g.done {
		return
	}
	g.done = true
	g.err = checkWaterPixels()
}

func checkWaterPixels() error {
	pixels := image.NewRGBA(image.Rect(0, 0, 700, 450))
	for y := range 450 {
		for x := range 700 {
			pixels.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: uint8(x ^ y), A: uint8(80 + (x+y)%176)})
		}
	}
	source := ebiten.NewImageFromImage(pixels)
	defer source.Deallocate()
	subSource := source.SubImage(image.Rect(10, 20, 690, 420)).(*ebiten.Image)
	actual, expected := ebiten.NewImage(700, 500), ebiten.NewImage(700, 500)
	defer actual.Deallocate()
	defer expected.Deallocate()
	got, want := make([]byte, 700*500*4), make([]byte, 700*500*4)
	for index, crop := range []image.Rectangle{
		image.Rect(0, 288, 640, 368), image.Rect(10, 20, 75, 35), image.Rect(5, 15, 75, 35), {},
	} {
		input := subSource
		if index == 0 {
			input = source
		}
		for _, wave := range []bool{false, true} {
			config := DefaultWaterReflectionConfig()
			config.Source = crop
			config.X, config.Horizon = 5, 50
			if index == 0 {
				config.X, config.Horizon = 0, 400
			}
			if index > 0 {
				config.Tint.Scale(.8, .6, .9, .75)
				if !wave {
					config.ScaleY = .75
				}
			}
			if wave {
				// Every row rounds to an exact four-pixel shift at float32 precision.
				config.Wave = WaterWave{Amplitude: 4, Wavelength: 1e15, Phase: math.Pi / 2}
			}
			reflection, err := NewWaterReflection(config)
			if err != nil {
				return err
			}
			for frame := range 2 {
				// Reusing the pass after mutating its input catches stale snapshots.
				if frame == 1 {
					subSource.Fill(color.NRGBA{R: 190, G: 70, B: 40, A: 128})
				} else {
					source.WritePixels(pixels.Pix)
				}
				actual.Fill(color.RGBA{B: 122, A: 255})
				expected.Fill(color.RGBA{B: 122, A: 255})
				reflection.Draw(actual, input, kit.Frame{Time: float64(frame)})
				clipped := crop
				if clipped.Empty() {
					clipped = input.Bounds()
				} else {
					clipped = clipped.Intersect(input.Bounds())
				}
				view := input.SubImage(clipped).(*ebiten.Image)
				op := ebiten.DrawImageOptions{ColorScale: config.Tint}
				op.GeoM.Scale(1, -config.ScaleY)
				x := config.X
				if wave {
					x += 4
				}
				op.GeoM.Translate(x, config.Horizon+float64(clipped.Dy())*config.ScaleY)
				op.ColorScale.ScaleAlpha(config.Alpha)
				expected.DrawImage(view, &op)
				actual.ReadPixels(got)
				expected.ReadPixels(want)
				if !bytes.Equal(got, want) {
					for i := range got {
						if got[i] != want[i] {
							return fmt.Errorf("water GPU mismatch: crop %d, wave %v, frame %d, byte %d: got %d, want %d", index, wave, frame, i, got[i], want[i])
						}
					}
				}
			}
			_ = reflection.Close()
		}
	}
	return checkWaterFade()
}

func checkWaterFade() error {
	source, destination := ebiten.NewImage(50, 100), ebiten.NewImage(50, 100)
	defer source.Deallocate()
	defer destination.Deallocate()
	source.Fill(color.RGBA{R: 100, G: 60, B: 20, A: 255})
	config := DefaultWaterReflectionConfig()
	config.Alpha, config.Fade = 1, 1
	reflection, err := NewWaterReflection(config)
	if err != nil {
		return err
	}
	defer reflection.Close()
	reflection.Draw(destination, source, kit.Frame{})
	pixels := make([]byte, 50*100*4)
	destination.ReadPixels(pixels)
	for y := range 100 {
		fade := 1 - (float64(y)+.5)/100
		for channel, value := range [4]float64{100, 60, 20, 255} {
			want := math.Round(value * fade)
			got := float64(pixels[(y*50+10)*4+channel])
			if math.Abs(got-want) > 1 {
				return fmt.Errorf("water fade mismatch: row %d, channel %d: got %g, want %g", y, channel, got, want)
			}
		}
	}
	return nil
}
