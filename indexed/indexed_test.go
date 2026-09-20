package indexed

import (
	"bytes"
	"testing"
)

func TestEveryPaletteIndexAndByteOrder(t *testing.T) {
	var p [256]uint32
	indices := make([]byte, 256)
	want := make([]byte, 1024)
	for i := range p {
		r, g, b, a := byte(i), byte(255-i), byte(i^0x55), byte(255)
		p[i] = uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
		indices[i] = byte(i)
		copy(want[i*4:], []byte{r, g, b, a})
	}
	got := make([]byte, len(want))
	if err := ExpandRGBA(got, indices, &p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("palette expansion changed pixel channels")
	}
	if n := testing.AllocsPerRun(100, func() { _ = ExpandRGBA(got, indices, &p) }); n != 0 {
		t.Fatal(n)
	}
}
