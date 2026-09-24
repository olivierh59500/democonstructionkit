package indexed

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestRotozoom256Reference(t *testing.T) {
	r, err := NewRotozoom256(Rotozoom256Config{Width: 160, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	source, rotated := make([]byte, rotozoomTextureSize), make([]byte, rotozoomTextureSize)
	for i := range source {
		source[i] = byte((i*37 + i/256*19) % 251)
		rotated[i] = byte((i*11 + i/256*53) % 247)
	}
	out := make([]byte, 160*100)
	for _, tc := range []struct {
		name         string
		x, y, xa, ya int
		want         string
	}{
		{"ordinary", 12, 34, 61, 40, "679bbf2157e8f14b5d93a365ef81c789b0e21969736c9941c084737be55203b8"},
		{"rotated", 250, 255, -30, 120, "3ecc148e073605a86709e425c45230e47afaca6a595d787643c3ea39d787cdb2"},
		{"wrapped", -12, 320, 0, -128, "089cfb513dc1cc9394997a4b377991c77a65f0b9722fcc376a73fa0f4ac9936a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clear(out)
			if err := r.Render(out, source, rotated, tc.x, tc.y, tc.xa, tc.ya); err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(out)); got != tc.want {
				t.Fatalf("indexed rotozoom hash = %s, want %s", got, tc.want)
			}
		})
	}
	var renderErr error
	if got := testing.AllocsPerRun(30, func() {
		renderErr = r.Render(out, source, rotated, 12, 34, 61, 40)
	}); got != 0 || renderErr != nil {
		t.Fatalf("render allocations = %v, error = %v", got, renderErr)
	}
}

func TestRotozoom256RejectsInvalidBuffers(t *testing.T) {
	for _, cfg := range []Rotozoom256Config{{}, {Width: 159, Height: 100}, {Width: 160, Height: -1}, {Width: 160, Height: 100, RowNumerator: 1}, {Width: 160, Height: 100, RowDenominator: 1}} {
		if _, err := NewRotozoom256(cfg); err == nil {
			t.Fatalf("accepted invalid rotozoom config %+v", cfg)
		}
	}
	r, err := NewRotozoom256(Rotozoom256Config{Width: 8, Height: 4})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Render(make([]byte, 31), make([]byte, rotozoomTextureSize), make([]byte, rotozoomTextureSize), 0, 0, 0, 0); err == nil {
		t.Fatal("accepted undersized output")
	}
}
