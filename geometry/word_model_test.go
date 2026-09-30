package geometry

import (
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestWordProjectionMatchesIndependentWrappedArithmetic(t *testing.T) {
	// Wide integers form every product; conversion back to long/word widths is
	// separate from the implementation and includes overflowing transforms.
	pose := WordPose{Matrix: [9]int16{32000, -17000, 7000, -28000, 23000, -12000, 16000, -30000, 26000}, Translation: [2]int16{-190, 310}, Depth: 1540}
	c := WordProjectionConfig{DepthShift: 9, DepthBias: 512, TranslationShift: 8, Center: [2]int16{175, 136}}
	wrap32 := func(n *big.Int) int32 { return int32(uint32(n.Int64())) }
	for i := 0; i < 65536; i++ {
		p := motion.WrappedPoint{X: int16(i), Y: int16(i*79 + 137), Z: int16(i*17 - 400)}
		var axes [3]int32
		for axis := range axes {
			total := new(big.Int)
			for column, v := range [3]int16{p.X, p.Y, p.Z} {
				product := new(big.Int).Mul(big.NewInt(int64(v)), big.NewInt(int64(pose.Matrix[column*3+axis])))
				total.Add(total, product)
			}
			axes[axis] = wrap32(total)
		}
		depth := int16(uint16(int16(axes[2]>>9)) + uint16(c.DepthBias) + uint16(pose.Depth))
		got, err := ProjectWordPoint(p, pose, c)
		if depth == 0 {
			if err == nil {
				t.Fatal("accepted zero depth")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		want := [2]int16{}
		for axis := range want {
			n := int32(uint32(axes[axis]) + uint32(int32(pose.Translation[axis])<<8))
			q := int64(n) / int64(depth)
			word := int16(q)
			if q < math.MinInt16 || q > math.MaxInt16 {
				word = int16(n)
			}
			want[axis] = int16(uint16(word) + uint16(c.Center[axis]))
		}
		if !got.Visible || got.X != want[0] || got.Y != want[1] {
			t.Fatalf("word %d: got%+v want%v", i, got, want)
		}
	}
}

func TestWordModelCopiesCachesAndRejectsPoseAtomically(t *testing.T) {
	points := []motion.WrappedPoint{{X: 2, Y: 3}, {X: 4, Y: 5}}
	m, err := NewWordModel(WordModelConfig{Points: points})
	if err != nil {
		t.Fatal(err)
	}
	points[0].X = 99
	p := WordPose{Matrix: [9]int16{1, 0, 0, 0, 1, 0, 0, 0, 1}, Depth: 1}
	if err := m.Project(p); err != nil {
		t.Fatal(err)
	}
	before := append([]WordScreenPoint(nil), m.Points()...)
	if before[0].X != 2 || before[0].Y != 3 {
		t.Fatal("input geometry stayed borrowed")
	}
	if n := testing.AllocsPerRun(100, func() {
		if err := m.Project(p); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatal("cached pose allocates", n)
	}
	p.Depth = 0
	if m.Project(p) == nil {
		t.Fatal("accepted failed pose")
	}
	if !reflect.DeepEqual(m.Points(), before) {
		t.Fatal("failed pose mutated the visible bank")
	}
	c := WordProjectionConfig{ZeroDepth: WordZeroDepthHide}
	v, err := ProjectWordPoint(motion.WrappedPoint{}, WordPose{}, c)
	if err != nil || v.Visible {
		t.Fatal("zero-depth hide failed")
	}
	c = WordProjectionConfig{Overflow: WordDivisionClamp}
	v, err = ProjectWordPoint(motion.WrappedPoint{X: 32767}, WordPose{Matrix: [9]int16{32767}, Depth: 1}, c)
	if err != nil || v.X != 32767 {
		t.Fatal("clamped quotient failed", v, err)
	}
	c.Overflow = WordDivisionReject
	if _, err = ProjectWordPoint(motion.WrappedPoint{X: 32767}, WordPose{Matrix: [9]int16{32767}, Depth: 1}, c); err == nil {
		t.Fatal("accepted overflowing quotient")
	}
	for _, cfg := range []WordProjectionConfig{{DepthShift: 32}, {TranslationShift: 32}, {Overflow: 3}, {ZeroDepth: 2}} {
		if _, err := NewWordModel(WordModelConfig{Points: []motion.WrappedPoint{{}}, Projection: cfg}); err == nil {
			t.Fatal("accepted invalid projection", cfg)
		}
	}
}

func TestWordWindingWrapsSubtractionBeforeProduct(t *testing.T) {
	a, b, c := WordScreenPoint{X: -32768, Y: 32000}, WordScreenPoint{X: 32767, Y: -32000}, WordScreenPoint{X: 12000, Y: 15000}
	x1, y1, x2, y2 := int16(uint16(b.X)-uint16(a.X)), int16(uint16(b.Y)-uint16(a.Y)), int16(uint16(c.X)-uint16(a.X)), int16(uint16(c.Y)-uint16(a.Y))
	want := int32(uint32(int64(x1)*int64(y2)) - uint32(int64(y1)*int64(x2)))
	if got := WordWinding(a, b, c); got != want {
		t.Fatal(got, want)
	}
}
