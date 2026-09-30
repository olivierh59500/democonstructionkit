//go:build dck_gpu_rendercheck

package sprites

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestDepthQueueGPUKeepsVariableFrameHeightsOrderClippingAndBorrowedAtlas(t *testing.T) {
	testutil.RequireGPU(t)
	frames := []image.Rectangle{image.Rect(0, 0, 3, 4), image.Rect(3, 0, 6, 2), image.Rect(6, 0, 9, 1)}
	paints := []color.NRGBA{{R: 200, A: 128}, {G: 200, A: 128}, {B: 200, A: 128}}
	art := ebiten.NewImage(9, 4)
	defer art.Deallocate()
	for i, frame := range frames {
		art.SubImage(frame).(*ebiten.Image).Fill(paints[i])
	}
	calls := 0
	queue, err := NewDepthQueue(DepthQueueConfig{Points: []geometry.Vec2{{X: 1, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 1}},
		Depth: 30, Near: 25, Far: 30, Spacing: 5,
		Project: func(slot int, p geometry.Vec2, depth int) (FieldSample, bool) {
			calls++
			return FieldSample{X: p.X, Y: p.Y, Z: float64(depth), Scale: 1, Image: slot}, true
		}, Style: FieldStyle{Image: art, Frames: frames}})
	if err != nil {
		t.Fatal(err)
	}
	queue.Step(0)
	background := color.NRGBA{R: 20, G: 30, B: 40, A: 255}
	canvas := ebiten.NewImage(8, 7)
	defer canvas.Deallocate()
	clip := image.Rect(2, 0, 5, 6)
	view := canvas.SubImage(clip).(*ebiten.Image)
	canvas.Fill(background)
	queue.Draw(view)
	pixels := make([]byte, 8*7*4)
	canvas.ReadPixels(pixels)
	// Draw the three solid rectangles in queue order with an independent CPU
	// source-over implementation. Frame heights are 4,2,1, rather than uniform.
	want := image.NewRGBA(image.Rect(0, 0, 8, 7))
	draw.Draw(want, want.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)
	for i, frame := range frames {
		box := image.Rect(1, 1, 1+frame.Dx(), 1+frame.Dy()).Intersect(clip)
		draw.Draw(want, box, image.NewUniform(paints[i]), image.Point{}, draw.Over)
	}
	for y := 0; y < 7; y++ {
		for x := 0; x < 8; x++ {
			paint := want.RGBAAt(x, y)
			at := (y*8 + x) * 4
			for channel, expected := range []byte{paint.R, paint.G, paint.B, paint.A} {
				delta := int(pixels[at+channel]) - int(expected)
				if delta < -1 || delta > 1 {
					t.Fatalf("queue draw at(%d,%d) channel%d=%d want%d", x, y, channel, pixels[at+channel], expected)
				}
			}
		}
	}
	canvas.Fill(background)
	queue.Draw(view)
	second := make([]byte, len(pixels))
	canvas.ReadPixels(second)
	for i, value := range pixels {
		if second[i] != value {
			t.Fatal("repeated draw changed a prepared population")
		}
	}
	if calls != 3 {
		t.Fatal("drawing advanced projection or transport")
	}
	// The same prepared population can use the common outline skin.
	style := queue.Style
	style.Outline = &FieldOutline{Width: 1}
	canvas.Clear()
	queue.DrawStyle(view, style)
	canvas.ReadPixels(second)
	if second[(1*8+2)*4+3] != 255 || second[(3*8+2)*4+3] != 0 {
		t.Fatal("alternate skin did not retain cached frame dimensions")
	}
	queue.Close()
	queue.Close()
	art.Fill(color.NRGBA{G: 255, A: 255})
	canvas.Clear()
	canvas.DrawImage(art, nil)
	canvas.ReadPixels(second)
	if second[1] != 255 || second[3] != 255 {
		t.Fatal("queue closing disposed borrowed atlas artwork")
	}
	queue.Draw(view)
}
