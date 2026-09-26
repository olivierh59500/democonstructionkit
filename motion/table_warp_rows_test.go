package motion

import "testing"

func TestTableWarpRowsMatchesDigiLogoAndSharedBounce(t *testing.T) {
	curve := []float64{0, 40, -40, 7.5, 12, -11, 3}
	bounceConfig := WaveClockConfig{Wave: Wave{Offset: 200, Amplitude: -160,
		Speed: 1, Rectify: true}, Step: .03}
	program, err := NewTableWarpRows(TableWarpRowsConfig{
		Curve: curve, Rows: 170, Step: 1,
		XBase: 320, HalfWidth: 167.5,
		YBase: 160, RowOffset: -.5, SourceHeight: 1,
		BounceDivisor: 2, Bounce: bounceConfig,
	})
	if err != nil {
		t.Fatal(err)
	}
	reference, err := NewWaveClock(bounceConfig)
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 1000; tick++ {
		bounce := reference.At(0)
		if program.Bounce() != bounce || program.Counter() != tick {
			t.Fatalf("tick %d clock=%v/%d, want %v/%d", tick, program.Bounce(), program.Counter(), bounce, tick)
		}
		for row := 0; row < 170; row++ {
			want := SampledRowPose{SourceY: float64(row), SourceHeight: 1,
				X:      320 + curve[(tick+row)%len(curve)] - 167.5,
				Y:      160 + bounce/2 + float64(row) - .5,
				ScaleX: 1, ScaleY: 1}
			if got := program.Sample(0, 0, row); got != want {
				t.Fatalf("tick %d row %d = %+v, want %+v", tick, row, got, want)
			}
		}
		reference.Step()
		program.Advance()
	}
	if got := testing.AllocsPerRun(100, func() {
		for row := 0; row < 170; row++ {
			program.Sample(0, 0, row)
		}
		program.Advance()
	}); got != 0 {
		t.Fatalf("table-warped rows allocated %.2f objects", got)
	}
}
