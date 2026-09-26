package effects

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// SolidSphereConfig builds a faceted sphere for SolidMeshCarousel or any other
// solid-mesh renderer. Rows and Columns are latitude and longitude intervals.
// Face colors alternate by row and column; PoleBand replaces the first and
// last latitude faces with one closing ring using PoleColors. LongitudeOffset
// rotates the source mesh in degrees without changing its animation clock.
type SolidSphereConfig struct {
	Radius          float64
	Rows, Columns   int
	LongitudeOffset float64
	BodyColors      [2]uint32
	PoleColors      [2]uint32
	PoleBand        bool
}

// SolidSphereModel emits four corners per latitude cell so every face can
// retain an independent material and the source face order stays stable.
// The model's polar axis is Z; rotation and projection belong to the renderer.
func SolidSphereModel(config SolidSphereConfig) (SolidMeshModel, error) {
	if config.Radius <= 0 || math.IsNaN(config.Radius) || math.IsInf(config.Radius, 0) ||
		math.IsNaN(config.LongitudeOffset) || math.IsInf(config.LongitudeOffset, 0) ||
		config.Rows < 3 || config.Rows > 64 || config.Columns < 3 || config.Columns > 64 {
		return SolidMeshModel{}, fmt.Errorf("effects: invalid solid sphere dimensions")
	}
	for _, shade := range [...]uint32{config.BodyColors[0], config.BodyColors[1], config.PoleColors[0], config.PoleColors[1]} {
		if shade > 0xffffff {
			return SolidMeshModel{}, fmt.Errorf("effects: solid sphere color exceeds RGB range")
		}
	}
	points := make([]geometry.Vec3, 0, config.Rows*config.Columns*4)
	faces := make([]SolidFace, 0, config.Rows*config.Columns)
	latitudeStep := 180 / float64(config.Rows)
	longitudeStep := 360 / float64(config.Columns)
	point := func(latitude, longitude float64) geometry.Vec3 {
		latitude *= math.Pi / 180
		longitude *= math.Pi / 180
		return geometry.Vec3{
			X: config.Radius * math.Cos(latitude) * math.Cos(longitude),
			Y: config.Radius * math.Cos(latitude) * math.Sin(longitude),
			Z: -config.Radius * math.Sin(latitude),
		}
	}
	for row := 0; row < config.Rows; row++ {
		latitude := -90 + float64(row)*latitudeStep
		for column := 0; column < config.Columns; column++ {
			longitude := config.LongitudeOffset + float64(column)*longitudeStep
			start := len(points)
			a := point(latitude, longitude)
			b := point(latitude+latitudeStep, longitude)
			c := point(latitude+latitudeStep, longitude+longitudeStep)
			d := point(latitude, longitude+longitudeStep)
			if config.PoleBand && row == 0 {
				d = c
			}
			points = append(points, a, b, c, d)
			if !config.PoleBand || row > 0 && row < config.Rows-1 {
				faces = append(faces, SolidFace{
					Indices: [4]int{start, start + 1, start + 2, start + 3},
					Color:   config.BodyColors[(row+column)&1],
				})
			}
		}
	}
	if config.PoleBand {
		south := (config.Rows - 1) * config.Columns * 4
		for column := 0; column < config.Columns; column++ {
			a, b := column*4+1, column*4+2
			c, d := south+(column+1)*4, south+column*4
			if column == config.Columns-1 {
				b, c = 1, south
			}
			faces = append(faces, SolidFace{
				Indices: [4]int{a, b, c, d},
				Color:   config.PoleColors[column&1],
			})
		}
	}
	return SolidMeshModel{Points: points, Groups: [][]SolidFace{faces}}, nil
}
