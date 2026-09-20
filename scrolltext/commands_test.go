package scrolltext

import "testing"

func TestOptionalControlsAndOriginalDomSyntax(t *testing.T) {
	plain, err := Parse("A{speed:20}B", nil)
	if err != nil || len(plain) != 1 || plain[0].Text != "A{speed:20}B" {
		t.Fatal(plain, err)
	}
	tokens, err := Parse("A{speed:20}{font:large}É{{B", Braces)
	if err != nil || len(tokens) != 5 || tokens[1].Kind != Speed || tokens[2].Text != "large" || tokens[3].Text != "É" || tokens[4].Text != "{B" { // Escaped text may occupy a separate token.
		if err != nil || len(tokens) != 6 || tokens[4].Text != "{" || tokens[5].Text != "B" {
			t.Fatal(tokens, err)
		}
	}
	dom, err := Parse("ONE^Cs3;TWO", DomSizes)
	if err != nil || len(dom) != 3 || dom[1].Kind != Font || dom[1].Text != "3" {
		t.Fatal(dom, err)
	}
}
func TestMalformedControlIsNotSilentlyRendered(t *testing.T) {
	for _, s := range []string{"{speed:no}", "{unknown:1}", "{font:}", "{pause:2"} {
		if _, err := Parse(s, Braces); err == nil {
			t.Fatal(s)
		}
	}
}
