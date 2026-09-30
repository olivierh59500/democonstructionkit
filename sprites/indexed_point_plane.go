package sprites

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

type PointPlaneCollision uint8

const (
	PointPlaneOR PointPlaneCollision = iota
	PointPlaneXOR
)

// PointPlaneSample is one integer pixel and its palette index or bit-plane mask.
type PointPlaneSample struct {
	X, Y int
	Mask byte
}

// IndexedPointPlaneConfig separates projection, pixel collisions and material.
// Sample is called once per point in ascending order, only when preparing a
// frame. Palette is copied; missing indices draw transparent. White optionally
// borrows a one-pixel white image. DrawZero paints touched pixels whose final
// mask is zero, including XOR cancellations, with Palette[0].
type IndexedPointPlaneConfig struct {
	Width, Height, Count int
	Sample               func(index int) (PointPlaneSample, bool)
	Collision            PointPlaneCollision
	Palette              []color.NRGBA
	Offset               image.Point
	DrawZero             bool
	White                *ebiten.Image
	Blend                ebiten.Blend
	BatchTriangles       int
}

// IndexedPointPlane owns a bounded sparse raster and its batched drawing state.
// Only touched pixels are cleared. A separate membership bitset deduplicates
// pixels even when an XOR cancellation temporarily returns their mask to zero.
// Sample and palette updates allocate nothing apart from work done by a caller's
// callback. Drawing reuses bounded geometry and lazily creates its optional white
// pixel on first use; no GPU readback, full image clearing or pixel upload is
// required. Artwork supplied in White survives Close. The component does not
// advance a caller's motion or music clocks.
type IndexedPointPlane struct {
	width, height, count int
	sample               func(int) (PointPlaneSample, bool)
	collision            PointPlaneCollision
	masks                []byte
	member               []uint64
	touched              []int
	paints               [256][4]float32
	offset               image.Point
	drawZero             bool
	white                *ebiten.Image
	ownedWhite           bool
	batch                *render.Batch
	blend                ebiten.Blend
	closed               bool
}

func validPointPlaneOffset(p image.Point) bool {
	return p.X >= -(1<<24) && p.X <= 1<<24 && p.Y >= -(1<<24) && p.Y <= 1<<24
}

func NewIndexedPointPlane(c IndexedPointPlaneConfig) (*IndexedPointPlane, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 || c.Width*c.Height > 16<<20 ||
		c.Count < 0 || c.Count > 1_000_000 || c.Count > 0 && c.Sample == nil ||
		c.Collision > PointPlaneXOR || len(c.Palette) < 1 || len(c.Palette) > 256 ||
		!validPointPlaneOffset(c.Offset) || c.BatchTriangles < 0 || c.BatchTriangles > 20000 || c.BatchTriangles == 1 {
		return nil, fmt.Errorf("sprites: invalid indexed point plane dimensions, population or material")
	}
	if c.White != nil && (c.White.Bounds().Dx() != 1 || c.White.Bounds().Dy() != 1) {
		return nil, fmt.Errorf("sprites: indexed point plane white material must contain one pixel")
	}
	capacity := c.BatchTriangles
	if capacity == 0 {
		capacity = max(2, min(20000, c.Count*2))
	}
	area := c.Width * c.Height
	p := &IndexedPointPlane{width: c.Width, height: c.Height, count: c.Count, sample: c.Sample,
		collision: c.Collision, masks: make([]byte, area), member: make([]uint64, (area+63)/64),
		touched: make([]int, 0, min(area, c.Count)), offset: c.Offset, drawZero: c.DrawZero,
		white: c.White, batch: render.NewBatch(capacity), blend: c.Blend}
	p.ownedWhite = p.white == nil
	if p.blend == (ebiten.Blend{}) {
		p.blend = ebiten.BlendSourceOver
	}
	p.SetPalette(c.Palette)
	return p, nil
}

// Reset removes the prepared frame in O(touched pixels) work.
func (p *IndexedPointPlane) Reset() {
	if p == nil || p.closed {
		return
	}
	for _, index := range p.touched {
		p.masks[index] = 0
		p.member[index>>6] &^= uint64(1) << uint(index&63)
	}
	p.touched = p.touched[:0]
}

