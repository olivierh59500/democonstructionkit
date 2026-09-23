package ebiten

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/sound"
	"github.com/olivierh59500/democonstructionkit/sound/output"
)

func waveFixture() []byte {
	b := make([]byte, 44+64*4)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:], uint32(len(b)-8))
	copy(b[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:], 16)
	binary.LittleEndian.PutUint16(b[20:], 1)
	binary.LittleEndian.PutUint16(b[22:], 2)
	binary.LittleEndian.PutUint32(b[24:], 48000)
	binary.LittleEndian.PutUint32(b[28:], 192000)
	binary.LittleEndian.PutUint16(b[32:], 4)
	binary.LittleEndian.PutUint16(b[34:], 16)
	copy(b[36:], "data")
	binary.LittleEndian.PutUint32(b[40:], 256)
	for i := 0; i < 64; i++ {
		binary.LittleEndian.PutUint16(b[44+i*4:], 12000)
		binary.LittleEndian.PutUint16(b[46+i*4:], uint16(65536-12000))
	}
	return b
}

func TestOpenPlaybackChoosesMatchingPCMDevicePath(t *testing.T) {
	for _, format := range []sound.PCMFormat{sound.Float32, sound.PCM16} {
		session, err := output.Begin(48000)
		if err != nil {
			t.Fatal(err)
		}
		p, err := Open(nil, "music.wav", waveFixture(), sound.Options{Loop: true, PCMFormat: format, Gain: .5})
		if err != nil {
			session.Close()
			t.Fatal(err)
		}
		p.Play()
		var pcm bytes.Buffer
		if err := session.Advance(1, 60, &pcm); err != nil {
			t.Fatal(err)
		}
		want := float32(6000) / 32768
		for i := 0; i < pcm.Len(); i += 8 {
			left := math.Float32frombits(binary.LittleEndian.Uint32(pcm.Bytes()[i:]))
			right := math.Float32frombits(binary.LittleEndian.Uint32(pcm.Bytes()[i+4:]))
			if left != want || right != -want {
				t.Fatalf("format%d: %g,%g", format, left, right)
			}
		}
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		session.Close()
	}
}
