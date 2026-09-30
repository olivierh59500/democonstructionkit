package composite

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestContourBankBoundsLifetimeAndBudgets(t *testing.T) {
	for _, c := range []ContourBankConfig{{}, {Width: 8192, Height: 8192, Slots: 1, Layers: 1},
		{Width: 8, Height: 8, Slots: 0, Layers: 1}, {Width: 8, Height: 8, Slots: 1, Layers: 9},
		{Width: 8, Height: 8, Slots: 1, Layers: 1, FillRule: ebiten.FillRule(99)}} {
		if b, err := NewContourBank(c); err == nil {
			b.Close()
			t.Fatal("accepted invalid contour bank")
		}
	}
	b, err := NewContourBank(ContourBankConfig{Width: 8, Height: 8, Slots: 3, Layers: 2, FillRule: ebiten.FillRuleEvenOdd})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if b.Image(0, 0) == b.Image(0, 1) || b.Image(0, 0) == b.Image(1, 0) || b.Image(-1, 0) != nil || b.Image(3, 0) != nil || b.Image(0, 2) != nil {
		t.Fatal("slot/layer identities changed")
	}
	if b.Paint(3, 0, true, nil) == nil || b.ClearSlot(-1) == nil {
		t.Fatal("invalid contour indices were accepted")
	}
	b.Close()
	if b.Image(0, 0) != nil || b.Paint(0, 0, false, nil) == nil {
		t.Fatal("closed bank still exposes images")
	}
	var absent *ContourBank
	if absent.Image(0, 0) != nil || absent.Paint(0, 0, false, nil) == nil {
		t.Fatal("nil bank was accepted")
	}
}
