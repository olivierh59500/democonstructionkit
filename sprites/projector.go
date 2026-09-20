// Package sprites provides configurable projected sprite fields and vectorballs.
package sprites

import (
	"cmp"
	"github.com/hajimehoshi/ebiten/v2"
	"slices"
)

type Point struct {
	X, Y, Z float64
	Image   int
}

// Projection keeps every camera/sign/depth convention explicit. Matrix includes
// model scale, as in the original vectorball demos. A zero Matrix is identity.
type Projection struct {
	Matrix                           [9]float64
	Translate                        Point
	Focal, CenterX, CenterY          float64
	YUp, AscendingDepth, ScaleImages bool
	Options                          ebiten.DrawImageOptions
}
type projected struct {
	x, y, depth, scale float64
	image              int
}

// Projector reuses its depth buffer; source images and model points are borrowed.
type Projector struct{ points []projected }

func (p *Projector) Draw(dst *ebiten.Image, points []Point, images []*ebiten.Image, c Projection) {
	if c.Focal <= 0 {
		return
	}
	m := c.Matrix
	if m == ([9]float64{}) {
		m = [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}
	}
	if cap(p.points) < len(points) {
		p.points = make([]projected, len(points))
	} else {
		p.points = p.points[:len(points)]
	}
	for i, v := range points {
		x := v.X*m[0] + v.Y*m[1] + v.Z*m[2]
		y := v.X*m[3] + v.Y*m[4] + v.Z*m[5]
		z := v.X*m[6] + v.Y*m[7] + v.Z*m[8]
		x += c.Translate.X
		y += c.Translate.Y
		z += c.Translate.Z
		scale := c.Focal / (c.Focal + z)
		screenY := c.CenterY + y*scale
		if c.YUp {
			screenY = c.CenterY - y*scale
		}
		p.points[i] = projected{x: c.CenterX + x*scale, y: screenY, depth: z, scale: scale, image: v.Image}
	}
	slices.SortFunc(p.points, func(a, b projected) int {
		order := cmp.Compare(a.depth, b.depth)
		if !c.AscendingDepth {
			return -order
		}
		return order
	})
	for _, point := range p.points {
		if point.image < 0 || point.image >= len(images) || images[point.image] == nil {
			continue
		}
		img := images[point.image]
		op := c.Options
		op.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
		if c.ScaleImages {
			op.GeoM.Scale(point.scale, point.scale)
		}
		op.GeoM.Translate(point.x, point.y)
		dst.DrawImage(img, &op)
	}
}
