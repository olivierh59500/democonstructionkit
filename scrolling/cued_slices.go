package scrolling

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// ErrSliceFilmUnbound reports drawing a cued strip transport before its
// borrowed film has been attached, or after the owner has closed that film.
var ErrSliceFilmUnbound = errors.New("scrolling: cued slice artwork is not available")

type cuedSlicesTransport struct {
	program *SliceProgram
	film    *DNAFrames
	draw    DNADrawConfig
	err     error
}

func (c *cuedSlicesTransport) bindFilm(film *DNAFrames) error {
	if film == nil || film.Image == nil || film.Count != c.program.Clock().RotationFrames() {
		return fmt.Errorf("scrolling: cued slice film and rotation clock differ")
	}
	c.film, c.err = film, nil
	return nil
}

func (c *cuedSlicesTransport) Update(kit.Frame) error { return c.program.Step() }

func (c *cuedSlicesTransport) Draw(dst *ebiten.Image) {
	c.err = nil
	if c.film == nil || c.film.Image == nil {
		c.err = ErrSliceFilmUnbound
		return
	}
	c.program.Draw(dst, c.film, c.draw)
}

func (c *cuedSlicesTransport) DrawError() error { return c.err }
