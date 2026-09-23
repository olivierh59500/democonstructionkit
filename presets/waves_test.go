package presets

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestWaveRecipesPreserveAuthoredSamples(t *testing.T) {
	bilizir, err := motion.CompileWaveTable(BilizirWaveSections()...)
	if err != nil {
		t.Fatal(err)
	}
	var reference []float64
	for _, section := range []struct {
		n            int
		a, b, sa, sb float64
	}{
		{389, 20, 30, 7.0 / 180.0 * math.Pi, 3.0 / 180.0 * math.Pi},
		{120, 4, 0, 72.0 / 180.0 * math.Pi, 0},
		{68, 40, 0, 8.0 / 180.0 * math.Pi, 0},
		{389, 20, 30, 7.0 / 180.0 * math.Pi, 3.0 / 180.0 * math.Pi},
		{36, 4, 0, 72.0 / 180.0 * math.Pi, 0},
		{189, 30, 0, 8.0 / 180.0 * math.Pi, 0},
	} {
		for i := 0; i < section.n; i++ {
			reference = append(reference, section.a*math.Sin(float64(i)*section.sa)+section.b*math.Cos(float64(i)*section.sb))
		}
	}
	if len(bilizir) != len(reference) {
		t.Fatalf("wave length %d", len(bilizir))
	}
	for i, v := range reference {
		if bilizir[i] != v {
			t.Fatalf("Bilizir sample %d: %.20g != %.20g", i, bilizir[i], v)
		}
	}
	tcb, err := motion.CompileWaveTable(TCBLogoWaveSections()...)
	if err != nil {
		t.Fatal(err)
	}
	if len(tcb) != 1814 {
		t.Fatalf("TCB length %d", len(tcb))
	}
	for i, v := range tcb {
		want := 0.0
		if i >= 40 && i < 844 {
			want = 8 * math.Sin(float64(i-40)*.05-2)
		}
		if i >= 844 && i < 1654 {
			want = 8 * math.Sin(float64(i-844)*.15)
		}
		if v != want {
			t.Fatalf("TCB sample %d: %.20g != %.20g", i, v, want)
		}
	}
	// Each preset is mutable without changing another instance or a repeated part.
	a, b := BilizirWaveSections(), BilizirWaveSections()
	a[0].Terms[0].Amplitude = 999
	if b[0].Terms[0].Amplitude != 20 || a[3].Terms[0].Amplitude != 20 {
		t.Fatal("preset sections alias")
	}
}

func TestCopperRecipePreservesAuthoredTable(t *testing.T) {
	values := BilizirCopperOffsets()
	if len(values) != 1024 {
		t.Fatalf("length %d", len(values))
	}
	h := sha256.New()
	var word [4]byte
	for _, v := range values {
		binary.LittleEndian.PutUint32(word[:], uint32(v))
		h.Write(word[:])
	}
	// Digest of all 1,024 signed little-endian samples in the former demo table.
	if got := fmt.Sprintf("%x", h.Sum(nil)); got != "3d34745df3cef97b4d93a990dbdcedde34214ed580e914cf3c9ff1d65deb98ba" {
		t.Fatalf("table changed: %s", got)
	}
	values[0] = 0
	if BilizirCopperOffsets()[0] != 264 {
		t.Fatal("preset data aliases")
	}
}
