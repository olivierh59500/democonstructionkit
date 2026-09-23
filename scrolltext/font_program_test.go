package scrolltext

import "testing"

func TestFontProgramMasksAndCursor(t *testing.T) {
	p, err := NewFontProgram("Aé{font:large}BC{font:small}{font:large}D{font:small}!", Braces, "small")
	if err != nil {
		t.Fatal(err)
	}
	if p.Len() != 6 {
		t.Fatalf("glyph count %d", p.Len())
	}
	if got := p.MaskedText("small", ' '); got != "Aé   !" {
		t.Fatalf("small %q", got)
	}
	if got := p.MaskedText("large", '_'); got != "__BCD_" {
		t.Fatalf("large %q", got)
	}
	for _, tc := range []struct {
		i    int
		font string
	}{{-1, "small"}, {0, "small"}, {1, "small"}, {2, "large"}, {4, "large"}, {5, "small"}, {999, "small"}} {
		if got := p.FontAt(tc.i); got != tc.font {
			t.Fatalf("font at %d = %q, want %q", tc.i, got, tc.font)
		}
	}
	if n := testing.AllocsPerRun(100, func() { p.FontAt(4) }); n != 0 {
		t.Fatalf("lookup allocated %g", n)
	}
}

func TestFontProgramRejectsLostControls(t *testing.T) {
	for _, text := range []string{"{speed:20}A", "{font:}", "{font:large", string([]byte{0xff})} {
		if _, err := NewFontProgram(text, Braces, "small"); err == nil {
			t.Fatalf("accepted %q", text)
		}
	}
	p, err := NewFontProgram("^Cs2;A^Cs0;B", DomSizes, "0")
	if err != nil || p.MaskedText("2", ' ') != "A " || p.FontAt(0) != "2" {
		t.Fatalf("font syntax: %v %v", p, err)
	}
}
