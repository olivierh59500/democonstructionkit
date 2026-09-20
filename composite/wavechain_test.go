package composite

import (
	"image"
	"testing"
)

func TestWaveChainsOwnTheirPhases(t *testing.T) {
	p := WavePass{Size: image.Pt(37, 23), Wave: WaveStrips{Thickness: 8, Waves: []StripWave{{Phase: 2, Speed: .5}}}}
	a, err := NewWaveChain(p)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := NewWaveChain(p)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	a.Advance()
	a.Pass(0).Thickness = 16
	if b.Pass(0).Waves[0].Phase != 2 || p.Wave.Waves[0].Phase != 2 || a.Pass(0).Waves[0].Phase != 2.5 {
		t.Fatal("wave chain phases alias another composition")
	}
	if b.Pass(0).Thickness != 8 {
		t.Fatal("band size leaked between effects")
	}
}
