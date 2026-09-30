package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestFanMapsEachVertexOnceAndKeepsTriangleOrder(t *testing.T) {
	b := NewBatch(16)
	counts := [5]int{}
	at := func(i int) ebiten.Vertex { counts[i]++; return ebiten.Vertex{DstX: float32(i)} }
	b.Fan(5, at)
	for i, count := range counts {
		if count != 1 {
			t.Fatalf("vertex %d mapped %d times", i, count)
		}
	}
	want := []float32{0, 1, 2, 0, 2, 3, 0, 3, 4}
	if len(b.vertices) != len(want) {
		t.Fatal("fan triangle count changed")
	}
	for i, v := range b.vertices {
		if v.DstX != want[i] {
			t.Fatal("fan vertex order changed", i, v.DstX, want[i])
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { b.vertices = b.vertices[:0]; b.indices = b.indices[:0]; b.Fan(5, at) }); allocs != 0 {
		t.Fatalf("fan mapping allocated: %v", allocs)
	}
}
