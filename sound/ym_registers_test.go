package sound

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

func registerFrame(frame int) [14]uint8 {
	return [14]uint8{
		uint8(32 + frame), 1,
		uint8(64 + frame), 2,
		uint8(96 + frame), 3,
		uint8(frame & 31), 0x38,
		uint8(frame & 15), uint8((frame + 3) & 15), uint8((frame + 6) & 15),
		100, 0, 9,
	}
}

func registerYM() []byte {
	const frames = 100
	data := make([]byte, 4+frames*14)
	copy(data, "YM3!")
	for f := 0; f < frames; f++ {
		for r, value := range registerFrame(f) {
			data[4+r*frames+f] = value
		}
	}
	return data
}

func TestYMRegistersFollowDecodedSamples(t *testing.T) {
	s, err := NewYM(registerYM(), YMOptions{SampleRate: 48000, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	initial, ok := s.YMRegisters()
	if !ok {
		t.Fatal("new YM stream has no register snapshot")
	}
	if _, err := s.Read(make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	// The first read decodes 1024 samples. At 48000 Hz, each 50 Hz YM frame
	// spans 960 samples, so the snapshot already contains the second frame.
	got, ok := s.YMRegisters()
	if want := registerFrame(1); !ok || got != want {
		t.Fatalf("first decode snapshot = %v, %v; want %v, true", got, ok, want)
	}
	position := s.Position()
	got[8] = 255
	again, ok := s.YMRegisters()
	if !ok || again != registerFrame(1) || s.Position() != position {
		t.Fatal("snapshot changed playback or exposed mutable decoder storage")
	}
	if _, err := io.ReadFull(s, make([]byte, blockFrames*frameBytes-1)); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.YMRegisters(); !ok || got != again {
		t.Fatal("consuming staged audio advanced the decoder snapshot")
	}
	if _, err := s.Read(make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.YMRegisters(); !ok || got != registerFrame(2) {
		t.Fatalf("second decode snapshot = %v, %v", got, ok)
	}
	if _, err := s.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.YMRegisters(); !ok || got != initial {
		t.Fatalf("rewind snapshot = %v, %v; want initial %v", got, ok, initial)
	}
	if _, err := s.Read(make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	if got, ok := s.YMRegisters(); !ok || got != registerFrame(1) {
		t.Fatalf("replayed snapshot = %v, %v", got, ok)
	}
}

func TestYMRegistersUnavailableWithoutYMDecoder(t *testing.T) {
	pcm, err := NewPCM16(bytes.NewReader(make([]byte, 16)), PCM16Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer pcm.Close()
	module, err := NewModule(testMOD(), ModuleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer module.Close()
	ym, err := NewYM(registerYM(), YMOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ym.Close(); err != nil {
		t.Fatal(err)
	}
	for name, s := range map[string]*Stream{
		"nil": nil, "zero": {}, "pcm": pcm, "module": module, "closed": ym,
	} {
		t.Run(name, func(t *testing.T) {
			if got, ok := s.YMRegisters(); ok || got != ([14]uint8{}) {
				t.Fatalf("register snapshot = %v, %v; want zero, false", got, ok)
			}
		})
	}
}

func TestYMRegistersConcurrentPlaybackSeekAndClose(t *testing.T) {
	s, err := NewYM(registerYM(), YMOptions{Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	start := make(chan struct{})
	readOnce := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(4)
	go func() {
		defer workers.Done()
		<-start
		buffer := make([]byte, 257)
		for i := 0; i < 1000; i++ {
			_, _ = s.Read(buffer)
			if i == 0 {
				close(readOnce)
			}
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 100; i++ {
			_, _ = s.Seek(0, io.SeekStart)
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 1000; i++ {
			registers, ok := s.YMRegisters()
			if !ok && registers != ([14]uint8{}) {
				t.Error("unavailable snapshot retained stale registers")
				return
			}
			if ok && registers[7] != 0xff {
				frame := int(registers[0]) - 32
				if frame < 0 || frame >= 100 || registers != registerFrame(frame) {
					t.Errorf("inconsistent snapshot during playback: %v", registers)
					return
				}
			}
		}
	}()
	go func() {
		defer workers.Done()
		<-readOnce
		_ = s.Close()
	}()
	close(start)
	workers.Wait()
	if registers, ok := s.YMRegisters(); ok || registers != ([14]uint8{}) {
		t.Fatalf("closed stream snapshot = %v, %v", registers, ok)
	}
}
