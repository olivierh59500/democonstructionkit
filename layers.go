package democonstructionkit

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// TimedLayer composes a child in slice order. LocalTime shifts Frame.Time to
// its activation; Tick remains the caller's global tick. Assets stay borrowed.
type TimedLayer struct {
	Effect    Effect
	Window    timeline.Window
	LocalTime bool
	Blend     ebiten.Blend
}
type Layers struct {
	layers []TimedLayer
	alpha  []float64
	active []bool
	canvas *ebiten.Image
	closed bool
}

func NewLayers(width, height int, layers ...TimedLayer) (*Layers, error) {
	if width < 1 || height < 1 || len(layers) == 0 {
		return nil, fmt.Errorf("kit: invalid timed layers")
	}
	for _, l := range layers {
		if l.Effect == nil {
			return nil, fmt.Errorf("kit: nil timed layer")
		}
		if err := l.Window.Validate(); err != nil {
			return nil, err
		}
	}
	return &Layers{layers: append([]TimedLayer(nil), layers...), alpha: make([]float64, len(layers)), active: make([]bool, len(layers)), canvas: render.NewSurface(width, height)}, nil
}

// SetWindow changes or retriggers one layer on the graphics goroutine.
func (l *Layers) SetWindow(index int, w timeline.Window) error {
	if index < 0 || index >= len(l.layers) {
		return fmt.Errorf("kit: layer index out of range")
	}
	if err := w.Validate(); err != nil {
		return err
	}
	l.layers[index].Window = w
	return nil
}
func (l *Layers) Update(frame Frame) error {
	if l.closed {
		return nil
	}
	for i, entry := range l.layers {
		local, alpha, on := entry.Window.At(frame.Time)
		l.alpha[i], l.active[i] = alpha, on
		if !on {
			continue
		}
		f := frame
		if entry.LocalTime {
			f.Time = local
		}
		if err := entry.Effect.Update(f); err != nil {
			return fmt.Errorf("layer %d: %w", i, err)
		}
	}
	return nil
}
func (l *Layers) Draw(dst *ebiten.Image) {
	if l.closed || dst == nil {
		return
	}
	for i, entry := range l.layers {
		if !l.active[i] || l.alpha[i] <= 0 {
			continue
		}
		if l.alpha[i] == 1 && (entry.Blend == (ebiten.Blend{}) || entry.Blend == ebiten.BlendSourceOver) {
			entry.Effect.Draw(dst)
			continue
		}
		l.canvas.Clear()
		entry.Effect.Draw(l.canvas)
		op := ebiten.DrawImageOptions{Blend: entry.Blend}
		op.ColorScale.ScaleAlpha(float32(l.alpha[i]))
		dst.DrawImage(l.canvas, &op)
	}
}
func (l *Layers) Close() error {
	if l.closed {
		return nil
	}
	l.closed = true
	l.canvas.Deallocate()
	var err error
	for _, entry := range l.layers {
		err = errors.Join(err, Close(entry.Effect))
	}
	return err
}
