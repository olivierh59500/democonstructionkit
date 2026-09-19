// Command checkaudio decodes original music through the shared adapters, without
// opening an audio device or importing Ebitengine's windowing backend.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/sound"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	root := flag.String("demos", "../../demos", "local demos directory")
	flag.Parse()
	for _, demo := range presets.Demos() {
		data, err := os.ReadFile(filepath.Join(*root, filepath.FromSlash(demo.Music)))
		if err != nil {
			return err
		}
		tracks := [][]byte{data}
		if demo.Module {
			tracks = nil
			for i := 0; i < 2; i++ {
				track, err := presets.RealityModule(data, i)
				if err != nil {
					return err
				}
				tracks = append(tracks, track)
			}
		}
		for i, data := range tracks {
			if err := check(data, demo.Module); err != nil {
				return fmt.Errorf("%s track %d: %w", demo.Name, i, err)
			}
		}
		fmt.Printf("OK %-31s %d track(s), 1 second decoded\n", demo.Name, len(tracks))
	}
	return nil
}
func check(data []byte, module bool) error {
	var stream *sound.Stream
	var err error
	if module {
		stream, err = sound.NewModule(data, sound.ModuleOptions{Interpolation: true})
	} else {
		stream, err = sound.NewYM(data, sound.YMOptions{Loop: true})
	}
	if err != nil {
		return err
	}
	defer stream.Close()
	buffer := make([]byte, 48000*8)
	if _, err = io.ReadFull(stream, buffer); err != nil {
		return err
	}
	for i := 0; i < len(buffer); i += 4 {
		v := math.Float32frombits(binary.LittleEndian.Uint32(buffer[i:]))
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) || v < -1 || v > 1 {
			return fmt.Errorf("invalid PCM sample %f", v)
		}
	}
	return nil
}
