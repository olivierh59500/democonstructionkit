// Package ebiten captures original game implementations at deterministic ticks.
package ebiten

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

type Config struct {
	Directory     string
	Frames        []int
	Width, Height int
}
type capture struct {
	config        Config
	factory       func() (ebiten.Game, error)
	game          ebiten.Game
	surface       *ebiten.Image
	frame, index  int
	pending, done bool
	err           error
}

// Run constructs the game inside Ebitengine. Factories disable device audio and
// seed randomness identically for the reference and candidate, without changing
// the rendering code being compared. Every simulated Update is followed by Draw.
func Run(config Config, factory func() (ebiten.Game, error)) error {
	if config.Directory == "" || len(config.Frames) == 0 || config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("capture: invalid configuration")
	}
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		return err
	}
	ebiten.SetWindowSize(config.Width, config.Height)
	ebiten.SetWindowTitle("Original demo fidelity capture")
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(false)
	c := &capture{config: config, factory: factory}
	if err := ebiten.RunGame(c); err != nil {
		return err
	}
	return c.err
}
func (c *capture) Layout(int, int) (int, int) { return c.config.Width, c.config.Height }
func (c *capture) Update() error {
	if c.done {
		return ebiten.Termination
	}
	if c.game == nil {
		var err error
		c.game, err = c.factory()
		if err != nil {
			return err
		}
		c.game.Layout(c.config.Width, c.config.Height)
		c.surface = ebiten.NewImageWithOptions(image.Rect(0, 0, c.config.Width, c.config.Height), &ebiten.NewImageOptions{Unmanaged: true})
		c.pending = true
	}
	return nil
}
func (c *capture) Draw(screen *ebiten.Image) {
	if c.game == nil || c.done {
		return
	}
	for i := 0; i < 30 && !c.done; i++ {
		if !c.pending {
			if err := c.game.Update(); err != nil {
				c.err = err
				c.done = true
				return
			}
			c.frame++
		}
		c.pending = false
		c.surface.Clear()
		c.game.Draw(c.surface)
		if c.frame == c.config.Frames[c.index] {
			img := image.NewRGBA(c.surface.Bounds())
			c.surface.ReadPixels(img.Pix)
			name := filepath.Join(c.config.Directory, fmt.Sprintf("%06d.png", c.frame))
			f, err := os.Create(name)
			if err == nil {
				err = png.Encode(f, img)
				closeErr := f.Close()
				if err == nil {
					err = closeErr
				}
			}
			if err != nil {
				c.err = err
				c.done = true
				return
			}
			c.index++
			if c.index == len(c.config.Frames) {
				c.done = true
			}
		}
	}
	screen.DrawImage(c.surface, nil)
}
