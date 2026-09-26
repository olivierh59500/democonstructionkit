package scrolling

import "github.com/hajimehoshi/ebiten/v2"

// Draw retains a scene's separately configured filmstrip art and row profile.
func (p *SliceProgram) Draw(dst *ebiten.Image, film *DNAFrames, config DNADrawConfig) {
	if film != nil {
		film.DrawSlices(dst, p.stream.Slices(), p.stream.Head(), config)
	}
}
