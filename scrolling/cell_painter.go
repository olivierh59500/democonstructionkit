package scrolling

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

type CellShape uint8

const (
	CellFlat CellShape = iota
	CellCuboid
)

// FlatCellConfig sizes cells in glyph coordinates. ScreenGap removes constant
// destination pixels after mapping. Rectangles retains axis-aligned legacy
// rectangle sampling; otherwise all four corners follow the glyph's GeoM.
type FlatCellConfig struct {
	Size, ScreenGap geometry.Vec2
	Rectangles      bool
}

// CuboidCellConfig describes one reusable voxel, independently of font size.
// Vertices 0..3 have positive local Y, vertices 4..7 negative local Y.
// Colors may contain eight per-vertex colors; nil selects white. FlipY converts
// upward-positive source coordinates to DCK's downward-positive screen axis.
// Cells intersecting the camera's near plane are conservatively discarded.
type CuboidCellConfig struct {
	Size                         geometry.Vec3
	Camera                       geometry.Camera
	Origin                       geometry.Vec3
	Colors                       []color.NRGBA
	FlipY, IgnoreMappedY         bool
	ShowRearFaces, UnsortedFaces bool
}

// CellPose can rotate, scale, hide or reposition any individual voxel. Scale's
// zero value means one; Hidden explicitly suppresses a cell.
type CellPose struct {
	Center, Rotation geometry.Vec3
	Scale            float64
	Hidden           bool
}

type CellSample struct {
	Sample
	Cell font.Cell
	X, Y float64
}

// CellPainterConfig is usable directly or through Mode.Cells in scrolling.New.
// Fonts have independent cell dimensions, character sets and colors. Rows is a
// pure function of font name, row and time; its values are cached once per time
// sample. InvalidateRows refreshes it after changing captured parameters.
// Pose optionally edits the cuboid center/rotation per lit cell.
type CellPainterConfig struct {
	Fonts          map[string]*font.CellBank
	Font           string
	Shape          CellShape
	Flat           FlatCellConfig
	Cuboid         CuboidCellConfig
	Rows           func(fontName string, row int, seconds float64) geometry.Vec3
	Pose           func(CellSample, CellPose) CellPose
	Color          *color.NRGBA
	OutlineColor   *color.NRGBA
	LineWidth      float64
	Outlined       bool
	Wireframe      func(Sample) bool
	CullBounds     image.Rectangle // Empty uses destination bounds; this limits culling, not raster clipping.
	Blend          ebiten.Blend
	White          *ebiten.Image
	BatchTriangles int
}

type cellFontRows struct {
	bank  *font.CellBank
	rows  []geometry.Vec3
	time  float64
	valid bool
}

// CellPainter owns bounded geometry buffers and an optional white texture.
// Compiled fonts and an explicitly supplied texture remain caller-owned.
// Paint never scans an atlas or reads GPU pixels; drawing does not advance time.
type CellPainter struct {
	config      CellPainterConfig
	fonts       map[string]*cellFontRows
	defaultFont string
	white       *ebiten.Image
	ownWhite    bool
	batch       *render.Batch
	corners     [8]geometry.Vec3
	colors      [8]ebiten.Vertex
	flat, line  ebiten.Vertex
	outlined    bool
}

var cellFaces = [6][6]int{
	{0, 3, 1, 1, 3, 2}, {0, 1, 4, 1, 5, 4}, {1, 2, 5, 2, 6, 5},
	{7, 2, 3, 7, 6, 2}, {0, 7, 3, 4, 7, 0}, {4, 6, 7, 4, 5, 6},
}
var cellFaceEdges = [6][4]int{
	{0, 3, 2, 1}, {0, 1, 5, 4}, {1, 2, 6, 5},
	{7, 6, 2, 3}, {0, 4, 7, 3}, {4, 5, 6, 7},
}

