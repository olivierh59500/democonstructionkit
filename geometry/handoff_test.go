package geometry

import "testing"

func TestHandoffMatchesBothEndpointsAndInterruptedTransition(t *testing.T) {
	var h Handoff
	from := []Vec3{{1, 2, 3}, {-4, 8, 9}}
	incoming := []Vec3{{100, 200, 300}, {9, -5, 7}}
	out := make([]Vec3, 2)
	if err := h.Begin(from, incoming, 10, 2); err != nil {
		t.Fatal(err)
	}
	h.Apply(out, incoming, 10)
	for i := range from {
		if out[i] != from[i] {
			t.Fatal("entry position changed")
		}
	}
	h.Apply(out, incoming, 12)
	for i := range from {
		if out[i] != incoming[i] {
			t.Fatal("effect did not converge")
		}
	}
	h.Apply(out, incoming, 11)
	previous := append([]Vec3(nil), out...)
	if err := h.Begin(previous, from, 11, .5); err != nil {
		t.Fatal(err)
	}
	h.Apply(out, from, 11)
	for i := range out {
		if out[i] != previous[i] {
			t.Fatal("interrupted transition jumped")
		}
	}
	copy(out, from)
	h.Apply(out, out, 11.5)
	for i := range out {
		if out[i] != from[i] {
			t.Fatal("in-place application failed")
		}
	}
}
