package democonstructionkit

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// ImagePass transforms a live layer into a complete output image. An overlay
// callback first draws source, then its lens/reflection/etc. Never retain the
// borrowed images. Enabled=nil means always on. Close is optional ownership.
type ImagePass struct {
	Apply   func(dst, source *ebiten.Image, frame Frame)
	Enabled func(Frame) bool
	Close   func() error
}
type Pipeline struct {
	source Effect
	passes []ImagePass
	active []bool
	images [2]*ebiten.Image
	frame  Frame
	count  int
	closed bool
}

func NewPipeline(source Effect, width, height int, passes ...ImagePass) (*Pipeline, error) {
	if source == nil || width < 1 || height < 1 {
		return nil, fmt.Errorf("kit: invalid image pipeline")
	}
	for _, p := range passes {
		if p.Apply == nil {
			return nil, fmt.Errorf("kit: nil image pass")
		}
	}
	p := &Pipeline{source: source, passes: append([]ImagePass(nil), passes...), active: make([]bool, len(passes))}
	p.images = [2]*ebiten.Image{render.NewSurface(width, height), render.NewSurface(width, height)}
	return p, nil
}
func (p *Pipeline) Update(f Frame) error {
	if p.closed {
		return nil
	}
	p.frame = f
	p.count = 0
	for i, pass := range p.passes {
		p.active[i] = pass.Enabled == nil || pass.Enabled(f)
		if p.active[i] {
			p.count++
		}
	}
	return p.source.Update(f)
}
func (p *Pipeline) Draw(dst *ebiten.Image) {
	if p.closed || dst == nil {
		return
	}
	if p.count == 0 {
		p.source.Draw(dst)
		return
	}
	front, back := p.images[0], p.images[1]
	front.Clear()
	p.source.Draw(front)
	for i, pass := range p.passes {
		if !p.active[i] {
			continue
		}
		back.Clear()
		pass.Apply(back, front, p.frame)
		front, back = back, front
	}
	dst.DrawImage(front, nil)
}
func (p *Pipeline) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	for _, img := range p.images {
		img.Deallocate()
	}
	err := Close(p.source)
	for _, pass := range p.passes {
		if pass.Close != nil {
			err = errors.Join(err, pass.Close())
		}
	}
	return err
}