func NewCellPainter(c CellPainterConfig) (*CellPainter, error) {
	if len(c.Fonts) == 0 || len(c.Fonts) > 256 || c.Shape > CellCuboid {
		return nil, fmt.Errorf("scrolling: invalid cell font bank or shape")
	}
	if c.LineWidth == 0 {
		c.LineWidth = 1
	}
	if c.BatchTriangles == 0 {
		c.BatchTriangles = 4096
	}
	if !finite(c.LineWidth) || c.LineWidth <= 0 || c.BatchTriangles < 2 || c.BatchTriangles > 20000 {
		return nil, fmt.Errorf("scrolling: invalid cell outline or triangle budget")
	}
	if c.White != nil && c.White.Bounds() != image.Rect(0, 0, 1, 1) {
		return nil, fmt.Errorf("scrolling: cell texture must be one white pixel at zero origin")
	}
	if c.Flat.Size == (geometry.Vec2{}) {
		c.Flat.Size = geometry.Vec2{X: 1, Y: 1}
	}
	for _, value := range []float64{c.Flat.Size.X, c.Flat.Size.Y, c.Flat.ScreenGap.X, c.Flat.ScreenGap.Y,
		c.Cuboid.Size.X, c.Cuboid.Size.Y, c.Cuboid.Size.Z, c.Cuboid.Origin.X, c.Cuboid.Origin.Y, c.Cuboid.Origin.Z,
		c.Cuboid.Camera.Center.X, c.Cuboid.Camera.Center.Y, c.Cuboid.Camera.Focal, c.Cuboid.Camera.Near} {
		if !finite(value) {
			return nil, fmt.Errorf("scrolling: nonfinite cell geometry")
		}
	}
	if c.Flat.Size.X <= 0 || c.Flat.Size.Y <= 0 || c.Flat.ScreenGap.X < 0 || c.Flat.ScreenGap.Y < 0 {
		return nil, fmt.Errorf("scrolling: invalid flat cell dimensions")
	}
	if c.Shape == CellCuboid && (c.Cuboid.Size.X <= 0 || c.Cuboid.Size.Y <= 0 || c.Cuboid.Size.Z <= 0 ||
		c.Cuboid.Camera.Focal <= 0 || c.Cuboid.Camera.Near <= 0 || len(c.Cuboid.Colors) != 0 && len(c.Cuboid.Colors) != 8) {
		return nil, fmt.Errorf("scrolling: invalid cuboid cell dimensions, camera or colors")
	}
	p := &CellPainter{config: c, fonts: make(map[string]*cellFontRows), defaultFont: c.Font, outlined: c.Outlined}
	rows := 0
	for name, bank := range c.Fonts {
		if bank == nil || bank.Height() <= 0 {
			return nil, fmt.Errorf("scrolling: missing cell font %q", name)
		}
		rows += bank.Height()
		if rows > 65536 {
			return nil, fmt.Errorf("scrolling: excessive cell row cache")
		}
		p.fonts[name] = &cellFontRows{bank: bank, rows: make([]geometry.Vec3, bank.Height())}
		if p.defaultFont == "" && len(c.Fonts) == 1 {
			p.defaultFont = name
		}
	}
	if p.defaultFont == "" {
		p.defaultFont = "default"
	}
	if _, exists := p.fonts[p.defaultFont]; !exists {
		return nil, fmt.Errorf("scrolling: missing default cell font %q", p.defaultFont)
	}
	paint, line := color.NRGBA{255, 255, 255, 255}, color.NRGBA{255, 255, 255, 255}
	if c.Color != nil {
		paint = *c.Color
	}
	if c.OutlineColor != nil {
		line = *c.OutlineColor
	}
	p.flat, p.line = render.Vertex(0, 0, 0, 0, paint), render.Vertex(0, 0, 0, 0, line)
	for i := range p.colors {
		vertexPaint := paint
		if len(c.Cuboid.Colors) == 8 {
			vertexPaint = c.Cuboid.Colors[i]
		}
		p.colors[i] = render.Vertex(0, 0, 0, 0, vertexPaint)
	}
	h := c.Cuboid.Size.Scale(0.5)
	p.corners = [8]geometry.Vec3{{X: -h.X, Y: h.Y, Z: -h.Z}, {X: h.X, Y: h.Y, Z: -h.Z}, {X: h.X, Y: h.Y, Z: h.Z}, {X: -h.X, Y: h.Y, Z: h.Z},
		{X: -h.X, Y: -h.Y, Z: -h.Z}, {X: h.X, Y: -h.Y, Z: -h.Z}, {X: h.X, Y: -h.Y, Z: h.Z}, {X: -h.X, Y: -h.Y, Z: h.Z}}
	p.white = c.White
	if p.white == nil {
		p.white = ebiten.NewImage(1, 1)
		p.white.Fill(color.White)
		p.ownWhite = true
	}
	p.batch = render.NewBatch(c.BatchTriangles)
	p.batch.Options.Blend = c.Blend
	// Keep only immutable scalar options and callbacks in the stored configuration.
	p.config.Fonts, p.config.Cuboid.Colors = nil, nil
	p.config.Color, p.config.OutlineColor = nil, nil
	return p, nil
}

func (p *CellPainter) SetOutlined(value bool) { p.outlined = value }
func (p *CellPainter) InvalidateRows() {
	for _, f := range p.fonts {
		f.valid = false
	}
}

