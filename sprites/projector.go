// Package sprites provides configurable projected sprite fields and vectorballs.
package sprites

import (
	"cmp"
	"image/color"
	"reflect"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

type Point struct {
	X, Y, Z float64
	Image   int
}

// IndexedPointReader lets a mutable geometry.PointSequence render directly
// without copying its coordinates and sprite indices into a second point bank.
type IndexedPointReader interface {
	geometry.PointReader
	ImageIndex(index int) int
}

// Projection keeps every camera/sign/depth convention explicit. Matrix includes
// model scale, as in the original vectorball demos. A zero Matrix is identity.
type Projection struct {
	Matrix                           [9]float64
	Translate                        Point
	Focal, CenterX, CenterY          float64
	YUp, AscendingDepth, ScaleImages bool
	// Batch merges consecutive sorted sprites that share one image. Custom
	// DrawImageOptions retain the individual DrawImage path.
	Batch bool
	// CullPositiveModelZ drops the far half of a centered object after model
	// rotation and before camera translation. It is opt-in for opaque shells.
	CullPositiveModelZ bool
	Options            ebiten.DrawImageOptions
}
type projected struct {
	x, y, depth, scale float64
	image              int
}

// Projector reuses its depth buffer; source images and model points are borrowed.
type Projector struct {
	points []projected
	batch  *render.Batch
}

func (p *Projector) Draw(dst *ebiten.Image, points []Point, images []*ebiten.Image, c Projection) {
	if c.Focal <= 0 {
		return
	}
	m := p.prepare(len(points), c.Matrix)
	for _, v := range points {
		p.project(v, m, c)
	}
	p.drawProjected(dst, images, c)
}

// DrawIndexed projects authored point collections directly. A point sequence
// may mutate XYZ while keeping its original image indices, and several
// projectors can draw the same borrowed collection with different cameras.
func (p *Projector) DrawIndexed(dst *ebiten.Image, points IndexedPointReader, images []*ebiten.Image, c Projection) {
	if points == nil || c.Focal <= 0 {
		return
	}
	count := points.Len()
	m := p.prepare(count, c.Matrix)
	for index := 0; index < count; index++ {
		point := points.XYZ(index)
		p.project(Point{X: point.X, Y: point.Y, Z: point.Z, Image: points.ImageIndex(index)}, m, c)
	}
	p.drawProjected(dst, images, c)
}

func (p *Projector) prepare(count int, matrix [9]float64) [9]float64 {
	if matrix == ([9]float64{}) {
		matrix = [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}
	}
	if cap(p.points) < count {
		p.points = make([]projected, 0, count)
	} else {
		p.points = p.points[:0]
	}
	return matrix
}

func (p *Projector) project(v Point, m [9]float64, c Projection) {
	x := v.X*m[0] + v.Y*m[1] + v.Z*m[2]
	y := v.X*m[3] + v.Y*m[4] + v.Z*m[5]
	z := v.X*m[6] + v.Y*m[7] + v.Z*m[8]
	if c.CullPositiveModelZ && z > 0 {
		return
	}
	x += c.Translate.X
	y += c.Translate.Y
	z += c.Translate.Z
	scale := c.Focal / (c.Focal + z)
	screenY := c.CenterY + y*scale
	if c.YUp {
		screenY = c.CenterY - y*scale
	}
	p.points = append(p.points, projected{x: c.CenterX + x*scale, y: screenY, depth: z, scale: scale, image: v.Image})
}

func (p *Projector) drawProjected(dst *ebiten.Image, images []*ebiten.Image, c Projection) {
	slices.SortFunc(p.points, func(a, b projected) int {
		order := cmp.Compare(a.depth, b.depth)
		if !c.AscendingDepth {
			return -order
		}
		return order
	})
	if c.Batch && reflect.DeepEqual(c.Options, ebiten.DrawImageOptions{}) {
		p.drawBatched(dst, images, c.ScaleImages)
		return
	}
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

func (p *Projector) drawBatched(dst *ebiten.Image, images []*ebiten.Image, scaleImages bool) {
	if p.batch == nil {
		p.batch = render.NewBatch(8192)
	}
	var current *ebiten.Image
	for _, point := range p.points {
		if point.image < 0 || point.image >= len(images) || images[point.image] == nil {
			continue
		}
		img := images[point.image]
		if img != current {
			p.batch.Begin(dst, img)
			current = img
		}
		bounds := img.Bounds()
		width, height := float64(bounds.Dx()), float64(bounds.Dy())
		if scaleImages {
			width *= point.scale
			height *= point.scale
		}
		p.batch.Rect(point.x-width/2, point.y-height/2, width, height, bounds, color.White)
	}
	p.batch.Flush()
}
