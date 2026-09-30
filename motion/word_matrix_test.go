package motion

import (
	"math"
	"testing"
)

func TestWordMatrixMatchesIndependentOrderedProducts(t *testing.T) {
	sines := make([]int16, 2048)
	for i := range sines {
		sines[i] = int16(math.Round(math.Sin(float64(i)*2*math.Pi/2048) * 32767))
	}
	for _, middle := range []WordMiddleAssociation{WordMiddleYFirst, WordMiddleXFirst} {
		cfg := WordEulerMatrixConfig{Sines: sines, Period: 4096, Quantum: 2, Quarter: 1024, ProductShift: 6, ThirdAxisShift: 7, MiddleAssociation: middle}
		matrix, err := NewWordEulerMatrix(cfg)
		if err != nil {
			t.Fatal(err)
		}
		for tick := 0; tick < 65536; tick++ {
			angles := [3]int16{int16(tick), int16(tick*7 + 913), int16(tick*13 - 171)}
			var w [6]int64
			for axis, a := range angles {
				n := (int(a) & 4095) &^ 1
				w[axis*2] = int64(sines[n/2])
				w[axis*2+1] = int64(sines[((n+1024)&4095)/2])
			}
			sx, cx, sy, cy, sz, cz := w[0], w[1], w[2], w[3], w[4], w[5]
			high := func(n int64) int64 { return int64(int16(int32(n) >> 16)) }
			scaled := func(n int64) int16 { return int16(high(n) >> 6) }
			term := high(sz*sy*2) * sx
			if middle == WordMiddleXFirst {
				term = high(sz*sx*2) * sy
			}
			want := [9]int16{scaled(cx * cy), scaled(sx*cz - high(sz*sy*2)*cx), scaled(sx*sz + high(cz*sy*2)*cx), -scaled(sx * cy), scaled(cx*cz + term), scaled(cx*sz - high(cz*sy*2)*sx), -int16(sy >> 7), -scaled(cy * sz), scaled(cy * cz)}
			got, ok := matrix.Sample(angles)
			if !ok || got != want {
				t.Fatalf("tick%d middle%d got%v want%v", tick, middle, got, want)
			}
		}
	}
}

func TestWordMatrixOwnershipPoliciesAndAllocations(t *testing.T) {
	c := WordEulerMatrixConfig{Sines: []int16{0, 32767, 0, -32767}, Period: 4, Quantum: 1, Quarter: 1, ProductShift: 6, ThirdAxisShift: 7}
	m, err := NewWordEulerMatrix(c)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := m.Sample([3]int16{})
	c.Sines[1] = 0
	after, _ := m.Sample([3]int16{})
	if after != before {
		t.Fatal("sine bank not copied")
	}
	if n := testing.AllocsPerRun(100, func() { m.Sample([3]int16{1, 2, 3}) }); n != 0 {
		t.Fatal("matrix sampling allocates", n)
	}
	for _, bad := range []WordEulerMatrixConfig{{}, {Sines: []int16{1}, Period: 1, Quantum: 1, ProductShift: 32}, {Sines: []int16{1}, Period: 1, Quantum: 1, ThirdAxisShift: 16}, {Sines: []int16{1}, Period: 1, Quantum: 1, MiddleAssociation: 2}} {
		if _, err := NewWordEulerMatrix(bad); err == nil {
			t.Fatal("invalid config", bad)
		}
	}
}
