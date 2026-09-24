package indexed

import "fmt"

const rotozoomTextureSize = 256 * 256

// Rotozoom256Config describes a fixed-point 256×256 indexed texture sampler.
// Width must be a multiple of four. RowNumerator/RowDenominator scale the
// vertical step using unsigned 32-bit arithmetic; both zero select 307/256.
// Texture rotation remains a separate input so palette-index banks are never
// converted to RGBA or silently interpolated.
type Rotozoom256Config struct {
	Width, Height                int
	RowNumerator, RowDenominator uint32
}

// Rotozoom256 keeps four horizontal offset tables for allocation-free drawing.
// It reproduces the original word-packed fixed-point addressing, including
// 16-bit texture wrap, signed step conversion and steep-angle source switching.
// A renderer for live RGBA images can use composite.RotozoomBackground instead.
type Rotozoom256 struct {
	width, height                  int
	rowNumerator, rowDenominator   uint32
	modaLo, modaHi, modbLo, modbHi []uint16
}

func NewRotozoom256(c Rotozoom256Config) (*Rotozoom256, error) {
	if c.Width < 4 || c.Width%4 != 0 || c.Height < 1 || c.Width > 4096 || c.Height > 4096 || c.Width*c.Height > 1<<24 {
		return nil, fmt.Errorf("indexed: invalid rotozoom output dimensions")
	}
	if c.RowNumerator == 0 && c.RowDenominator == 0 {
		c.RowNumerator, c.RowDenominator = 307, 256
	}
	if c.RowNumerator == 0 || c.RowDenominator == 0 || c.RowNumerator > 1<<16 || c.RowDenominator > 1<<16 {
		return nil, fmt.Errorf("indexed: invalid rotozoom row scale")
	}
	n := c.Width / 4
	return &Rotozoom256{width: c.Width, height: c.Height, rowNumerator: c.RowNumerator, rowDenominator: c.RowDenominator,
		modaLo: make([]uint16, n), modaHi: make([]uint16, n), modbLo: make([]uint16, n), modbHi: make([]uint16, n)}, nil
}

// Render writes Width×Height palette indices to the start of dst. Source and
// rotated each contain 256×256 indices. X, Y and the two signed step inputs use
// the native integer units of the indexed effect; the source is selected by the
// larger step magnitude. The caller may reuse buffers across every frame.
func (r *Rotozoom256) Render(dst, source, rotated []byte, x, y, xa, ya int) error {
	if r == nil || len(dst) < r.width*r.height || len(source) < rotozoomTextureSize || len(rotated) < rotozoomTextureSize {
		return fmt.Errorf("indexed: incomplete rotozoom buffers")
	}
	xpos, ypos := int32(x)<<16, int32(y)<<16
	xadd, yadd := int32(int16(ya))<<6, int32(int16(xa))<<6
	texture := source
	if absRotoStep(xadd) > absRotoStep(yadd) {
		texture = rotated
		xadd, yadd = -yadd, xadd
		xpos, ypos = -ypos, xpos
	}

	var si, di uint16
	var al, ah uint8
	cx, dx := uint16(uint32(yadd)&0xffff), uint16(uint32(xadd)&0xffff)
	bl, bh := uint8(uint32(yadd)>>16), uint8(uint32(xadd)>>16)
	bh = uint8(-int8(bh))
	olddx := dx
	dx = 0 - dx
	if olddx != 0 {
		bh--
	}
	for i := range r.modaLo {
		for _, table := range []*[]uint16{&r.modaLo, &r.modbLo, &r.modaHi, &r.modbHi} {
			t := uint32(si) + uint32(cx)
			si = uint16(t)
			al = uint8(uint16(al) + uint16(bl) + uint16(uint8(t>>16)))
			t = uint32(di) + uint32(dx)
			di = uint16(t)
			ah = uint8(uint16(ah) + uint16(bh) + uint16(uint8(t>>16)))
			(*table)[i] = uint16(al) | uint16(ah)<<8
		}
	}
	xadd = int32((uint32(xadd) * r.rowNumerator) / r.rowDenominator)
	yadd = int32((uint32(yadd) * r.rowNumerator) / r.rowDenominator)
	out := dst[:r.width*r.height]
	for row := 0; row < r.height; row++ {
		ypos += yadd
		xpos += xadd
		base := uint16(((uint32(ypos) >> 8) & 0xff00) | ((uint32(xpos) >> 16) & 0x00ff))
		for i := range r.modaLo {
			out[0] = texture[int(uint16(base+r.modaLo[i]))]
			out[1] = texture[int(uint16(base+r.modaHi[i]))]
			out[2] = texture[int(uint16(base+r.modbLo[i]))]
			out[3] = texture[int(uint16(base+r.modbHi[i]))]
			out = out[4:]
		}
	}
	return nil
}

func absRotoStep(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
