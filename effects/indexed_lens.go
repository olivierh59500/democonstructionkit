package effects

import (
	"encoding/binary"
	"fmt"
)

// IndexedLensConfig describes a lens built from little-endian displacement
// tables. Width and Height describe the lens, CanvasWidth and CanvasHeight the
// indexed image. Source and destination may include trailing padding.
//
// Each table starts with Height records: a uint16 byte offset and a uint16 count.
// A record points to an int16 base displacement followed by its entries. Dense
// entries are int16 source displacements for consecutive destination pixels;
// Sparse entries are (int16 destination, int16 source) displacement pairs;
// Restore entries are int16 displacements copied without a palette mask.
// Dense rows preserve the original 128-pixel word alignment convention. Counts
// below four are inactive. Height must be even. Masks are ORed into palette
// indices, allowing independent tinted palette banks without RGB conversions.
// The constructor validates and compiles tables once; drawing does not allocate.
// Use composite.Magnifier for an analytic lens over arbitrary live RGBA images.
type IndexedLensConfig struct {
	Width, Height             int
	CanvasWidth, CanvasHeight int
	Dense                     []byte
	Sparse                    [2][]byte
	Restore                   []byte
	Masks                     [3]byte
}

type indexedLensSpan struct {
	destination, source, second int
	pair                        bool
	mask                        byte
}

// IndexedLens retains the exact integer mapping and palette-bank behavior of
// a table-driven lens. Its compiled rows can be reused with any indexed image
// of the configured dimensions. It does not retain the input tables.
type IndexedLens struct {
	width, height, stride, canvasSize int
	rows                              [][2][]indexedLensSpan
}

func NewIndexedLens(c IndexedLensConfig) (*IndexedLens, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Height%2 != 0 || c.CanvasWidth <= 0 || c.CanvasHeight <= 0 || c.CanvasWidth > int(^uint(0)>>1)/c.CanvasHeight {
		return nil, fmt.Errorf("effects: invalid indexed lens dimensions")
	}
	if c.Height > len(c.Dense)/4 || c.Height > len(c.Sparse[0])/4 || c.Height > len(c.Sparse[1])/4 || c.Height > len(c.Restore)/4 {
		return nil, fmt.Errorf("effects: incomplete indexed lens row tables")
	}
	l := &IndexedLens{width: c.Width, height: c.Height, stride: c.CanvasWidth, canvasSize: c.CanvasWidth * c.CanvasHeight, rows: make([][2][]indexedLensSpan, c.Height)}
	for y := range l.rows {
		for parity := 0; parity < 2; parity++ {
			spans, err := compileIndexedLensRow(c.Dense, y, parity, 0, c.Masks[0])
			if err != nil {
				return nil, err
			}
			for i, table := range c.Sparse {
				extra, err := compileIndexedLensRow(table, y, parity, 1, c.Masks[i+1])
				if err != nil {
					return nil, err
				}
				spans = append(spans, extra...)
			}
			extra, err := compileIndexedLensRow(c.Restore, y, parity, 2, 0)
			if err != nil {
				return nil, err
			}
			l.rows[y][parity] = append(spans, extra...)
		}
	}
	return l, nil
}

func compileIndexedLensRow(data []byte, row, parity, kind int, mask byte) ([]indexedLensSpan, error) {
	header := row * 4
	offset := int(binary.LittleEndian.Uint16(data[header:]))
	count := int(binary.LittleEndian.Uint16(data[header+2:]))
	if kind != 2 {
		count = int(int16(count))
	}
	if count < 0 {
		return nil, fmt.Errorf("effects: negative indexed lens row count")
	}
	if count == 0 || kind == 0 && count < 4 {
		return nil, nil
	}
	words := count
	if kind == 1 {
		words *= 2
	}
	if offset < 0 || offset > len(data)-2 || words > (len(data)-offset-2)/2 {
		return nil, fmt.Errorf("effects: truncated indexed lens row %d", row)
	}
	base := int(int16(binary.LittleEndian.Uint16(data[offset:])))
	cursor := offset + 2
	read := func(at int) int { return int(int16(binary.LittleEndian.Uint16(data[at:]))) }
	out := make([]indexedLensSpan, 0, count)
	switch kind {
	case 0:
		destination := base
		if (parity+base)&1 != 0 {
			out = append(out, indexedLensSpan{destination: destination, source: base + read(cursor), mask: mask})
			destination++
			cursor += 2
			count--
		}
		pairs := count / 2
		start := max(0, 64-pairs)
		for i := start; i < 64 && i < start+pairs; i++ {
			at := 63 - i
			out = append(out, indexedLensSpan{destination: destination + at*2, source: base + read(cursor+at*4), second: base + read(cursor+at*4+2), pair: true, mask: mask})
		}
		if count&1 != 0 {
			out = append(out, indexedLensSpan{destination: destination + (count &^ 1), source: base + read(cursor+(count&^1)*2), mask: mask})
		}
	case 1:
		for i := 0; i < count; i++ {
			out = append(out, indexedLensSpan{destination: base + read(cursor), source: base + read(cursor+2), mask: mask})
			cursor += 4
		}
	case 2:
		for i := 0; i < count; i++ {
			at := base + read(cursor)
			out = append(out, indexedLensSpan{destination: at, source: at})
			cursor += 2
		}
	}
	return out, nil
}

// Draw applies the lens centered on (centerX, centerY), preserving destination
// pixels outside its mapping. Source must remain unchanged while drawing: use
// separate source/destination buffers when their memory overlaps. Invalid or
// clipped offsets are skipped. No palette conversion or GPU readback is needed.
func (l *IndexedLens) Draw(destination, source []byte, centerX, centerY int) {
	if l == nil {
		return
	}
	top := centerX - l.width/2 + (centerY-l.height/2)*l.stride
	bottom := centerX - l.width/2 + (centerY+l.height/2-1)*l.stride
	for y := 0; y < l.height/2; y++ {
		l.drawRow(destination, source, top, y)
		l.drawRow(destination, source, bottom, l.height-1-y)
		top += l.stride
		bottom -= l.stride
	}
}

func (l *IndexedLens) drawRow(destination, source []byte, base, row int) {
	if base < 0 || base > l.canvasSize {
		return
	}
	for _, p := range l.rows[row][base&1] {
		dst, src := base+p.destination, base+p.source
		if dst < 0 || dst >= len(destination) || src < 0 || src >= len(source) {
			continue
		}
		if p.pair {
			second := base + p.second
			if dst+1 >= len(destination) || second < 0 || second >= len(source) {
				continue
			}
			destination[dst+1] = source[second] | p.mask
		}
		destination[dst] = source[src] | p.mask
	}
}