// Sample clears the previous sparse raster and prepares one complete population.
// Points outside the configured plane are discarded before indexing storage.
func (p *IndexedPointPlane) Sample() error {
	if p == nil || p.closed {
		return fmt.Errorf("sprites: closed indexed point plane")
	}
	p.Reset()
	for i := 0; i < p.count; i++ {
		point, visible := p.sample(i)
		if !visible || point.X < 0 || point.X >= p.width || point.Y < 0 || point.Y >= p.height {
			continue
		}
		index := point.Y*p.width + point.X
		bit := uint64(1) << uint(index&63)
		if p.member[index>>6]&bit == 0 {
			p.member[index>>6] |= bit
			p.touched = append(p.touched, index)
		}
		if p.collision == PointPlaneXOR {
			p.masks[index] ^= point.Mask
		} else {
			p.masks[index] |= point.Mask
		}
	}
	return nil
}

func (p *IndexedPointPlane) Update(kit.Frame) error { return p.Sample() }

// SetPalette updates small premultiplied vertex colors without allocation.
// Indices beyond the supplied palette become transparent. Positions and masks
// remain unchanged, so another palette can redraw the same sampled frame.
func (p *IndexedPointPlane) SetPalette(colors []color.NRGBA) error {
	if p == nil || p.closed || len(colors) < 1 || len(colors) > 256 {
		return fmt.Errorf("sprites: invalid indexed point plane palette")
	}
	clear(p.paints[:])
	for index, paint := range colors {
		r, g, b, a := paint.RGBA()
		p.paints[index] = [4]float32{float32(r) / 65535, float32(g) / 65535, float32(b) / 65535, float32(a) / 65535}
	}
	return nil
}

// Masks and Touched borrow read-only CPU storage until the next Sample or Reset.
func (p *IndexedPointPlane) Masks() []byte  { return p.masks }
func (p *IndexedPointPlane) Touched() []int { return p.touched }

func (p *IndexedPointPlane) Draw(dst *ebiten.Image) {
	if p != nil {
		p.DrawAt(dst, p.offset)
	}
}

// DrawAt reuses the prepared raster at an independently selected integer origin.
// The destination's bounds clip pixels, including a nonzero-origin sub-image.
func (p *IndexedPointPlane) DrawAt(dst *ebiten.Image, offset image.Point) {
	if p == nil || p.closed || dst == nil || !validPointPlaneOffset(offset) {
		return
	}
	if p.white == nil {
		// Delay the optional GPU resource until drawing. Constructors and CPU
		// preparation can then be used by offline clocks without a game loop.
		p.white = ebiten.NewImage(1, 1)
		p.white.Fill(color.White)
	}
	p.batch.Options.Blend = p.blend
	p.batch.Begin(dst, p.white)
	source := p.white.Bounds()
	u, v := float32(source.Min.X), float32(source.Min.Y)
	for _, index := range p.touched {
		mask := p.masks[index]
		if mask == 0 && !p.drawZero {
			continue
		}
		paint := p.paints[mask]
		x, y := float32(index%p.width+offset.X), float32(index/p.width+offset.Y)
		vertex := ebiten.Vertex{DstX: x, DstY: y, SrcX: u, SrcY: v,
			ColorR: paint[0], ColorG: paint[1], ColorB: paint[2], ColorA: paint[3]}
		quad := [4]ebiten.Vertex{vertex, vertex, vertex, vertex}
		quad[1].DstX, quad[1].SrcX = x+1, u+1
		quad[2].DstX, quad[2].DstY, quad[2].SrcX, quad[2].SrcY = x+1, y+1, u+1, v+1
		quad[3].DstY, quad[3].SrcY = y+1, v+1
		p.batch.Quad(quad)
	}
	p.batch.Flush()
}

func (p *IndexedPointPlane) Close() error {
	if p == nil || p.closed {
		return nil
	}
	p.closed = true
	if p.ownedWhite && p.white != nil {
		p.white.Deallocate()
	}
	p.white, p.batch, p.sample = nil, nil, nil
	p.masks, p.member, p.touched = nil, nil, nil
	return nil
}

var _ kit.Effect = (*IndexedPointPlane)(nil)
