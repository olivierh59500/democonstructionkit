package geometry

import (
	"fmt"
	"math"
)

// SphereSampling chooses a point distribution around the positive Z polar
// axis. RandomAngles reproduces a polar/azimuth angle draw; EqualArea spreads
// random samples evenly over the surface; Fibonacci is deterministic.
type SphereSampling uint8

const (
	SphereRandomAngles SphereSampling = iota
	SphereEqualArea
	SphereFibonacci
)

// SphereCloudConfig describes a static point population. NextFloat must return
// values in [0,1] and is called exactly twice per point for random modes. A
// caller-owned generator keeps seeded demos reproducible without coupling DCK
// to any particular random-number implementation.
type SphereCloudConfig struct {
	Count     int
	Radius    float64
	Center    Vec3
	Sampling  SphereSampling
	NextFloat func() float64
}

// SphereCloud builds one bounded point bank. Animation, projection and sprite
// material remain independent effects so the same bank can be reused.
func SphereCloud(config SphereCloudConfig) ([]Vec3, error) {
	if config.Count < 1 || config.Count > 4096 || config.Radius <= 0 ||
		!sphereCloudFinite(config.Radius, config.Center.X, config.Center.Y, config.Center.Z) ||
		config.Sampling > SphereFibonacci ||
		config.Sampling != SphereFibonacci && config.NextFloat == nil {
		return nil, fmt.Errorf("geometry: invalid sphere cloud configuration")
	}
	points := make([]Vec3, config.Count)
	for index := range points {
		var x, y, z float64
		switch config.Sampling {
		case SphereRandomAngles:
			polarSample, azimuthSample, err := sphereCloudSamples(config.NextFloat)
			if err != nil {
				return nil, err
			}
			polar, azimuth := math.Pi*polarSample, 2*math.Pi*azimuthSample
			x = config.Radius * math.Sin(polar) * math.Cos(azimuth)
			y = config.Radius * math.Sin(polar) * math.Sin(azimuth)
			z = config.Radius * math.Cos(polar)
		case SphereEqualArea:
			polarSample, azimuthSample, err := sphereCloudSamples(config.NextFloat)
			if err != nil {
				return nil, err
			}
			unitZ := 1 - 2*polarSample
			ring := math.Sqrt(math.Max(0, 1-unitZ*unitZ))
			azimuth := 2 * math.Pi * azimuthSample
			x = config.Radius * ring * math.Cos(azimuth)
			y = config.Radius * ring * math.Sin(azimuth)
			z = config.Radius * unitZ
		case SphereFibonacci:
			unitZ := 1 - 2*(float64(index)+.5)/float64(config.Count)
			ring := math.Sqrt(math.Max(0, 1-unitZ*unitZ))
			azimuth := float64(index) * math.Pi * (3 - math.Sqrt(5))
			x = config.Radius * ring * math.Cos(azimuth)
			y = config.Radius * ring * math.Sin(azimuth)
			z = config.Radius * unitZ
		}
		points[index] = Vec3{X: config.Center.X + x, Y: config.Center.Y + y, Z: config.Center.Z + z}
	}
	return points, nil
}

func sphereCloudSamples(next func() float64) (float64, float64, error) {
	first, second := next(), next()
	if !sphereCloudFinite(first, second) || first < 0 || first > 1 || second < 0 || second > 1 {
		return 0, 0, fmt.Errorf("geometry: sphere cloud sample outside [0,1]")
	}
	return first, second, nil
}

func sphereCloudFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
