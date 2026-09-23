package effects

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestSolidCubeBatchIsBoundedAndAllocationFree(t *testing.T) {
	c, err := NewSolidCube(DefaultSolidCubeConfig(40))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.Rotation = geometry.Vec3{X: .3, Y: .7, Z: 1.1}
	b := NewSolidCubeBatch(2)
	defer b.Close()
	if !b.Add(c, 10, 20) || !b.Add(c, 30, 40) || b.Add(c, 50, 60) {
		t.Fatal("batch did not enforce fixed capacity")
	}
	if n := testing.AllocsPerRun(1000, func() { b.Reset(); b.Add(c, 10, 20); b.Add(c, 30, 40) }); n != 0 {
		t.Fatalf("batch allocated %g objects", n)
	}
	for _, index := range b.indices {
		if int(index) >= len(b.vertices) {
			t.Fatal("batch rebased an index outside the mesh")
		}
	}
	b.Close()
	b.Close()
	if b.Add(c, 0, 0) {
		t.Fatal("closed batch accepted geometry")
	}
}

func TestSolidCubeRecurrenceAcceptsExternalRotationChanges(t *testing.T) {
	cfg := DefaultSolidCubeConfig(40)
	cfg.RotationRecurrence = true
	c, err := NewSolidCube(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for i := 0; i < 2300; i++ {
		if i == 1300 {
			c.Rotation = geometry.Vec3{X: -3, Y: 7, Z: -.3}
		}
		c.Rotate(.03, .04, .07)
		c.Geometry(100, 100)
		sx, cx := math.Sincos(c.Rotation.X)
		sy, cy := math.Sincos(c.Rotation.Y)
		sz, cz := math.Sincos(c.Rotation.Z)
		if math.Abs(c.sinX-sx) > 1e-11 || math.Abs(c.cosX-cx) > 1e-11 || math.Abs(c.sinY-sy) > 1e-11 || math.Abs(c.cosY-cy) > 1e-11 || math.Abs(c.sinZ-sz) > 1e-11 || math.Abs(c.cosZ-cz) > 1e-11 {
			t.Fatalf("recurrence drift at %d", i)
		}
	}
}
