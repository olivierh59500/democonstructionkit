// Command compare-frames verifies arbitrary PNG capture trees with DCK's
// channel-aware fidelity comparison. It checks RGB and alpha independently.
package main

import (
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"flag"

	"github.com/olivierh59500/democonstructionkit/fidelity"
)

func main() {
	reference := flag.String("reference", "", "reference PNG directory")
	candidate := flag.String("candidate", "", "candidate PNG directory")
	diffDir := flag.String("diff", "", "optional amplified difference directory")
	flag.Parse()
	if err := compareTrees(*reference, *candidate, *diffDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func compareTrees(reference, candidate, diffDir string) error {
	if reference == "" || candidate == "" {
		return fmt.Errorf("compare-frames: reference and candidate directories are required")
	}
	referenceNames, err := pngNames(reference)
	if err != nil {
		return err
	}
	candidateNames, err := pngNames(candidate)
	if err != nil {
		return err
	}
	if len(referenceNames) == 0 || len(referenceNames) != len(candidateNames) {
		return fmt.Errorf("compare-frames: capture trees have different PNG counts")
	}
	failed := false
	for index, name := range referenceNames {
		if name != candidateNames[index] {
			return fmt.Errorf("compare-frames: capture trees have different PNG paths")
		}
		a, err := readPNG(filepath.Join(reference, name))
		if err != nil {
			return err
		}
		b, err := readPNG(filepath.Join(candidate, name))
		if err != nil {
			return err
		}
		result, diff, err := fidelity.Compare(a, b)
		if err != nil {
			return fmt.Errorf("compare-frames: %s: %w", name, err)
		}
		fmt.Printf("%s: %d/%d pixels differ; max channel error %d\n", name, result.DifferentPixels, result.Pixels, result.MaxChannelError)
		if result.DifferentPixels == 0 {
			continue
		}
		failed = true
		if diffDir != "" {
			path := filepath.Join(diffDir, name)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			encodeErr := png.Encode(file, diff)
			closeErr := file.Close()
			if encodeErr != nil {
				return encodeErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
	if failed {
		return fmt.Errorf("compare-frames: visual differences found")
	}
	return nil
}

func pngNames(directory string) ([]string, error) {
	var names []string
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".png") {
			return nil
		}
		name, err := filepath.Rel(directory, path)
		if err == nil {
			names = append(names, name)
		}
		return err
	})
	return names, err
}

func readPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}
