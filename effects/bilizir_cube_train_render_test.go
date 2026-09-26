//go:build dck_bilizir_cube_rendercheck && !dck_jelly_rendercheck

package effects_test

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
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/presets"
)

// Opt-in GPU parity check for Bilizir's former twelve individual cube draws.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &bilizirCubeRenderCheck{}
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

type bilizirCubeRenderCheck struct {
	done bool
	err  error
}

func (*bilizirCubeRenderCheck) Layout(int, int) (int, int) { return 800, 600 }
func (game *bilizirCubeRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *bilizirCubeRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = compareBilizirCubePixels()
}

func compareBilizirCubePixels() error {
	train, err := effects.NewSolidCubeTrain(presets.BilizirCubeTrain(800, 12))
	if err != nil {
		return err
	}
	defer train.Close()
	var cubes [12]*effects.SolidCube
	var phases [12]float64
	for i := range cubes {
		cubes[i], err = effects.NewSolidCube(presets.BilizirCube(20))
		if err != nil {
			return err
		}
		defer cubes[i].Close()
		cubes[i].Rotation.X = float64(i) * .3
		cubes[i].Rotation.Y = float64(i) * .5
		cubes[i].Rotation.Z = float64(i) * .2
		phases[i] = .15 * float64(i+1)
	}
	actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 600), &ebiten.NewImageOptions{Unmanaged: true})
	expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 600), &ebiten.NewImageOptions{Unmanaged: true})
	defer actual.Deallocate()
	defer expected.Deallocate()
	got, want := make([]byte, 800*600*4), make([]byte, 800*600*4)
	for tick := 0; tick <= 3000; tick++ {
		if tick > 0 {
			speed := 1.0
			if tick >= 1000 && tick < 2000 {
				speed = 1.5
			} else if tick >= 2000 {
				speed = .5
			}
			if err := train.SetSpeed(speed); err != nil {
				return err
			}
			if err := train.Update(kit.Frame{}); err != nil {
				return err
			}
			for i, cube := range cubes {
				phases[i] += .04 * speed
				index := float64(i)
				cube.Rotate(.02*speed*(1+index*.1), .03*speed*(1+index*.15), .01*speed*(1+index*.05))
			}
		}
		if tick != 0 && tick != 1 && tick != 60 && tick != 999 && tick != 1000 && tick != 1999 && tick != 2000 && tick != 3000 {
			continue
		}
		actual.Fill(color.Black)
		expected.Fill(color.Black)
		train.Draw(actual)
		for i, cube := range cubes {
			x := 380.0 + 380*math.Sin(phases[i])
			y := 186.0 + 84*math.Cos(phases[i]*2.5)
			cube.DrawAt(expected, x, y)
		}
		actual.ReadPixels(got)
		expected.ReadPixels(want)
		if !bytes.Equal(got, want) {
			pixel := 0
			for got[pixel] == want[pixel] {
				pixel++
			}
			return fmt.Errorf("Bilizir cubes differ at tick %d pixel (%d,%d): got %v want %v",
				tick, (pixel/4)%800, pixel/(4*800), got[pixel/4*4:pixel/4*4+4], want[pixel/4*4:pixel/4*4+4])
		}
	}
	fmt.Println("Bilizir cube train matches individual cube draws at eight sampled ticks")
	return nil
}
