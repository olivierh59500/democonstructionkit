package effects

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// PointCloud renders depth-sorted vectorballs, particle fields or 3D text sprites.
// Shape writes into a reusable point slice for morph/sine/rotor choreography.
type PointCloud struct {
	State
	Points    []geometry.Vec3
	Camera    geometry.Camera
	Transform Transform
	Animate   func(float64) Transform
	Shape     func([]geometry.Vec3, float64)
	Sprite    *ebiten.Image
	Size      float64
	Color     color.Color
	projected []projectedPoint
	batch     *render.Batch
}
type projectedPoint struct{ x, y, z, scale float64 }

func NewPointCloud(points []geometry.Vec3, sprite *ebiten.Image, camera geometry.Camera, size float64) (*PointCloud, error) {
	if len(points) == 0 || sprite == nil || size <= 0 || camera.Near <= 0 || camera.Focal <= 0 {
		return nil, fmt.Errorf("effects: invalid point cloud")
	}
	return &PointCloud{Points: append([]geometry.Vec3(nil), points...), Camera: camera, Transform: Transform{Scale: 1}, Sprite: sprite, Size: size, Color: color.White, projected: make([]projectedPoint, 0, len(points)), batch: render.NewBatch(2048)}, nil
}
func (p *PointCloud) Update(f kit.Frame) error {
	p.Frame = f
	if p.Shape != nil {
		p.Shape(p.Points, f.Time)
	}
	tr := p.Transform
	if p.Animate != nil {
		tr = p.Animate(f.Time)
	}
	r := geometry.RotateXYZ(tr.Rotation)
	p.projected = p.projected[:0]
	for _, v := range p.Points {
		v = r.Apply(v.Scale(tr.Scale)).Add(tr.Position)
		pos, scale, visible := p.Camera.Project(v)
		if visible {
			p.projected = append(p.projected, projectedPoint{x: pos.X, y: pos.Y, z: v.Z, scale: scale})
		}
	}
	sort.SliceStable(p.projected, func(i, j int) bool { return p.projected[i].z > p.projected[j].z })
	return nil
}
func (p *PointCloud) Draw(dst *ebiten.Image) {
	p.batch.Begin(dst, p.Sprite)
	bounds := p.Sprite.Bounds()
	aspect := float64(bounds.Dy()) / float64(bounds.Dx())
	for _, v := range p.projected {
		size := p.Size * v.scale
		p.batch.Rect(v.x-size/2, v.y-size*aspect/2, size, size*aspect, bounds, p.Color)
	}
	p.batch.Flush()
}

// Starfield is reproducible at any absolute time, including after a timeline loop.
type Starfield struct {
	State
	Width, Height       int
	Speed, Depth, Focal float64
	Center              geometry.Vec2
	Color               color.NRGBA
	stars               []geometry.Vec3
	white               *ebiten.Image
	batch               *render.Batch
}

func NewStarfield(width, height, count int, seed uint64) (*Starfield, error) {
	if width <= 0 || height <= 0 || count < 0 {
		return nil, fmt.Errorf("effects: invalid starfield")
	}
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	w := ebiten.NewImage(1, 1)
	w.Fill(color.White)
	s := &Starfield{Width: width, Height: height, Speed: 90, Depth: 800, Focal: 250, Center: geometry.Vec2{X: float64(width) / 2, Y: float64(height) / 2}, Color: color.NRGBA{200, 220, 255, 255}, white: w, batch: render.NewBatch(2048)}
	for i := 0; i < count; i++ {
		s.stars = append(s.stars, geometry.Vec3{X: (rng.Float64() - .5) * float64(width) * 4, Y: (rng.Float64() - .5) * float64(height) * 4, Z: rng.Float64()})
	}
	return s, nil
}
func (s *Starfield) Draw(dst *ebiten.Image) {
	if s.Depth <= 1 || s.Focal <= 0 {
		return
	}
	s.batch.Begin(dst, s.white)
	for _, p := range s.stars {
		z := 1 + motion.Wrap(p.Z*(s.Depth-1)-s.Frame.Time*s.Speed, s.Depth-1)
		scale := s.Focal / z
		x, y := s.Center.X+p.X*scale, s.Center.Y+p.Y*scale
		if x < 0 || y < 0 || x >= float64(s.Width) || y >= float64(s.Height) {
			continue
		}
		c := s.Color
		c.A = uint8(float64(c.A) * (1 - z/s.Depth))
		size := math.Max(1, math.Min(4, scale))
		s.batch.Rect(x, y, size, size, s.white.Bounds(), c)
	}
	s.batch.Flush()
}
func (s *Starfield) Close() error { s.white.Deallocate(); return nil }
