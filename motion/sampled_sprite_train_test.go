package motion

import (
	"math"
	"testing"
)

func TestSampledSpriteTrainMatchesStarwarsThroughStrictWrap(t *testing.T) {
	const period = 64
	px, py := make([]float64, period), make([]float64, period)
	for i := range px {
		px[i] = math.Sin(float64(i)*.03)*100 + float64(i)/10
		py[i] = math.Cos(float64(i)*.05)*40 - float64(i)/7
	}
	train, err := NewSampledSpriteTrain(SampledSpriteTrainConfig{
		PathX: px, PathY: py, Count: 8, Spacing: 5,
		ExtraAfter: 3, ExtraOffset: 5, WrapAfterLength: true,
		XAmplitude: 10, YAmplitude: 5, XRate: .07, YRate: .09,
		XIndexPhase: 1, YIndexPhase: 1, YCos: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	legacyX, legacyY := append(append([]float64(nil), px...), px...),
		append(append([]float64(nil), py...), py...)
	cursor, wraps := 0, 0
	for tick := 0; tick < 5000; tick++ {
		cursor++
		if cursor > period {
			cursor = 0
			wraps++
		}
		train.Step()
		if train.Cursor() != cursor {
			t.Fatalf("tick %d cursor=%d, want %d", tick, train.Cursor(), cursor)
		}
		for i, pose := range train.Samples() {
			extra := 0
			if i >= 3 {
				extra = 5
			}
			index := cursor + i*5 + extra
			wantX := legacyX[index] + 10*math.Sin(float64(cursor)*.07+float64(i))
			wantY := legacyY[index] + 5*math.Cos(float64(cursor)*.09+float64(i))
			if pose != (SampledSpritePose{Index: i, Frame: i, SampleIndex: index, X: wantX, Y: wantY}) {
				t.Fatalf("tick %d sprite %d = %+v, want x=%v y=%v index=%d", tick, i, pose, wantX, wantY, index)
			}
		}
	}
	if wraps < 2 {
		t.Fatalf("only %d path wraps", wraps)
	}
	if got := testing.AllocsPerRun(100, train.Step); got != 0 {
		t.Fatalf("sampled sprite train step allocated %.2f objects", got)
	}
}

func BenchmarkSampledSpriteTrain8(b *testing.B) {
	px, py := make([]float64, 766), make([]float64, 766)
	for i := range px {
		px[i] = float64(i % 320)
		py[i] = float64(i % 200)
	}
	train, err := NewSampledSpriteTrain(SampledSpriteTrainConfig{
		PathX: px, PathY: py, Count: 8, Spacing: 5,
		ExtraAfter: 3, ExtraOffset: 5, WrapAfterLength: true,
		XAmplitude: 10, YAmplitude: 5, XRate: .07, YRate: .09,
		XIndexPhase: 1, YIndexPhase: 1, YCos: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		train.Step()
	}
}
