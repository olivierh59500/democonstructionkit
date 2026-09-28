//go:build dck_path_entrycheck

package scrolling

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type pathEntryCheck struct {
	scroll  *Scrolling
	glyph   *ebiten.Image
	checked bool
	err     error
}

func (c *pathEntryCheck) Layout(int, int) (int, int) { return 24, 16 }
func (c *pathEntryCheck) Update() error              { return nil }
func (c *pathEntryCheck) Draw(*ebiten.Image) {
	if c.checked {
		return
	}
	c.checked = true
	output := ebiten.NewImage(24, 16)
	defer output.Deallocate()
	for distance := -8; distance <= 16; distance++ {
		output.Clear()
		state := IdentityState()
		state.X, state.Shape = float64(distance), "path"
		c.scroll.DrawAt(output, state)
		pixels := make([]byte, 24*16*4)
		output.ReadPixels(pixels)
		visible := 0
		for i := 3; i < len(pixels); i += 4 {
			if pixels[i] != 0 {
				visible++
			}
		}
		left, right := max(4, 4+distance), min(20, 12+distance)
		want := max(0, right-left) * 4
		if visible != want {
			c.err = fmt.Errorf("path entry at distance %d draws %d pixels, want %d", distance, visible, want)
			return
		}
	}
}

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory, err := os.MkdirTemp("", "dck-path-entry-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *pathEntryCheck
	err = capture.Run(capture.Config{Directory: directory, Frames: []int{0}, Width: 24, Height: 16}, func() (ebiten.Game, error) {
		path, err := motion.NewPolyline([]motion.Point{{X: 4, Y: 4}, {X: 20, Y: 4}}, false)
		if err != nil {
			return nil, err
		}
		mode, err := AlongPath(PathConfig{Path: path, Extrapolate: true, Viewport: image.Rect(4, 0, 20, 16)})
		if err != nil {
			return nil, err
		}
		glyph := ebiten.NewImage(8, 4)
		glyph.Fill(color.White)
		scroll, err := New(Config{Glyphs: []Glyph{{Image: glyph, Rune: 'A', Advance: 8}},
			Modes: map[string]Mode{"path": mode}, Shape: "path"})
		if err != nil {
			glyph.Deallocate()
			return nil, err
		}
		check = &pathEntryCheck{scroll: scroll, glyph: glyph}
		return check, nil
	})
	if check != nil {
		_ = check.scroll.Close()
		check.glyph.Deallocate()
		if err == nil {
			err = check.err
		}
		if err == nil && !check.checked {
			err = fmt.Errorf("path entry was not drawn")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
