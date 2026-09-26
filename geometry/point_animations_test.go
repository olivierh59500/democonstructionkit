package geometry_test

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

type pointBuffer struct{ points []geometry.Vec3 }

func (b *pointBuffer) Len() int                          { return len(b.points) }
func (b *pointBuffer) XYZ(index int) geometry.Vec3       { return b.points[index] }
func (b *pointBuffer) SetXYZ(index int, p geometry.Vec3) { b.points[index] = p }

func TestSinusGridMatchesSourceRowsWithoutAllocating(t *testing.T) {
	grid, err := geometry.NewSinusGrid(geometry.SinusGridConfig{
		Columns: 8, Rows: 8, Amplitude: 200,
		SpatialNumerator: math.Pi, SpatialDivisor: 10,
		TravelStep: math.Pi / 55, EnvelopeStep: math.Pi / 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer := &pointBuffer{points: make([]geometry.Vec3, 64)}
	var phaseSin, phaseCos [15]float64
	for i := range phaseSin {
		phaseSin[i], phaseCos[i] = math.Sincos(float64(i) * math.Pi / 10)
	}
	travel, envelope := 0.0, 0.0
	for tick := 0; tick < 300; tick++ {
		amplitude := 200 * math.Sin(envelope)
		sinTravel, cosTravel := math.Sincos(travel)
		grid.Step(buffer)
		for row := 0; row < 8; row++ {
			for column := 0; column < 8; column++ {
				phase := row + column
				want := amplitude * (cosTravel*phaseCos[phase] - sinTravel*phaseSin[phase])
				if got := buffer.points[row*8+column].Z; got != want {
					t.Fatalf("tick %d point (%d,%d) Z = %v, want %v", tick, column, row, got, want)
				}
			}
		}
		travel += math.Pi / 55
		envelope += math.Pi / 60
	}
	if allocations := testing.AllocsPerRun(100, func() { grid.Step(buffer) }); allocations != 0 {
		t.Fatalf("sinus grid allocates %v per step", allocations)
	}
}

func TestRotorsAndYOrbitMatchSourceClocks(t *testing.T) {
	rotors, err := geometry.NewRotors(geometry.RotorConfig{
		Radii: []float64{80, 150, 210, 160}, Offset: 100, PhaseStep: .08, RequireFull: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	orbit, err := geometry.NewYOrbit(geometry.YOrbitConfig{
		Center: geometry.Vec3{Z: 850}, Radius: 100,
		PhaseStep: math.Pi / 144, RatioStep: .025, Min: 0, Max: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	buffer := &pointBuffer{points: make([]geometry.Vec3, 8)}
	for i := range buffer.points {
		buffer.points[i].Y = float64(i * 7)
	}
	phase, orbitPhase, ratio, ratioStep := 0.0, 0.0, 0.0, .025
	radii := [4]float64{80, 150, 210, 160}
	for tick := 0; tick < 250; tick++ {
		sinPhase, cosPhase := math.Sincos(phase)
		rotors.Step(buffer)
		for i, radius := range radii {
			first, second := buffer.points[i*2], buffer.points[i*2+1]
			x, z := radius*sinPhase, radius*cosPhase
			if first.X != x || first.Z != z-100 || second.X != -x || second.Z != -z-100 ||
				first.Y != float64(i*14) || second.Y != float64((i*2+1)*7) {
				t.Fatalf("tick %d rotor %d points %+v %+v", tick, i, first, second)
			}
		}
		phase += .08
		orbitPhase += math.Pi / 144
		if ratioStep > 0 {
			ratio += ratioStep
			if ratio >= 1 {
				ratioStep = 0
			}
		}
		sinOrbit, cosOrbit := math.Sincos(orbitPhase)
		want := geometry.Vec3{X: ratio * 100 * cosOrbit, Z: 850 + ratio*100*sinOrbit}
		if got := orbit.Step(); got != want || orbit.Pose() != want {
			t.Fatalf("tick %d orbit %+v, want %+v", tick, got, want)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { rotors.Step(buffer); orbit.Step() }); allocations != 0 {
		t.Fatalf("rotor/orbit step allocates %v", allocations)
	}
}

func TestBounceCurveKeepsEveryAuthoredSample(t *testing.T) {
	bounce, err := geometry.NewBounceCurve(geometry.BounceCurveConfig{
		Radius: 1200, Offset: 210,
		RadiusDecay: 5.7143 / 2.4, OffsetDecay: 1 / 2.4,
		StepDegrees: 3 / 2.4, BackwardStart: 89,
		FinalStepDegrees: 1, Loops: 3, Divisor: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	var want []float64
	radius, offset := 1200.0, 210.0
	appendSample := func(angle float64) {
		want = append(want, -(offset-radius*math.Cos(angle*math.Pi/180))/3)
		radius -= 5.7143 / 2.4
		offset -= 1 / 2.4
	}
	for cycle := 1; cycle < 4; cycle++ {
		for angle := 0.0; angle < 90; angle += 3 / 2.4 {
			appendSample(angle)
		}
		for angle := 89.0; angle >= 0; angle -= 3 / 2.4 {
			appendSample(angle)
		}
	}
	for angle := 0.0; angle < 90; angle++ {
		appendSample(angle)
	}
	if bounce.Len() != len(want) {
		t.Fatalf("bounce length = %d, want %d", bounce.Len(), len(want))
	}
	for tick := 0; tick < 2*len(want); tick++ {
		if got := bounce.Step(); got != want[tick%len(want)] || bounce.At() != got {
			t.Fatalf("bounce tick %d = %v, want %v", tick, got, want[tick%len(want)])
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { bounce.Step() }); allocations != 0 {
		t.Fatalf("bounce step allocates %v", allocations)
	}
}

func TestPointAnimationsRejectInvalidGeometry(t *testing.T) {
	if _, err := geometry.NewSinusGrid(geometry.SinusGridConfig{Rows: 8}); err == nil {
		t.Fatal("accepted a zero-width sinus grid")
	}
	if _, err := geometry.NewRotors(geometry.RotorConfig{Radii: []float64{math.NaN()}}); err == nil {
		t.Fatal("accepted a nonfinite rotor")
	}
	if _, err := geometry.NewYOrbit(geometry.YOrbitConfig{Max: -1}); err == nil {
		t.Fatal("accepted inverted orbit ratio bounds")
	}
	if _, err := geometry.NewBounceCurve(geometry.BounceCurveConfig{StepDegrees: 0, FinalStepDegrees: 1, Divisor: 3}); err == nil {
		t.Fatal("accepted a stalled bounce sweep")
	}
}
