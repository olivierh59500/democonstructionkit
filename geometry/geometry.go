// Package geometry supplies renderer-independent 3D math and clipping.
package geometry

import "math"

type Vec2 struct{ X, Y float64 }
type Vec3 struct{ X, Y, Z float64 }

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}
func (a Vec3) Unit() Vec3 {
	n := math.Sqrt(a.Dot(a))
	if n == 0 {
		return Vec3{}
	}
	return a.Scale(1 / n)
}
func Lerp(a, b Vec3, t float64) Vec3 { return a.Add(b.Sub(a).Scale(t)) }

// Rotation is an XYZ Euler rotation precomputed once for a batch of points.
type Rotation [9]float64

func RotateXYZ(a Vec3) Rotation {
	sx, cx := math.Sincos(a.X)
	sy, cy := math.Sincos(a.Y)
	sz, cz := math.Sincos(a.Z)
	return Rotation{cz * cy, cz*sy*sx - sz*cx, cz*sy*cx + sz*sx, sz * cy, sz*sy*sx + cz*cx, sz*sy*cx - cz*sx, -sy, cy * sx, cy * cx}
}
func (m Rotation) Apply(p Vec3) Vec3 {
	return Vec3{m[0]*p.X + m[1]*p.Y + m[2]*p.Z, m[3]*p.X + m[4]*p.Y + m[5]*p.Z, m[6]*p.X + m[7]*p.Y + m[8]*p.Z}
}

// Camera looks along positive Z. Screen Y increases downward.
type Camera struct {
	Center      Vec2
	Focal, Near float64
}

func (c Camera) Project(p Vec3) (Vec2, float64, bool) {
	if c.Focal <= 0 || c.Near <= 0 || p.Z < c.Near || math.IsNaN(p.Z) {
		return Vec2{}, 0, false
	}
	s := c.Focal / p.Z
	return Vec2{c.Center.X + p.X*s, c.Center.Y + p.Y*s}, s, true
}

// Vertex carries texture coordinates in pixels through the clipping stage.
type Vertex struct {
	Position Vec3
	UV       Vec2
}

// ClipNear appends a clipped polygon (0–4 vertices). Supply a reusable buffer.
func ClipNear(dst []Vertex, triangle [3]Vertex, near float64) []Vertex {
	previous := triangle[2]
	for _, current := range triangle {
		inside, wasInside := current.Position.Z >= near, previous.Position.Z >= near
		if inside != wasInside {
			t := (near - previous.Position.Z) / (current.Position.Z - previous.Position.Z)
			v := Vertex{Position: Lerp(previous.Position, current.Position, t), UV: Vec2{previous.UV.X + (current.UV.X-previous.UV.X)*t, previous.UV.Y + (current.UV.Y-previous.UV.Y)*t}}
			v.Position.Z = near
			dst = append(dst, v)
		}
		if inside {
			dst = append(dst, current)
		}
		previous = current
	}
	return dst
}

// Morph writes interpolated points. Unequal shapes are resampled cyclically;
// callers choose a stable output count (usually max(len(from), len(to))).
func Morph(dst, from, to []Vec3, t float64) {
	if len(from) == 0 || len(to) == 0 {
		clear(dst)
		return
	}
	t = math.Max(0, math.Min(1, t))
	for i := range dst {
		dst[i] = Lerp(from[i%len(from)], to[i%len(to)], t)
	}
}
