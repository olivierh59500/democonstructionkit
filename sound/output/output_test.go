package output

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"testing"
	"time"
)

type constantPCM struct {
	value    float32
	floating bool
}

func (r constantPCM) Read(p []byte) (int, error) {
	if r.floating {
		for i := 0; i < len(p); i += 4 {
			binary.LittleEndian.PutUint32(p[i:], math.Float32bits(r.value))
		}
	} else {
		for i := 0; i < len(p); i += 2 {
			binary.LittleEndian.PutUint16(p[i:], uint16(int16(r.value*32768)))
		}
	}
	return len(p), nil
}

func recording(t *testing.T, rate int) *Session {
	t.Helper()
	s, err := Begin(rate)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestMixDelayedCuesVolumeAndPause(t *testing.T) {
	s := recording(t, 48000)
	ctx := NewContext(48000)
	pcm16, _ := ctx.NewPlayer(constantPCM{value: .5})
	float, _ := ctx.NewPlayerF32(constantPCM{value: .25, floating: true})
	pcm16.SetVolume(.5)
	var b bytes.Buffer
	if err := s.Advance(1, 60, &b); err != nil {
		t.Fatal(err)
	}
	for _, v := range b.Bytes() {
		if v != 0 {
			t.Fatal("delayed cue must leave silence")
		}
	}
	pcm16.Play()
	float.Play()
	b.Reset()
	if err := s.Advance(2, 60, &b); err != nil {
		t.Fatal(err)
	}
	if got := math.Float32frombits(binary.LittleEndian.Uint32(b.Bytes())); got != .5 {
		t.Fatalf("mixed sample %v", got)
	}
	if pcm16.Position() != time.Second/60 {
		t.Fatalf("position %s", pcm16.Position())
	}
	pcm16.Pause()
	b.Reset()
	if err := s.Advance(3, 60, &b); err != nil {
		t.Fatal(err)
	}
	if got := math.Float32frombits(binary.LittleEndian.Uint32(b.Bytes())); got != .25 {
		t.Fatalf("paused sample %v", got)
	}
	if pcm16.Position() != time.Second/60 {
		t.Fatal("paused player advanced")
	}
	float.Close()
	b.Reset()
	if err := s.Advance(4, 60, &b); err != nil {
		t.Fatal(err)
	}
	for _, v := range b.Bytes() {
		if v != 0 {
			t.Fatal("closed player still audible")
		}
	}
}

func TestFractionalSamplesStaySynchronized(t *testing.T) {
	s := recording(t, 44100)
	p, _ := NewContext(44100).NewPlayerF32(constantPCM{value: .125, floating: true})
	p.Play()
	start := Now()
	var b bytes.Buffer
	// A rate not dividing 44100 exposes cumulative rounding errors.
	for tick := int64(1); tick <= 59*10; tick++ {
		if err := s.Advance(tick, 59, &b); err != nil {
			t.Fatal(err)
		}
	}
	if b.Len() != 44100*10*8 || p.Position() != 10*time.Second {
		t.Fatalf("drift: %d bytes at %s", b.Len(), p.Position())
	}
	if Now().Sub(start) != 10*time.Second {
		t.Fatal("visual and audio clocks differ")
	}
	if err := s.Advance(1, 59, io.Discard); err == nil {
		t.Fatal("clock rewind was accepted")
	}
}

func TestEOFSeekAndValidation(t *testing.T) {
	s := recording(t, 48000)
	if _, err := Begin(48000); err == nil {
		t.Fatal("overlapping recording accepted")
	}
	if _, err := NewContext(44100).NewPlayer(bytes.NewReader(nil)); err == nil {
		t.Fatal("mismatched rate accepted")
	}
	p, err := NewContext(48000).NewPlayer(bytes.NewReader(make([]byte, 80*4)))
	if err != nil {
		t.Fatal(err)
	}
	p.Play()
	if err = s.Advance(1, 60, io.Discard); err != nil {
		t.Fatal(err)
	}
	if p.IsPlaying() {
		t.Fatal("EOF player remains active")
	}
	if p.Position() != time.Duration(80)*time.Second/48000 {
		t.Fatal("EOF position includes padded silence")
	}
	if err = p.Rewind(); err != nil {
		t.Fatal(err)
	}
	if p.Position() != 0 {
		t.Fatal("rewind did not reset position")
	}
	if err = p.SetPosition(-1); err == nil {
		t.Fatal("negative seek accepted")
	}
	p.Close()
	p.Close()
	if err = p.Rewind(); err != io.ErrClosedPipe {
		t.Fatalf("closed seek: %v", err)
	}
}
