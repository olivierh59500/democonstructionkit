package geometry

import (
	"math"
	"sort"
	"testing"
)

func TestRotatingDiscCloudMatchesSourceProjectionAndDepthOrder(t *testing.T) {
	seed := uint32(42)
	random := func() float64 {
		seed = seed*1664525 + 1013904223
		return float64(seed) / 4294967296
	}
	random() // The screen initializes its random generator once before assets.
	points := make([]Vec3, 125)
	for i := range points {
		a, b := math.Pi*random(), 2*math.Pi*random()
		points[i] = Vec3{X: 100 * math.Sin(a) * math.Cos(b),
			Y: 100 * math.Sin(a) * math.Sin(b), Z: 100 * math.Cos(a)}
	}
	focal := 138 / math.Tan(20*math.Pi/180)
	cloud, err := NewRotatingDiscCloud(RotatingDiscCloudConfig{
		Points: points, CenterX: 208, CenterY: 138, DepthBase: 900,
		YOffset: 16, Focal: focal, AngleStep: .03, RadiusScale: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	angle := 0.0
	for tick := 0; tick < 1000; tick++ {
		angle += .03
		sin, cos := math.Sincos(angle)
		want := make([]ProjectedDiscPose, len(points))
		for i, p := range points {
			x, z := p.X*cos+p.Z*sin, p.Z*cos-p.X*sin
			scale := focal / (900 - z)
			want[i] = ProjectedDiscPose{X: 208 + x*scale,
				Y: 138 - (p.Y+16)*scale, Z: z, Radius: scale}
		}
		sort.SliceStable(want, func(i, j int) bool { return want[i].Z < want[j].Z })
		if err := cloud.Step(); err != nil {
			t.Fatal(err)
		}
		if cloud.Angle() != angle || !cloud.Ready() {
			t.Fatalf("tick %d cloud angle %v, ready %v", tick, cloud.Angle(), cloud.Ready())
		}
		for i, got := range cloud.Poses() {
			if got != want[i] {
				t.Fatalf("tick %d disc %d = %+v, want %+v", tick, i, got, want[i])
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		if err := cloud.Step(); err != nil {
			t.Fatal(err)
		}
	}); allocations != 0 {
		t.Fatalf("disc cloud step allocated %v times", allocations)
	}
	previous := cloud.Angle()
	if err := cloud.SetAngleStep(.06); err != nil {
		t.Fatal(err)
	}
	if err := cloud.Step(); err != nil || cloud.Angle() != previous+.06 {
		t.Fatalf("live speed change reset the angle: %v, %v", cloud.Angle(), err)
	}
	cloud.Reset()
	if cloud.Angle() != 0 || cloud.Ready() {
		t.Fatal("disc cloud reset retained its previous pose")
	}
}

func TestRotatingDiscCloudKeepsLastPoseOnProjectionFailure(t *testing.T) {
	cloud, err := NewRotatingDiscCloud(RotatingDiscCloudConfig{
		Points: []Vec3{{Z: 1}}, DepthBase: 1, Focal: 1, RadiusScale: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cloud.Step(); err == nil || cloud.Angle() != 0 || cloud.Ready() {
		t.Fatal("camera-plane crossing changed the displayed state")
	}
}
