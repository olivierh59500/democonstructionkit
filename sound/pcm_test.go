package sound

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"
)

func TestPCM16LoopAndSeek(t *testing.T) {
	pcm := []byte{0, 128, 255, 127, 0, 0, 0, 64}
	s, err := NewPCM16(bytes.NewReader(pcm), PCM16Options{Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got := make([]byte, 41)
	if _, err = io.ReadFull(s, got); err != nil {
		t.Fatal(err)
	}
	want := []float32{-1, 32767.0 / 32768, 0, .5}
	for i := 0; i < 10; i++ {
		v := math.Float32frombits(binary.LittleEndian.Uint32(got[i*4:]))
		if v != want[i%4] {
			t.Fatalf("sample %d = %g", i, v)
		}
	}
	if _, err = s.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	again := make([]byte, len(got))
	io.ReadFull(s, again)
	if !bytes.Equal(got, again) {
		t.Fatal("seek changed decoded samples")
	}
}
func TestPCM16EmptyLoopAndTruncatedFrame(t *testing.T) {
	for _, data := range [][]byte{nil, {1, 2, 3}} {
		s, err := NewPCM16(bytes.NewReader(data), PCM16Options{Loop: true})
		if err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 8)
		_, err = s.Read(b)
		if err == nil {
			t.Fatal("invalid PCM did not terminate")
		}
		s.Close()
	}
}
