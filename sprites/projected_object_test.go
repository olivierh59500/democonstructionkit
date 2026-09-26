package sprites

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestProjectedObjectMatchesFourVectorballShapesAndFlag(t *testing.T) {
	const size = 640.0
	flag := &Flag{Width: size, PinLeft: true,
		Wave: motion.Wave{Amplitude: size * .15, Spatial: 7 / size, Speed: 3}, RowPhase: 2 / size}
	cases := []struct {
		name string
		cfg  ProjectedObjectConfig
		rest []Point
	}{
		{name: "cube", cfg: ProjectedObjectConfig{Cube: &CubeConfig{Size: size, Segments: 4, Fill: Edges, Image: 120}}},
		{name: "pyramid", cfg: ProjectedObjectConfig{Pyramid: &PyramidConfig{Width: size, Height: size, Segments: 6, Fill: Solid, Image: 120}}},
		{name: "plane", cfg: ProjectedObjectConfig{Plane: &PlaneConfig{Width: size, Height: size * .65, Columns: 8, Rows: 8, Image: 120}}},
		{name: "flag", cfg: ProjectedObjectConfig{Plane: &PlaneConfig{Width: size, Height: size * .65, Columns: 12, Rows: 12, Image: 120}, Flag: flag}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := test.cfg
			c.Scale, c.Focal, c.CenterX, c.CenterY = .55, 1450, 320, 193
			c.Position = geometry.Vec3{Z: 850}
			c.RotationStep = geometry.Vec3{X: .012, Y: .017, Z: .005}
			var err error
			switch {
			case c.Cube != nil:
				test.rest, err = Cube(*c.Cube)
			case c.Pyramid != nil:
				test.rest, err = Pyramid(*c.Pyramid)
			case c.Plane != nil:
				test.rest, err = Plane(*c.Plane)
			}
			if err != nil {
				t.Fatal(err)
			}
			object, err := NewProjectedObject(c)
			if err != nil {
				t.Fatal(err)
			}
			wantPoints := make([]Point, len(test.rest))
			rotation := geometry.Vec3{}
			for tick := 0; tick < 5000; tick++ {
				seconds := float64(tick) / 60
				copy(wantPoints, test.rest)
				if c.Flag != nil {
					c.Flag.Apply(wantPoints, test.rest, seconds)
				}
				rotation.X += .012
				rotation.Y += .017
				rotation.Z += .005
				if err := object.Update(kit.Frame{Time: seconds}); err != nil {
					t.Fatal(err)
				}
				if got := object.Rotation(); got != rotation {
					t.Fatalf("tick %d rotation = %+v, want %+v", tick, got, rotation)
				}
				for i, point := range object.Points() {
					if point != wantPoints[i] {
						t.Fatalf("tick %d point %d = %+v, want %+v", tick, i, point, wantPoints[i])
					}
				}
			}
			if allocations := testing.AllocsPerRun(100, func() {
				_ = object.Update(kit.Frame{Time: 100})
			}); allocations != 0 {
				t.Fatalf("projected object update allocated %v times", allocations)
			}
		})
	}
}

func TestProjectedObjectKeepsIndependentEditableInstances(t *testing.T) {
	points := []Point{{X: -2, Y: 1, Image: 3}, {X: 2, Y: -1, Image: 4}}
	c := ProjectedObjectConfig{Points: points, Scale: 1, Focal: 600,
		RotationStep: geometry.Vec3{Y: .25}, Position: geometry.Vec3{Z: 850}}
	first, err := NewProjectedObject(c)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewProjectedObject(c)
	if err != nil {
		t.Fatal(err)
	}
	points[0].X = 99
	if first.Points()[0].X != -2 {
		t.Fatal("object retained the caller's mutable point slice")
	}
	if err := first.SetPosition(geometry.Vec3{X: 10, Z: 900}); err != nil {
		t.Fatal(err)
	}
	if err := first.SetRotationStep(geometry.Vec3{Y: .5}); err != nil {
		t.Fatal(err)
	}
	if err := first.SetScale(.75); err != nil {
		t.Fatal(err)
	}
	if err := first.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if first.Rotation().Y != .5 || second.Rotation().Y != 0 || second.Position().X != 0 {
		t.Fatal("editing one projected object changed another")
	}
	first.Reset()
	if first.Rotation() != (geometry.Vec3{}) || first.Points()[0].X != -2 {
		t.Fatal("object reset did not restore its authored pose")
	}
}

func TestProjectedObjectRejectsInvalidSourcesAndTransforms(t *testing.T) {
	c := ProjectedObjectConfig{Cube: &CubeConfig{Size: 20, Segments: 2},
		Plane: &PlaneConfig{Width: 20, Height: 20, Columns: 2, Rows: 2}, Scale: 1, Focal: 600}
	if _, err := NewProjectedObject(c); err == nil {
		t.Fatal("accepted two point sources")
	}
	c.Plane = nil
	c.Flag = &Flag{Width: 20, PinLeft: true}
	if _, err := NewProjectedObject(c); err == nil {
		t.Fatal("accepted a flag deformation on a cube")
	}
	c.Flag = nil
	c.Position.X = math.NaN()
	if _, err := NewProjectedObject(c); err == nil {
		t.Fatal("accepted a nonfinite projection position")
	}
}
