package composite

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// BlockRevealConfig reveals an image as cached cells in an editable order.
// Empty Order selects row-major cells. Background, when set, fills the entire
// destination before drawing the visible blocks; nil preserves earlier layers.
type BlockRevealConfig struct {
	Image                 *ebiten.Image
	CellWidth, CellHeight int
	Order                 []int
	Timing                motion.SteppedRevealConfig
	Background            color.Color
	OutputX, OutputY      float64
	Filter                ebiten.Filter
	Blend                 ebiten.Blend
}

// BlockReveal owns its small view cache and tick clock, but borrows the image.
// Any row, column, checkerboard or user-selected order can use the same effect.
type BlockReveal struct {
	config  BlockRevealConfig
	clock   *motion.SteppedReveal
	blocks  []*ebiten.Image
	order   []int
	columns int
}

func NewBlockReveal(c BlockRevealConfig) (*BlockReveal, error) {
	if c.Image == nil || c.Image.Bounds().Empty() || c.CellWidth < 1 || c.CellHeight < 1 ||
		c.CellWidth > 8192 || c.CellHeight > 8192 ||
		math.IsNaN(c.OutputX) || math.IsInf(c.OutputX, 0) ||
		math.IsNaN(c.OutputY) || math.IsInf(c.OutputY, 0) {
		return nil, fmt.Errorf("composite: invalid block reveal image or geometry")
	}
	bounds := c.Image.Bounds()
	columns := (bounds.Dx() + c.CellWidth - 1) / c.CellWidth
	rows := (bounds.Dy() + c.CellHeight - 1) / c.CellHeight
	if columns > (1<<20)/rows {
		return nil, fmt.Errorf("composite: too many reveal blocks")
	}
	count := columns * rows
	if c.Timing.Blocks == 0 {
		c.Timing.Blocks = count
	}
	if c.Timing.Blocks != count {
		return nil, fmt.Errorf("composite: reveal timing count differs from image grid")
	}
	clock, err := motion.NewSteppedReveal(c.Timing)
	if err != nil {
		return nil, err
	}
	order := make([]int, count)
	if len(c.Order) == 0 {
		for i := range order {
			order[i] = i
		}
	} else {
		if len(c.Order) != count {
			return nil, fmt.Errorf("composite: incomplete reveal order")
		}
		seen := make([]bool, count)
		for i, index := range c.Order {
			if index < 0 || index >= count || seen[index] {
				return nil, fmt.Errorf("composite: duplicate or invalid reveal block")
			}
			seen[index] = true
			order[i] = index
		}
	}
	blocks := make([]*ebiten.Image, count)
	for i := range blocks {
		x := bounds.Min.X + i%columns*c.CellWidth
		y := bounds.Min.Y + i/columns*c.CellHeight
		blocks[i] = c.Image.SubImage(image.Rect(x, y,
			min(x+c.CellWidth, bounds.Max.X), min(y+c.CellHeight, bounds.Max.Y))).(*ebiten.Image)
	}
	c.Order = nil
	return &BlockReveal{config: c, clock: clock, blocks: blocks, order: order, columns: columns}, nil
}

func (r *BlockReveal) Update(kit.Frame) error       { return r.clock.Step() }
func (r *BlockReveal) Done() bool                   { return r.clock.Done() }
func (r *BlockReveal) VisibleCount() int            { return r.clock.VisibleCount() }
func (r *BlockReveal) Clock() *motion.SteppedReveal { return r.clock }

func (r *BlockReveal) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil {
		return
	}
	if r.config.Background != nil {
		dst.Fill(r.config.Background)
	}
	for _, index := range r.order[:r.clock.VisibleCount()] {
		block := r.blocks[index]
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = r.config.Filter, r.config.Blend
		op.GeoM.Translate(r.config.OutputX+float64(index%r.columns*r.config.CellWidth),
			r.config.OutputY+float64(index/r.columns*r.config.CellHeight))
		dst.DrawImage(block, &op)
	}
}

func (r *BlockReveal) Close() error {
	if r != nil {
		r.blocks = nil
		r.order = nil
	}
	return nil
}
