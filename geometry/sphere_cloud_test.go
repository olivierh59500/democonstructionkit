package geometry

import (
	"math"
	"testing"
)

func TestSphereCloudRandomAnglesConsumesPairedSamplesInOrder(t *testing.T) {
	samples := []float64{0, .25, .5, .75}
	nextIndex := 0
	points, err := SphereCloud(SphereCloudConfig{
		Count: 2, Radius: 10, Sampling: SphereRandomAngles,
		NextFloat: func() float64 {
			value := samples[nextIndex]
			nextIndex++
			return value
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if nextIndex != 4 || len(points) != 2 {
		t.Fatalf("used %d random samples for %d points", nextIndex, len(points))
	}
	if math.Abs(points[0].X) > 1e-12 || math.Abs(points[0].Y) > 1e-12 ||
		math.Abs(points[0].Z-10) > 1e-12 ||
		math.Abs(points[1].X) > 1e-12 || math.Abs(points[1].Y+10) > 1e-12 ||
		math.Abs(points[1].Z) > 1e-12 {
		t.Fatalf("polar and azimuth samples produced %v and %v", points[0], points[1])
	}
}

func TestSphereCloudEqualAreaAndFibonacciModes(t *testing.T) {
	samples := []float64{.25, 0}
	index := 0
	equal, err := SphereCloud(SphereCloudConfig{Count: 1, Radius: 10,
		Sampling: SphereEqualArea, NextFloat: func() float64 { value := samples[index]; index++; return value }})
	if err != nil {
		t.Fatal(err)
	}
	if index != 2 || math.Abs(equal[0].X-5*math.Sqrt(3)) > 1e-12 ||
		math.Abs(equal[0].Y) > 1e-12 || math.Abs(equal[0].Z-5) > 1e-12 {
		t.Fatalf("equal-area point or sample count changed: %v, %d", equal[0], index)
	}
	spiral, err := SphereCloud(SphereCloudConfig{Count: 1, Radius: 10,
		Center: Vec3{X: 2, Y: 3, Z: 4}, Sampling: SphereFibonacci})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(spiral[0].X-12) > 1e-12 || spiral[0].Y != 3 || spiral[0].Z != 4 {
		t.Fatalf("single Fibonacci point = %v", spiral[0])
	}
}

func TestSphereCloudRejectsInvalidBoundsAndRandomValues(t *testing.T) {
	valid := SphereCloudConfig{Count: 4, Radius: 10, Sampling: SphereFibonacci}
	for _, edit := range []func(*SphereCloudConfig){
		func(c *SphereCloudConfig) { c.Count = 0 },
		func(c *SphereCloudConfig) { c.Count = 4097 },
		func(c *SphereCloudConfig) { c.Radius = math.NaN() },
		func(c *SphereCloudConfig) { c.Center.Z = math.Inf(1) },
		func(c *SphereCloudConfig) { c.Sampling = 99 },
		func(c *SphereCloudConfig) { c.Sampling = SphereRandomAngles },
		func(c *SphereCloudConfig) { c.Sampling = SphereEqualArea; c.NextFloat = func() float64 { return 2 } },
	} {
		config := valid
		edit(&config)
		if _, err := SphereCloud(config); err == nil {
			t.Fatalf("accepted invalid sphere cloud configuration %+v", config)
		}
	}
}
