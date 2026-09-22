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

func TestPCM16IntroductionPlaysOncePerSeek(t *testing.T) {
	frames := [][2]int16{{-32768, 32767}, {8192, -8192}, {16384, -16384}, {4096, -4096}}
	var data bytes.Buffer
	if err := binary.Write(&data, binary.LittleEndian, frames); err != nil {
		t.Fatal(err)
	}
	s, err := NewPCM16(bytes.NewReader(data.Bytes()), PCM16Options{Loop: true, LoopStartFrame: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// Include a partial sample to exercise the stream's decoded-ahead buffer.
	got := make([]byte, 11*frameBytes+3)
	if _, err := io.ReadFull(s, got); err != nil {
		t.Fatal(err)
	}
	for frame := 0; frame < 11; frame++ {
		sourceFrame := frame
		if frame >= len(frames) {
			sourceFrame = 2 + (frame-2)%2
		}
		for channel := 0; channel < 2; channel++ {
			value := math.Float32frombits(binary.LittleEndian.Uint32(got[frame*frameBytes+channel*4:]))
			want := float32(frames[sourceFrame][channel]) / 32768
			if value != want {
				t.Fatalf("frame %d channel %d = %g, want %g", frame, channel, value, want)
			}
		}
	}
	if _, err := s.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	again := make([]byte, len(got))
	if _, err := io.ReadFull(s, again); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, again) {
		t.Fatal("seeking to the beginning did not replay the introduction")
	}
}

func TestPCM16RejectsInvalidLoopStart(t *testing.T) {
	for name, options := range map[string]PCM16Options{
		"negative":         {Loop: true, LoopStartFrame: -1},
		"byte overflow":    {Loop: true, LoopStartFrame: math.MaxInt64/4 + 1},
		"looping disabled": {LoopStartFrame: 1},
	} {
		t.Run(name, func(t *testing.T) {
			if stream, err := NewPCM16(bytes.NewReader(nil), options); err == nil {
				stream.Close()
				t.Fatal("invalid loop start was accepted")
			}
		})
	}
}

func TestPCM16EmptyLoopRegionStopsAfterIntroduction(t *testing.T) {
	for name, start := range map[string]int64{
		"end":          1,
		"past end":     2,
		"largest seek": math.MaxInt64 / 4,
	} {
		t.Run(name, func(t *testing.T) {
			s, err := NewPCM16(bytes.NewReader([]byte{0, 128, 0, 64}), PCM16Options{Loop: true, LoopStartFrame: start})
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			buffer := make([]byte, 4*frameBytes)
			n, err := s.Read(buffer)
			if n != frameBytes || err != io.EOF {
				t.Fatalf("empty loop returned %d bytes, %v; want one introductory frame and EOF", n, err)
			}
			if n, err := s.Read(buffer); n != 0 || err != io.EOF {
				t.Fatalf("read after empty loop = %d, %v", n, err)
			}
			if _, err := s.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			if n, err := s.Read(buffer); n != frameBytes || err != io.EOF {
				t.Fatalf("rewinding an empty loop lost its introduction: %d, %v", n, err)
			}
		})
	}
}

func TestPCM16IntroductionLoopDoesNotAllocate(t *testing.T) {
	s, err := NewPCM16(bytes.NewReader(make([]byte, 32)), PCM16Options{Loop: true, LoopStartFrame: 3})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	buffer := make([]byte, 4097)
	if allocations := testing.AllocsPerRun(100, func() {
		if _, err := s.Read(buffer); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("%g allocations per callback", allocations)
	}
}
