package motion

import (
	"math"
	"math/big"
	"testing"
)

func TestTablePolarNativeIntegerProjection(t *testing.T) {
	wave := make([]int16, 256)
	for i := range wave {
		wave[i] = int16(math.Round(127 * math.Sin(float64(i)*2*math.Pi/256)))
	}
	mapping, err := NewTablePolar(TablePolarConfig{
		Wave: wave, Shift: 7, Center: [2]int64{173, 108}, Phases: [2]int{0, 64}, WordBits: 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Keep the source equations independent: byte phases and int16 coordinate
	// writes are compared against every angle and radius in the authored range.
	for angle := 0; angle < 256; angle++ {
		for radius := 0; radius <= 126; radius++ {
			want := FixedPoint{
				X: float64(int16(int(wave[byte(0-byte(angle))])*radius>>7) + 173),
				Y: float64(int16(int(wave[byte(64-byte(angle))])*radius>>7) + 108),
			}
			got, ok := mapping.Point(int64(angle), int64(radius))
			if !ok || got != want {
				t.Fatalf("angle %d radius %d: %v, %v; want %v", angle, radius, got, ok, want)
			}
		}
	}
}

func TestRationalGridNativePerspectiveTable(t *testing.T) {
	mapping, err := NewRationalGrid(RationalGridConfig{
		X: [3]int64{-81920, -4096, 1536}, Y: [3]int64{71680, -2560, -819},
		Denominator: [3]int64{330, 10, 5}, Center: [2]int64{173, 108}, WordBits: 16,
	})
	if err != nil {
		t.Fatal(err)
	}
	for row := 0; row < 11; row++ {
		for column := 0; column < 280; column++ {
			denominator := 330 + row*10 + column*5
			x := -81920 - row*4096 + column*1536
			y := 71680 - row*2560 - column*819
			want := FixedPoint{X: float64(int16(x/denominator) + 173), Y: float64(int16(y/denominator) + 108)}
			got, ok := mapping.Point(int64(column), int64(row))
			if !ok || got != want {
				t.Fatalf("column %d row %d: %v, %v; want %v", column, row, got, ok, want)
			}
		}
	}
}

func TestTablePolarCopiedWaveAndSignedShift(t *testing.T) {
	wave := []int16{3, -3, 7, -7}
	mapping, err := NewTablePolar(TablePolarConfig{
		Wave: wave, Shift: 1, Phases: [2]int{-1, 1}, Center: [2]int64{10, -5},
	})
	if err != nil {
		t.Fatal(err)
	}
	wave[3] = 99
	got, ok := mapping.Point(0, 1)
	if !ok || got != (FixedPoint{X: 6, Y: -7}) {
		t.Fatalf("copied negative samples: %v, %v", got, ok)
	}
	// Signed angular input wraps by the table length, even at int64 bounds.
	for _, angle := range []int64{math.MinInt64, math.MaxInt64, -100, 104} {
		got, ok = mapping.Point(angle, -2)
		wrapped, valid := mapping.Point(angle&3, -2)
		if !ok || !valid || got != wrapped {
			t.Fatalf("phase %d: %v versus %v", angle, got, wrapped)
		}
	}
}

func TestContourMappingsSignedWords(t *testing.T) {
	for bits := uint8(1); bits <= 32; bits++ {
		center := int64(1 << (bits - 1))
		table, err := NewTablePolar(TablePolarConfig{
			Wave: []int16{1}, Center: [2]int64{center, -center}, WordBits: bits,
		})
		if err != nil {
			t.Fatal(err)
		}
		grid, err := NewRationalGrid(RationalGridConfig{
			X: [3]int64{1 << bits}, Y: [3]int64{1 << bits}, Denominator: [3]int64{1},
			Center: [2]int64{center, -center}, WordBits: bits,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := FixedPoint{X: -float64(center), Y: -float64(center)}
		got, ok := table.Point(0, 1<<bits)
		projected, valid := grid.Point(0, 0)
		if !ok || !valid || got != want || projected != want {
			t.Fatalf("signed %d-bit wrapping: table=%v grid=%v; want %v", bits, got, projected, want)
		}
	}
}

func TestRationalGridSignedDivisionAndSingularity(t *testing.T) {
	mapping, err := NewRationalGrid(RationalGridConfig{
		X: [3]int64{-10, 2, 3}, Y: [3]int64{10, -3, -2}, Denominator: [3]int64{-3, 0, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, ok := mapping.Point(0, 0)
	if !ok || got != (FixedPoint{X: 3, Y: -3}) {
		t.Fatalf("negative denominator: %v, %v", got, ok)
	}
	got, ok = mapping.Point(1, 2)
	if !ok || got != (FixedPoint{X: 1, Y: -1}) {
		t.Fatalf("row/column coefficients or signed truncation: %v, %v", got, ok)
	}
	if point, valid := mapping.Point(3, 0); valid || point != (FixedPoint{}) {
		t.Fatalf("zero denominator: %v, %v", point, valid)
	}
}

func TestContourMappingValidation(t *testing.T) {
	for _, config := range []TablePolarConfig{
		{}, {Wave: []int16{1, 2, 3}}, {Wave: make([]int16, 131072)},
		{Wave: []int16{1}, Shift: 32}, {Wave: []int16{1}, WordBits: 33},
		{Wave: []int16{1}, Center: [2]int64{1<<53 + 1}},
	} {
		if _, err := NewTablePolar(config); err == nil {
			t.Fatalf("accepted invalid polar configuration: %+v", config)
		}
	}
	for _, config := range []RationalGridConfig{
		{}, {Denominator: [3]int64{1}, WordBits: 33},
		{Denominator: [3]int64{1}, Center: [2]int64{0, -(1<<53 + 1)}},
	} {
		if _, err := NewRationalGrid(config); err == nil {
			t.Fatalf("accepted invalid rational configuration: %+v", config)
		}
	}
	var table *TablePolar
	var grid *RationalGrid
	if _, ok := table.Point(0, 0); ok {
		t.Fatal("nil table accepted")
	}
	if _, ok := grid.Point(0, 0); ok {
		t.Fatal("nil grid accepted")
	}
	if _, ok := (&TablePolar{}).Point(0, 0); ok {
		t.Fatal("empty table accepted")
	}
	if _, ok := (&RationalGrid{}).Point(0, 0); ok {
		t.Fatal("empty grid accepted")
	}
}

func TestContourMappingOverflowIsRejected(t *testing.T) {
	polar, err := NewTablePolar(TablePolarConfig{Wave: []int16{-1, 2}, Shift: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]int64{{0, math.MinInt64}, {1, math.MaxInt64}} {
		if _, ok := polar.Point(pair[0], pair[1]); ok {
			t.Fatalf("polar multiplication overflow accepted: %v", pair)
		}
	}
	for _, config := range []RationalGridConfig{
		{X: [3]int64{0, 2}, Denominator: [3]int64{1}},
		{X: [3]int64{math.MaxInt64, 1}, Denominator: [3]int64{1}},
		{X: [3]int64{math.MinInt64}, Denominator: [3]int64{-1}},
		{X: [3]int64{1<<53 + 1}, Denominator: [3]int64{1}},
		{X: [3]int64{1 << 53}, Denominator: [3]int64{1}, Center: [2]int64{1}},
		{X: [3]int64{1}, Denominator: [3]int64{math.MaxInt64, 1}},
	} {
		grid, err := NewRationalGrid(config)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := grid.Point(0, math.MaxInt64); ok {
			t.Fatalf("rational overflow or precision loss accepted: %+v", config)
		}
	}
	grid, err := NewRationalGrid(RationalGridConfig{
		X: [3]int64{1 << 53}, Y: [3]int64{-(1 << 53)}, Denominator: [3]int64{1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := grid.Point(0, 0); !ok || got != (FixedPoint{X: 1 << 53, Y: -(1 << 53)}) {
		t.Fatalf("exact integer boundary rejected: %v, %v", got, ok)
	}
}

func TestContourArithmeticAgainstUnboundedIntegers(t *testing.T) {
	values := []int64{math.MinInt64, math.MinInt64 + 1, -1 << 32, -32768, -2, -1, 0, 1, 2, 32767, 1 << 32, math.MaxInt64 - 1, math.MaxInt64}
	for _, a := range values {
		for _, b := range values {
			product := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
			value, ok := contourMultiply(a, b)
			if ok != product.IsInt64() || ok && value != product.Int64() {
				t.Fatalf("multiply %d * %d = %d, %v; want %v", a, b, value, ok, product)
			}
			sum := new(big.Int).Add(big.NewInt(a), big.NewInt(b))
			value, ok = contourAdd(a, b)
			if ok != sum.IsInt64() || ok && value != sum.Int64() {
				t.Fatalf("add %d + %d = %d, %v; want %v", a, b, value, ok, sum)
			}
		}
	}
}

func TestContourMappingsDoNotAllocatePerPoint(t *testing.T) {
	table, err := NewTablePolar(TablePolarConfig{Wave: []int16{1, -1}})
	if err != nil {
		t.Fatal(err)
	}
	grid, err := NewRationalGrid(RationalGridConfig{X: [3]int64{1, 2, 3}, Denominator: [3]int64{1, 2, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if got := testing.AllocsPerRun(100, func() { table.Point(12, 9); grid.Point(12, 9) }); got != 0 {
		t.Fatalf("mapping allocated %v times per call", got)
	}
}
