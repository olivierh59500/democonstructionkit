package font

import (
	"reflect"
	"testing"
)

func TestTileAlphabetVariableWidthsAndFallback(t *testing.T) {
	a := TileAlphabet{Order: " AB!", Widths: []int{1, 3, 2, 0}, Stride: 3, Ignore: "\r\n", Uppercase: true}
	got, err := a.Compile("a\r\nb!? ")
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{3, 4, 5, 6, 7, 0, 0}; !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	a.Order = "AA  "
	if _, err = a.Compile("A"); err == nil {
		t.Fatal("duplicate order accepted")
	}
	a = TileAlphabet{First: 'A', Widths: []int{2, 1}, Stride: 2, Fallback: 1}
	got, err = a.Compile("AZB")
	if err != nil || !reflect.DeepEqual(got, []int{0, 1, 2, 2}) {
		t.Fatal(got, err)
	}
	a.Widths[0] = 3
	if _, err = a.Compile("A"); err == nil {
		t.Fatal("overlapping glyph tiles accepted")
	}
}
