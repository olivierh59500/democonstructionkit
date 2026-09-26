package geometry

import (
	"cmp"
	"math"
	"slices"
	"testing"
)

func TestMorphingMeshKeepsDMA3DTransitionAndProjection(t *testing.T) {
	shapes := [][]Vec3{
		{{X: 0, Y: 0, Z: 200}, {X: -100, Y: 25, Z: 100}, {X: 100, Y: 25, Z: 100}, {X: 0, Y: -25, Z: -100}},
		{{X: 0, Y: 0, Z: 100}, {X: -100, Y: 100, Z: 100}, {X: 100, Y: 100, Z: 100}, {X: 0, Y: -100, Z: -100}},
		{{X: 0, Y: 0, Z: 200}, {X: -100, Y: 100, Z: 100}, {X: 100, Y: 100, Z: 100}, {X: 0, Y: -200, Z: -200}},
	}
	faces := []MorphFace{
		{Vertices: [3]int{0, 2, 1}},
		{Vertices: [3]int{0, 3, 2}},
		{Vertices: [3]int{1, 2, 3}},
	}
	mesh, err := NewMorphingMesh(MorphingMeshConfig{
		Shapes: shapes, Faces: faces, MorphTicks: 120, CycleTicks: 240,
		RotationStep: Vec3{X: .01, Y: .02, Z: .04},
		CenterX:      320, CenterY: 240, Focal: 900, CameraZ: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	current := append([]Vec3(nil), shapes[0]...)
	rotation := Vec3{}
	shape, target := 0, 1
	timer := 0.0
	for tick := 1; tick <= 1000; tick++ {
		rotation.X += .01
		rotation.Y += .02
		rotation.Z += .04
		if rotation.X >= 2*math.Pi {
			rotation.X -= 2 * math.Pi
		}
		if rotation.Y >= 2*math.Pi {
			rotation.Y -= 2 * math.Pi
		}
		if rotation.Z >= 2*math.Pi {
			rotation.Z -= 2 * math.Pi
		}
		timer++
		if timer >= 240 {
			timer = 0
			shape = target
			target = (target + 1) % len(shapes)
		}
		if timer <= 120 {
			fraction := timer / 120
			for i := range current {
				a, b := shapes[shape][i], shapes[target][i]
				current[i] = Vec3{
					X: a.X + (b.X-a.X)*fraction,
					Y: a.Y + (b.Y-a.Y)*fraction,
					Z: a.Z + (b.Z-a.Z)*fraction,
				}
			}
		}
		cosX, sinX := math.Cos(rotation.X), math.Sin(rotation.X)
		cosY, sinY := math.Cos(rotation.Y), math.Sin(rotation.Y)
		cosZ, sinZ := math.Cos(rotation.Z), math.Sin(rotation.Z)
		transformed := make([]Vec3, len(current))
		projected := make([]ProjectedPoint, len(current))
		for i, v := range current {
			x, y, z := v.X, v.Y, v.Z
			newY := y*cosX - z*sinX
			newZ := y*sinX + z*cosX
			y, z = newY, newZ
			newX := x*cosY + z*sinY
			newZ = -x*sinY + z*cosY
			x, z = newX, newZ
			newX = x*cosZ - y*sinZ
			newY = x*sinZ + y*cosZ
			x, y = newX, newY
			transformed[i] = Vec3{X: x, Y: y, Z: z}
			cameraZ := z + 1000
			if cameraZ <= 0 {
				cameraZ = 1
			}
			scale := float32(900 / cameraZ)
			projected[i] = ProjectedPoint{X: float32(x)*scale + 320, Y: float32(y)*scale + 240}
		}
		sorted := make([]SortedMorphFace, len(faces))
		for i, f := range faces {
			v := f.Vertices
			sorted[i] = SortedMorphFace{Index: i, Depth: (transformed[v[0]].Z + transformed[v[1]].Z + transformed[v[2]].Z) / 3.0}
		}
		slices.SortFunc(sorted, func(a, b SortedMorphFace) int { return cmp.Compare(a.Depth, b.Depth) })
		mesh.Step()
		gotShape, gotTarget := mesh.Shape()
		if mesh.Tick() != tick || mesh.Timer() != timer || gotShape != shape || gotTarget != target || mesh.Rotation() != rotation {
			t.Fatalf("tick %d state: timer=%v shape=%d target=%d rotation=%+v", tick, mesh.Timer(), gotShape, gotTarget, mesh.Rotation())
		}
		for i := range current {
			if mesh.Current()[i] != current[i] || mesh.Transformed()[i] != transformed[i] || mesh.Projected()[i] != projected[i] {
				t.Fatalf("tick %d vertex %d changed: current=%+v transformed=%+v projected=%+v", tick, i, mesh.Current()[i], mesh.Transformed()[i], mesh.Projected()[i])
			}
		}
		for i := range sorted {
			if mesh.SortedFaces()[i] != sorted[i] {
				t.Fatalf("tick %d sorted face %d = %+v, want %+v", tick, i, mesh.SortedFaces()[i], sorted[i])
			}
		}
	}
	if got := testing.AllocsPerRun(100, mesh.Step); got != 0 {
		t.Fatalf("morphing mesh step allocated %.2f objects", got)
	}
	mesh.Reset()
	if mesh.Tick() != 0 || mesh.Timer() != 0 {
		t.Fatalf("reset left tick=%d timer=%v", mesh.Tick(), mesh.Timer())
	}
}

func TestMorphingMeshRejectsUnequalShapesAndInvalidFaces(t *testing.T) {
	base := MorphingMeshConfig{Shapes: [][]Vec3{{{X: 1}}, {{X: 2}}},
		Faces: []MorphFace{{Vertices: [3]int{0, 0, 0}}}, MorphTicks: 1, CycleTicks: 2, Focal: 1}
	if _, err := NewMorphingMesh(base); err != nil {
		t.Fatal(err)
	}
	base.Shapes[1] = append(base.Shapes[1], Vec3{})
	if _, err := NewMorphingMesh(base); err == nil {
		t.Fatal("accepted unequal shapes")
	}
	base.Shapes[1] = base.Shapes[1][:1]
	base.Faces[0].Vertices[1] = 1
	if _, err := NewMorphingMesh(base); err == nil {
		t.Fatal("accepted an out-of-range face index")
	}
}

func BenchmarkMorphingMesh14Vertices24Faces(b *testing.B) {
	shapes := make([][]Vec3, 3)
	for i := range shapes {
		shapes[i] = make([]Vec3, 14)
		for j := range shapes[i] {
			shapes[i][j] = Vec3{X: float64(j*17 - 90), Y: float64(i*40 + j*7 - 80), Z: float64(j*19 - 120)}
		}
	}
	faces := make([]MorphFace, 24)
	for i := range faces {
		faces[i].Vertices = [3]int{i % 14, (i + 3) % 14, (i + 8) % 14}
	}
	mesh, err := NewMorphingMesh(MorphingMeshConfig{Shapes: shapes, Faces: faces,
		MorphTicks: 120, CycleTicks: 240, RotationStep: Vec3{X: .01, Y: .02, Z: .04},
		CenterX: 320, CenterY: 240, Focal: 900, CameraZ: 1000})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		mesh.Step()
	}
}
