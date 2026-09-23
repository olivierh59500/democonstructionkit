package presets

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/effects"
)

func TestBilizirCubePreservesGeometryAndColors(t *testing.T) {
	h := sha256.New()
	var b [4]byte
	for _, size := range []float64{1, 20, 100} {
		c, err := effects.NewSolidCube(BilizirCube(size))
		if err != nil {
			t.Fatal(err)
		}
		for tick := 0; tick < 1000; tick++ {
			c.Rotate(.021, .037, .013)
			vertices, indices := c.Geometry(380.0+float64(tick%31)/8, 186.0-float64(tick%19)/7)
			for _, p := range vertices {
				for _, x := range [...]float32{p.DstX, p.DstY, p.SrcX, p.SrcY, p.ColorR, p.ColorG, p.ColorB, p.ColorA} {
					binary.LittleEndian.PutUint32(b[:], math.Float32bits(x))
					h.Write(b[:])
				}
			}
			for _, x := range indices {
				binary.LittleEndian.PutUint16(b[:2], x)
				h.Write(b[:2])
			}
		}
		if n := testing.AllocsPerRun(100, func() { c.Geometry(123, 456) }); n != 0 {
			t.Fatalf("geometry allocated %g", n)
		}
		c.Close()
		c.Close()
	}
	// Captured from the former standalone renderer: every float32 vertex field,
	// color and index for 3,000 distinct rotations/positions and three cube sizes.
	if got := fmt.Sprintf("%x", h.Sum(nil)); got != "7f61d14fe75e0403352967aa13ac17d53497bf0548a268379a35d7d63d5830a6" {
		t.Fatalf("cube geometry changed: %s", got)
	}
}

func TestSolidCubeOptions(t *testing.T) {
	cfg := effects.DefaultSolidCubeConfig(30)
	cfg.EdgeWidth = 0
	cube, err := effects.NewSolidCube(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	v, idx := cube.Geometry(0, 0)
	if len(v) != 24 || len(idx) != 36 {
		t.Fatalf("unoutlined geometry %d/%d", len(v), len(idx))
	}
	for _, change := range []func(*effects.SolidCubeConfig){
		func(c *effects.SolidCubeConfig) { c.Size = 0 },
		func(c *effects.SolidCubeConfig) { c.Perspective = 20 },
		func(c *effects.SolidCubeConfig) { c.EdgeWidth = -1 },
		func(c *effects.SolidCubeConfig) { c.Size = math.NaN() },
	} {
		bad := cfg
		change(&bad)
		if c, err := effects.NewSolidCube(bad); err == nil {
			c.Close()
			t.Fatalf("invalid cube %+v", bad)
		}
	}
}
