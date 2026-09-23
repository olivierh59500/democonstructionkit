//go:build dck_jelly_rendercheck

package effects

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/render"
)

var jellyCheckFrames = []int{0, 1, 90, 298, 299, 300, 301, 321, 343, 345, 390, 598, 599, 600, 601, 621, 643, 645, 690, 898, 899, 900, 901, 921, 943, 945, 990, 1198, 1199, 1200, 1201, 1221, 1243, 1245, 1290, 1498, 1499, 1500, 1501, 1521, 1543, 1545}

type jellyRenderCheck struct {
	cube             *JellyCube
	old              *jellyLegacy
	frame, checked   int
	actual, expected *ebiten.Image
	pixelsA, pixelsB []byte
	err              error
}

func (c *jellyRenderCheck) Layout(int, int) (int, int) { return 640, 400 }
func (c *jellyRenderCheck) Update() error {
	if c.err != nil {
		return c.err
	}
	c.frame++
	if err := c.cube.Update(kit.Frame{Time: float64(c.frame) / 60}); err != nil {
		return err
	}
	return c.old.update()
}
func (c *jellyRenderCheck) Draw(dst *ebiten.Image) {
	if c.err != nil || c.checked >= len(jellyCheckFrames) || c.frame != jellyCheckFrames[c.checked] {
		return
	}
	c.actual.Clear()
	c.expected.Clear()
	c.cube.Draw(c.actual)
	c.old.stCanvas = c.expected
	c.old.draw3DCube()
	c.actual.ReadPixels(c.pixelsA)
	c.expected.ReadPixels(c.pixelsB)
	if !bytes.Equal(c.pixelsA, c.pixelsB) {
		different := 0
		for i := 0; i < len(c.pixelsA); i += 4 {
			if !bytes.Equal(c.pixelsA[i:i+4], c.pixelsB[i:i+4]) {
				different++
			}
		}
		c.err = fmt.Errorf("jelly cube frame %d mode %d changed %d pixels", c.frame, c.old.rotationMode, different)
	}
	dst.DrawImage(c.actual, nil)
	c.checked++
}
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	dir, err := os.MkdirTemp("", "dck-jelly-")
	if err != nil {
		panic(err)
	}
	var check *jellyRenderCheck
	err = capture.Run(capture.Config{Directory: dir, Frames: jellyCheckFrames, Width: 640, Height: 400}, func() (ebiten.Game, error) {
		cube, err := NewJellyCube(DefaultJellyCubeConfig())
		if err != nil {
			return nil, err
		}
		old := newJellyLegacy()
		old.scrollIteration = 100
		old.zoom3d = 1
		old.whiteImg = cube.texture
		check = &jellyRenderCheck{cube: cube, old: old, actual: render.NewSurface(640, 400), expected: render.NewSurface(640, 400), pixelsA: make([]byte, 640*400*4), pixelsB: make([]byte, 640*400*4)}
		return check, nil
	})
	if err == nil && check != nil {
		err = check.err
		if err == nil && check.checked != len(jellyCheckFrames) {
			err = fmt.Errorf("only %d cube captures checked", check.checked)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("All %d complete cube captures match exactly: %s\n", check.checked, dir)
}
