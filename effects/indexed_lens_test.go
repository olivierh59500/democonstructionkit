package effects

import (
	"encoding/binary"
	"testing"
)

func indexedLensTestTable(words []int16, count int) []byte {
	table := make([]byte, 8)
	for row := 0; row < 2; row++ {
		binary.LittleEndian.PutUint16(table[row*4:], uint16(len(table)))
		binary.LittleEndian.PutUint16(table[row*4+2:], uint16(count))
		for _, word := range words {
			table = binary.LittleEndian.AppendUint16(table, uint16(word))
		}
	}
	return table
}

func indexedLensTestConfig() IndexedLensConfig {
	return IndexedLensConfig{Width: 4, Height: 2, CanvasWidth: 8, CanvasHeight: 4,
		Dense:   indexedLensTestTable([]int16{0, 3, 2, 1, 0}, 4),
		Sparse:  [2][]byte{indexedLensTestTable([]int16{0, 1, 0}, 1), indexedLensTestTable([]int16{0, 2, 1}, 1)},
		Restore: indexedLensTestTable([]int16{0, 3}, 1), Masks: [3]byte{0x40, 0x80, 0xc0}}
}

func TestIndexedLensPaletteBanksAndAlignment(t *testing.T) {
	lens, err := NewIndexedLens(indexedLensTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	source, destination := make([]byte, 32), make([]byte, 32)
	for i := range source {
		source[i] = byte(i)
	}
	for _, x := range []int{4, 5} {
		clear(destination)
		lens.Draw(destination, source, x, 2)
		want := make([]byte, 32)
		for _, at := range []int{x - 2 + 8, x - 2 + 16} {
			want[at] = byte(at+3) | 0x40
			want[at+1] = byte(at) | 0x80
			want[at+2] = byte(at+1) | 0xc0
			want[at+3] = byte(at + 3)
		}
		for i := range want {
			if destination[i] != want[i] {
				t.Fatalf("x=%d pixel %d: got %d, want %d", x, i, destination[i], want[i])
			}
		}
	}
	if got := testing.AllocsPerRun(100, func() { lens.Draw(destination, source, 4, 2) }); got != 0 {
		t.Fatalf("draw allocated %g times", got)
	}
}

func TestIndexedLensRejectsTruncatedTables(t *testing.T) {
	for table := 0; table < 4; table++ {
		original := indexedLensTestConfig()
		data := [][]byte{original.Dense, original.Sparse[0], original.Sparse[1], original.Restore}[table]
		for size := 0; size < len(data); size++ {
			c := indexedLensTestConfig()
			switch table {
			case 0:
				c.Dense = data[:size]
			case 1, 2:
				c.Sparse[table-1] = data[:size]
			case 3:
				c.Restore = data[:size]
			}
			if _, err := NewIndexedLens(c); err == nil {
				t.Fatalf("table %d accepted only %d/%d bytes", table, size, len(data))
			}
		}
	}
}

func TestIndexedLensClipsShortBuffers(t *testing.T) {
	lens, err := NewIndexedLens(indexedLensTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 32; n++ {
		destination := make([]byte, n)
		for m := 0; m < 32; m++ {
			source := make([]byte, m)
			for _, xy := range [][2]int{{4, 2}, {5, 2}, {0, 0}, {-20, -20}, {20, 20}} {
				lens.Draw(destination, source, xy[0], xy[1])
			}
		}
	}
}