func cellColor(base ebiten.Vertex, cell font.Cell, scale ebiten.ColorScale) ebiten.Vertex {
	if cell.Color != (color.NRGBA{255, 255, 255, 255}) {
		r, g, b, a := cell.Color.RGBA()
		base.ColorR *= float32(r) / 65535
		base.ColorG *= float32(g) / 65535
		base.ColorB *= float32(b) / 65535
		base.ColorA *= float32(a) / 65535
	}
	base.ColorR *= scale.R()
	base.ColorG *= scale.G()
	base.ColorB *= scale.B()
	base.ColorA *= scale.A()
	return base
}

func cellVertex(base ebiten.Vertex, x, y, u, v float64) ebiten.Vertex {
	base.DstX, base.DstY, base.SrcX, base.SrcY = float32(x), float32(y), float32(u), float32(v)
	return base
}

func (p *CellPainter) rect(base ebiten.Vertex, x, y, w, h float64) {
	p.batch.Quad([4]ebiten.Vertex{cellVertex(base, x, y, 0, 0), cellVertex(base, x+w, y, 1, 0), cellVertex(base, x+w, y+h, 1, 1), cellVertex(base, x, y+h, 0, 1)})
}

// Paint implements scrolling.Painter and therefore works with the same font,
// control, repeat and mapper pipeline as ordinary scrolling.
func (p *CellPainter) Paint(dst *ebiten.Image, sample Sample, op ebiten.DrawImageOptions) {
	if p == nil || p.white == nil || dst == nil || !finite(sample.Time) {
		return
	}
	name := sample.Glyph.Font
	if name == "" {
		name = p.defaultFont
	}
	f := p.fonts[name]
	if f == nil {
		return
	}
	cells := f.bank.Glyph(sample.Glyph.Rune)
	if len(cells) == 0 {
		return
	}
	if !f.valid || f.time != sample.Time {
		for row := range f.rows {
			f.rows[row] = geometry.Vec3{}
			if p.config.Rows != nil {
				f.rows[row] = p.config.Rows(name, row, sample.Time)
			}
		}
		f.valid, f.time = true, sample.Time
	}
	outlined := p.outlined
	if p.config.Wireframe != nil {
		outlined = p.config.Wireframe(sample)
	}
	p.batch.Begin(dst, p.white)
	for _, cell := range cells {
		shift := f.rows[cell.Y]
		if p.config.Shape == CellFlat {
			p.paintFlat(cell, shift, op, outlined)
			continue
		}
		x, y := op.GeoM.Apply(float64(cell.X), float64(cell.Y))
		if p.config.Cuboid.IgnoreMappedY {
			y = 0
		}
		pose := CellPose{Center: geometry.Vec3{X: x + p.config.Cuboid.Origin.X + shift.X, Y: y + p.config.Cuboid.Origin.Y + shift.Y, Z: p.config.Cuboid.Origin.Z + shift.Z}, Scale: 1}
		if p.config.Pose != nil {
			pose = p.config.Pose(CellSample{Sample: sample, Cell: cell, X: x, Y: y}, pose)
		}
		p.paintCuboid(dst, cell, pose, op.ColorScale, outlined)
	}
	p.batch.Flush()
}

func (p *CellPainter) paintFlat(cell font.Cell, shift geometry.Vec3, op ebiten.DrawImageOptions, outline bool) {
	c := p.config.Flat
	x, y := op.GeoM.Apply(float64(cell.X), float64(cell.Y))
	x1, y1 := op.GeoM.Apply(float64(cell.X)+c.Size.X, float64(cell.Y)+c.Size.Y)
	base := cellColor(p.flat, cell, op.ColorScale)
	if c.Rectangles {
		w, h := x1-x-c.ScreenGap.X, y1-y-c.ScreenGap.Y
		x, y = x+shift.X, y+shift.Y
		if outline {
			line := p.config.LineWidth
			if w <= 2*line || h <= 2*line {
				return
			}
			base = cellColor(p.line, cell, op.ColorScale)
			p.rect(base, x, y, w, line)
			p.rect(base, x, y+h-line, w, line)
			p.rect(base, x, y+line, line, h-2*line)
			p.rect(base, x+w-line, y+line, line, h-2*line)
		} else {
			p.rect(base, x, y, w, h)
		}
		return
	}
	x2, y2 := op.GeoM.Apply(float64(cell.X)+c.Size.X, float64(cell.Y))
	x3, y3 := op.GeoM.Apply(float64(cell.X), float64(cell.Y)+c.Size.Y)
	points := [4]geometry.Vec2{{X: x + shift.X, Y: y + shift.Y}, {X: x2 + shift.X - c.ScreenGap.X, Y: y2 + shift.Y},
		{X: x1 + shift.X - c.ScreenGap.X, Y: y1 + shift.Y - c.ScreenGap.Y}, {X: x3 + shift.X, Y: y3 + shift.Y - c.ScreenGap.Y}}
	if outline {
		base = cellColor(p.line, cell, op.ColorScale)
		for i := range points {
			p.stroke(base, points[i], points[(i+1)%4])
		}
	} else {
		p.batch.Quad([4]ebiten.Vertex{cellVertex(base, points[0].X, points[0].Y, 0, 0), cellVertex(base, points[1].X, points[1].Y, 1, 0), cellVertex(base, points[2].X, points[2].Y, 1, 1), cellVertex(base, points[3].X, points[3].Y, 0, 1)})
	}
}

