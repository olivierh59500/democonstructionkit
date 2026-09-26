package effects

import (
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestMaskWithUpdatesContentBeforeAlpha(t *testing.T) {
	var updateOrder []int
	mask, err := NewMaskWith(MaskConfig{
		Content: kit.Func{OnUpdate: func(kit.Frame) error { updateOrder = append(updateOrder, 1); return nil }},
		Alpha:   kit.Func{OnUpdate: func(kit.Frame) error { updateOrder = append(updateOrder, 2); return nil }},
		Width:   2, Height: 4,
		MaskY: 1, OutputX: 2, OutputY: 3, ClearTop: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mask.Close()
	if err := mask.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if len(updateOrder) != 2 || updateOrder[0] != 1 || updateOrder[1] != 2 {
		t.Fatalf("mask update order %v", updateOrder)
	}
}
