package effects

import (
	"math"
	"testing"
)

func TestSolidSphereModelPreservesAuthoredFaceOrderAndPoleBand(t *testing.T) {
	model, err := SolidSphereModel(SolidSphereConfig{
		Radius: 80, Rows: 8, Columns: 16, PoleBand: true,
		BodyColors: [2]uint32{0xffffff, 0xff0000},
		PoleColors: [2]uint32{0x00ff00, 0x0000ff},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Points) != 512 || len(model.Groups) != 1 || len(model.Groups[0]) != 112 {
		t.Fatalf("sphere has %d points and %d face groups", len(model.Points), len(model.Groups))
	}
	faces := model.Groups[0]
	if faces[0] != (SolidFace{Indices: [4]int{64, 65, 66, 67}, Color: 0xff0000}) ||
		faces[95] != (SolidFace{Indices: [4]int{444, 445, 446, 447}, Color: 0xff0000}) ||
		faces[96] != (SolidFace{Indices: [4]int{1, 2, 452, 448}, Color: 0x00ff00}) ||
		faces[111] != (SolidFace{Indices: [4]int{61, 1, 448, 508}, Color: 0x0000ff}) {
		t.Fatalf("sphere changed its ordered body or pole faces: %v, %v, %v, %v", faces[0], faces[95], faces[96], faces[111])
	}
	if model.Points[3] != model.Points[2] {
		t.Fatal("first polar corner is not paired with its neighboring vertex")
	}
	if math.Abs(model.Points[0].Z-80) > 1e-12 {
		t.Fatalf("south pole Z = %v, want 80", model.Points[0].Z)
	}
}

func TestSolidSphereModelSupportsOrdinaryGridAndLongitudeOffset(t *testing.T) {
	model, err := SolidSphereModel(SolidSphereConfig{
		Radius: 10, Rows: 4, Columns: 8, LongitudeOffset: 90,
		BodyColors: [2]uint32{0x112233, 0x445566},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Points) != 128 || len(model.Groups[0]) != 32 {
		t.Fatalf("ordinary sphere has %d points and %d faces", len(model.Points), len(model.Groups[0]))
	}
	point := model.Points[8*4]
	if math.Abs(point.X) > 1e-12 || math.Abs(point.Y-10/math.Sqrt2) > 1e-12 ||
		math.Abs(point.Z-10/math.Sqrt2) > 1e-12 {
		t.Fatalf("longitude offset produced %+v", point)
	}
	if model.Groups[0][0].Color != 0x112233 || model.Groups[0][1].Color != 0x445566 {
		t.Fatal("checker colors did not alternate")
	}
}

func TestSolidSphereModelRejectsUnboundedOrInvalidGeometry(t *testing.T) {
	valid := SolidSphereConfig{Radius: 10, Rows: 8, Columns: 16}
	for _, edit := range []func(*SolidSphereConfig){
		func(c *SolidSphereConfig) { c.Radius = 0 },
		func(c *SolidSphereConfig) { c.Radius = math.NaN() },
		func(c *SolidSphereConfig) { c.Rows = 2 },
		func(c *SolidSphereConfig) { c.Columns = 65 },
		func(c *SolidSphereConfig) { c.LongitudeOffset = math.Inf(1) },
		func(c *SolidSphereConfig) { c.BodyColors[1] = 0x1000000 },
	} {
		config := valid
		edit(&config)
		if _, err := SolidSphereModel(config); err == nil {
			t.Fatalf("accepted invalid sphere configuration %+v", config)
		}
	}
}
