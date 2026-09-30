package motion

import (
	"math"
	"testing"
)

func wordVelocitySines() []int16 {
	sines := make([]int16, 360)
	for i := range sines {
		sines[i] = int16(math.Round(math.Sin(float64(i)*math.Pi/180) * 32767))
	}
	return sines
}

// This reference retains matrix products in wide arithmetic and explicitly
// converts their low 32-bit words before taking signed high 16-bit results.
func wordVelocityMatrixReference(wave [6]int16, shifts [3]uint8) [3]int16 {
	mul := func(a, b int16) int64 { return int64(a) * int64(b) }
	wordHigh := func(value int64) int16 {
		return int16(int32(uint32(value)) >> 16)
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	x := wordHigh(mul(cx, cy))
	innerY := wordHigh(mul(sz, sy) * 2)
	y := wordHigh(mul(sx, cz) - mul(innerY, cx))
	innerZ := wordHigh(mul(cz, sy) * 2)
	z := wordHigh(mul(sx, sz) + mul(innerZ, cx))
	return [3]int16{int16(int32(x) >> shifts[0]), int16(int32(y) >> shifts[1]), int16(int32(z) >> shifts[2])}
}

func TestWordEulerVelocityMatchesNormalizedAndGuardedNativeEquations(t *testing.T) {
	sines := wordVelocitySines()
	for _, policy := range []WordVelocityPhase{WordPhaseNormalized, WordPhaseGuarded} {
		config := WordEulerVelocityConfig{Sines: sines, Period: 720, Quantum: 2, Quarter: 180,
			OutputShift: [3]uint8{11, 11, 9}, PhasePolicy: policy}
		if policy == WordPhaseGuarded {
			config.OutputShift = [3]uint8{10, 10, 8}
		}
		velocity, err := NewWordEulerVelocity(config)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 65536; i++ {
			angles := [3]int16{int16(i), int16(i*13 + 7), int16(i*29 + 11)}
			var wave [6]int16
			valid := true
			for axis, angle := range angles {
				if policy == WordPhaseNormalized {
					phase := int(angle) % 720
					if phase < 0 {
						phase += 720
					}
					n := phase &^ 1
					wave[axis*2] = sines[n/2]
					wave[axis*2+1] = sines[((n+180)%720)/2]
				} else {
					index := int(uint16(angle)&0xfffe) / 2
					if index < len(sines) {
						wave[axis*2] = sines[index]
					}
					quarter := uint16(angle) + 180
					if int16(quarter) > 718 {
						quarter -= 720
					} else if int16(quarter) < 0 {
						quarter += 720
					}
					index = int(quarter&0xfffe) / 2
					if index >= len(sines) {
						valid = false
						break
					}
					wave[axis*2+1] = sines[index]
				}
			}
			got, ok := velocity.Sample(angles)
			if ok != valid {
				t.Fatalf("policy %d angle case %d: validity %v/%v", policy, i, ok, valid)
			}
			if valid && got != wordVelocityMatrixReference(wave, config.OutputShift) {
				t.Fatalf("policy %d angles %v: matrix words %v differ", policy, angles, got)
			}
		}
	}
}

func TestWordEulerVelocityGuardNegativeOneAndModuloBoundaries(t *testing.T) {
	sines := wordVelocitySines()
	config := WordEulerVelocityConfig{Sines: sines, Period: 720, Quantum: 2, Quarter: 180,
		OutputShift: [3]uint8{10, 10, 8}, PhasePolicy: WordPhaseGuarded}
	guarded, err := NewWordEulerVelocity(config)
	if err != nil {
		t.Fatal(err)
	}
	sine, cosine, ok := guarded.phase(-1)
	if !ok || sine != 0 || cosine != sines[89] {
		t.Fatal("unsigned negative-one sine guard or odd quarter quantization changed", sine, cosine, ok)
	}
	for _, angles := range [][3]int16{{-1, 180, 0}, {0, -1, 0}, {0, 180, -1}} {
		if _, ok := guarded.Sample(angles); !ok {
			t.Fatal("negative-one guard should remain a valid native pose", angles)
		}
	}
	if _, _, ok := guarded.phase(539); ok {
		t.Fatal("single quarter correction producing an invalid guard address was normalized")
	}
	config.PhasePolicy = WordPhaseNormalized
	normalized, err := NewWordEulerVelocity(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []int16{-32768, -721, -720, -719, -1, 0, 1, 719, 720, 721, 32767} {
		n := int(phase) % 720
		if n < 0 {
			n += 720
		}
		n &^= 1
		sine, cosine, ok := normalized.phase(phase)
		if !ok || sine != sines[n/2] || cosine != sines[((n+180)%720)/2] {
			t.Fatal("normalized period boundary changed", phase, sine, cosine, ok)
		}
	}
}

func TestWordEulerVelocityCustomTableAndLongOverflow(t *testing.T) {
	// Extreme signed words force long addition and doubled-product wrapping,
	// independently of a smooth physical sine table.
	table := []int16{-32768, 32767, -32767, 32766}
	config := WordEulerVelocityConfig{Sines: table, Period: 12, Quantum: 3, Quarter: 3,
		OutputShift: [3]uint8{0, 1, 31}}
	velocity, err := NewWordEulerVelocity(config)
	if err != nil {
		t.Fatal(err)
	}
	table[0] = 0
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			for c := 0; c < 4; c++ {
				original := []int16{-32768, 32767, -32767, 32766}
				wave := [6]int16{original[a], original[(a+1)%4], original[b], original[(b+1)%4], original[c], original[(c+1)%4]}
				got, ok := velocity.Sample([3]int16{int16(a*3 + 2), int16(b*3 + 2), int16(c*3 + 2)})
				if !ok || got != wordVelocityMatrixReference(wave, config.OutputShift) {
					t.Fatal("custom quantum, copied bank or long overflow changed", got, ok)
				}
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { velocity.Sample([3]int16{1, 2, 3}) }); allocations != 0 {
		t.Fatal("word velocity allocated", allocations)
	}
}

func TestWordEulerVelocityValidation(t *testing.T) {
	valid := WordEulerVelocityConfig{Sines: []int16{0}, Period: 1, Quantum: 1}
	for _, invalid := range []WordEulerVelocityConfig{
		{}, {Sines: []int16{0}, Period: 1},
		{Sines: []int16{0}, Period: 32769, Quantum: 1},
		{Sines: []int16{0}, Period: 3, Quantum: 2},
		{Sines: []int16{0}, Period: 1, Quantum: 1, Quarter: 1},
		{Sines: []int16{0}, Period: 1, Quantum: 1, OutputShift: [3]uint8{32}},
		{Sines: []int16{0}, Period: 1, Quantum: 1, PhasePolicy: 2},
		{Sines: []int16{0, 1}, Period: 1, Quantum: 1},
	} {
		if _, err := NewWordEulerVelocity(invalid); err == nil {
			t.Fatal("invalid word velocity configuration accepted", invalid)
		}
	}
	if _, err := NewWordEulerVelocity(valid); err != nil {
		t.Fatal(err)
	}
	var absent *WordEulerVelocity
	if _, ok := absent.Sample([3]int16{}); ok {
		t.Fatal("nil velocity controller accepted a pose")
	}
}