type visibleCellFace struct {
	index int
	depth float64
}

func (p *CellPainter) paintCuboid(dst *ebiten.Image, cell font.Cell, pose CellPose, scale ebiten.ColorScale, outline bool) {
	if pose.Hidden {
		return
	}
	if pose.Scale == 0 {
		pose.Scale = 1
	}
	for _, v := range []float64{pose.Center.X, pose.Center.Y, pose.Center.Z, pose.Rotation.X, pose.Rotation.Y, pose.Rotation.Z, pose.Scale} {
		if !finite(v) {
			return
		}
	}
	c := p.config.Cuboid
	rotation := geometry.Rotation{}
	rotated := pose.Rotation != (geometry.Vec3{})
	if rotated {
		rotation = geometry.RotateXYZ(pose.Rotation)
	}
	var world [8]geometry.Vec3
	var screen [8]geometry.Vec2
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for i, corner := range p.corners {
		if pose.Scale != 1 {
			corner = corner.Scale(pose.Scale)
		}
		if rotated {
			corner = rotation.Apply(corner)
		}
		world[i] = pose.Center.Add(corner)
		v := world[i]
		if c.FlipY {
			v.Y = -v.Y
		}
		var valid bool
		screen[i], _, valid = c.Camera.Project(v)
		if !valid {
			return
		}
		minX, minY, maxX, maxY = min(minX, screen[i].X), min(minY, screen[i].Y), max(maxX, screen[i].X), max(maxY, screen[i].Y)
	}
	bounds := p.config.CullBounds
	if bounds.Empty() {
		bounds = dst.Bounds()
	}
	if maxX < float64(bounds.Min.X) || minX >= float64(bounds.Max.X) || maxY < float64(bounds.Min.Y) || minY >= float64(bounds.Max.Y) {
		return
	}
	var faces [6]visibleCellFace
	count := 0
	for face, indices := range cellFaces {
		a, b, d := world[indices[0]], world[indices[1]], world[indices[2]]
		if !c.ShowRearFaces && b.Sub(a).Cross(d.Sub(a)).Dot(a) >= 0 {
			continue
		}
		edges := cellFaceEdges[face]
		depth := (world[edges[0]].Z + world[edges[1]].Z + world[edges[2]].Z + world[edges[3]].Z) / 4
		at := count
		if !c.UnsortedFaces {
			for at > 0 && faces[at-1].depth < depth {
				faces[at] = faces[at-1]
				at--
			}
		}
		faces[at] = visibleCellFace{face, depth}
		count++
	}
	for _, face := range faces[:count] {
		if outline {
			base := cellColor(p.line, cell, scale)
			edges := cellFaceEdges[face.index]
			for i := 0; i < 4; i++ {
				p.stroke(base, screen[edges[i]], screen[edges[(i+1)%4]])
			}
			continue
		}
		indices := cellFaces[face.index]
		for triangle := 0; triangle < 6; triangle += 3 {
			var vertices [3]ebiten.Vertex
			for corner := 0; corner < 3; corner++ {
				index := indices[triangle+corner]
				v := screen[index]
				vertices[corner] = cellVertex(cellColor(p.colors[index], cell, scale), v.X, v.Y, 0, 0)
			}
			p.batch.Triangle(vertices[0], vertices[1], vertices[2])
		}
	}
}

func (p *CellPainter) stroke(base ebiten.Vertex, a, b geometry.Vec2) {
	dx, dy := b.X-a.X, b.Y-a.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	half := p.config.LineWidth / 2
	nx, ny := -dy/length*half, dx/length*half
	p.batch.Quad([4]ebiten.Vertex{cellVertex(base, a.X+nx, a.Y+ny, 0, 0), cellVertex(base, b.X+nx, b.Y+ny, 1, 0), cellVertex(base, b.X-nx, b.Y-ny, 1, 1), cellVertex(base, a.X-nx, a.Y-ny, 0, 1)})
}

func (p *CellPainter) Close() error {
	if p != nil {
		if p.ownWhite && p.white != nil {
			p.white.Deallocate()
		}
		p.white, p.batch, p.fonts = nil, nil, nil
		p.config = CellPainterConfig{}
	}
	return nil
}
