package sprites

import "github.com/olivierh59500/democonstructionkit/geometry"

// SphereConfig produces an evenly spaced vectorball surface. Count is the
// number of balls, not a mesh subdivision count. Use geometry.SphereCloud and
// ProjectedObjectConfig.Points for random or equal-area sampled clouds.
type SphereConfig struct {
	Radius float64
	Count  int
	Center geometry.Vec3
	Image  int
}

func Sphere(config SphereConfig) ([]Point, error) {
	positions, err := geometry.SphereCloud(geometry.SphereCloudConfig{
		Count: config.Count, Radius: config.Radius, Center: config.Center,
		Sampling: geometry.SphereFibonacci,
	})
	if err != nil {
		return nil, err
	}
	points := make([]Point, len(positions))
	for index, position := range positions {
		points[index] = Point{X: position.X, Y: position.Y, Z: position.Z, Image: config.Image}
	}
	return points, nil
}
