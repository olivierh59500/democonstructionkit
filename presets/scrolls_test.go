package presets

import "testing"

func TestProjectedPresetKeepsUnicodeGlyphsAndControlSlots(t *testing.T) {
	slots := TCBPlaneSlots("É^2Ω", 17)
	if len(slots) != 4 {
		t.Fatal(slots)
	}
	for i, want := range []rune{'É', 'É', 'É', 'Ω'} {
		if slots[i].Rune != want || slots[i].Advance != 17 {
			t.Fatal(i, slots[i])
		}
	}
	if slots[1].Form != 2 || slots[2].Form != -1 {
		t.Fatal(slots)
	}
	if len(TCBPlaneSlots("", 17)) != 0 {
		t.Fatal("empty text acquired glyph slots")
	}
}
