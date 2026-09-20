package composite

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Feedback owns two surfaces for accumulated image effects. Shift never reads
// from the image being written; callers can then inject any additional layers.
type Feedback struct {
	front, back *ebiten.Image
	size        image.Point
}

func NewFeedback(size image.Point) (*Feedback, error) {
	if size.X < 1 || size.Y < 1 {
		return nil, fmt.Errorf("composite: invalid feedback dimensions")
	}
	op := &ebiten.NewImageOptions{Unmanaged: true}
	return &Feedback{front: ebiten.NewImageWithOptions(image.Rectangle{Max: size}, op), back: ebiten.NewImageWithOptions(image.Rectangle{Max: size}, op), size: size}, nil
}

// Image is borrowed until the next Shift or Close; do not retain subimages.
func (f *Feedback) Image() *ebiten.Image { return f.front }
func (f *Feedback) Shift(dx, dy int, wrapX, wrapY bool) {
	if f.front == nil {
		return
	}
	f.back.Clear()
	xs, ys := []int{dx}, []int{dy}
	if wrapX {
		dx = ((dx % f.size.X) + f.size.X) % f.size.X
		xs = []int{dx, dx - f.size.X}
	}
	if wrapY {
		dy = ((dy % f.size.Y) + f.size.Y) % f.size.Y
		ys = []int{dy, dy - f.size.Y}
	}
	for _, y := range ys {
		for _, x := range xs {
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			f.back.DrawImage(f.front, &op)
		}
	}
	f.front, f.back = f.back, f.front
}
func (f *Feedback) Close() error {
	if f.front != nil {
		f.front.Deallocate()
		f.back.Deallocate()
		f.front, f.back = nil, nil
	}
	return nil
}
