package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

// Wireframe draws projected line segments with near-plane clipping.
type Wireframe struct {
	State
	Points    []geometry.Vec3
	Edges     [][2]int
	Camera    geometry.Camera
	Transform Transform
	Animate   func(float64) Transform
	Width     float64
	Color     color.Color
	white     *ebiten.Image
	projected []geometry.Vec3
	batch     *render.Batch
}

func NewWireframe(points []geometry.Vec3, edges [][2]int, camera geometry.Camera) (*Wireframe, error) {
	if camera.Near <= 0 || camera.Focal <= 0 {
		return nil, fmt.Errorf("effects: invalid wireframe camera")
	}
	for _, e := range edges {
		if e[0] < 0 || e[1] < 0 || e[0] >= len(points) || e[1] >= len(points) {
			return nil, fmt.Errorf("effects: invalid wireframe edge")
		}
	}
	w := ebiten.NewImage(1, 1)
	w.Fill(color.White)
	return &Wireframe{Points: append([]geometry.Vec3(nil), points...), Edges: append([][2]int(nil), edges...), Camera: camera, Transform: Transform{Scale: 1}, Width: 1, Color: color.White, white: w, projected: make([]geometry.Vec3, len(points)), batch: render.NewBatch(2048)}, nil
}
func (w *Wireframe) Update(f kit.Frame) error {
	w.Frame = f
	tr := w.Transform
	if w.Animate != nil {
		tr = w.Animate(f.Time)
	}
	r := geometry.RotateXYZ(tr.Rotation)
	for i, p := range w.Points {
		w.projected[i] = r.Apply(p.Scale(tr.Scale)).Add(tr.Position)
	}
	return nil
}
func (w *Wireframe) Draw(dst *ebiten.Image) {
	w.batch.Begin(dst, w.white)
	for _, e := range w.Edges {
		a, b := w.projected[e[0]], w.projected[e[1]]
		near := w.Camera.Near
		if a.Z < near && b.Z < near {
			continue
		}
		if a.Z < near {
			a = geometry.Lerp(a, b, (near-a.Z)/(b.Z-a.Z))
			a.Z = near
		}
		if b.Z < near {
			b = geometry.Lerp(b, a, (near-b.Z)/(a.Z-b.Z))
			b.Z = near
		}
		p, _, _ := w.Camera.Project(a)
		q, _, _ := w.Camera.Project(b)
		dx, dy := q.X-p.X, q.Y-p.Y
		length := math.Hypot(dx, dy)
		if length == 0 {
			continue
		}
		ox, oy := -dy*w.Width/(2*length), dx*w.Width/(2*length)
		w.batch.Quad([4]ebiten.Vertex{render.Vertex(p.X+ox, p.Y+oy, .5, .5, w.Color), render.Vertex(q.X+ox, q.Y+oy, .5, .5, w.Color), render.Vertex(q.X-ox, q.Y-oy, .5, .5, w.Color), render.Vertex(p.X-ox, p.Y-oy, .5, .5, w.Color)})
	}
	w.batch.Flush()
}
func (w *Wireframe) Close() error { w.white.Deallocate(); return nil }
