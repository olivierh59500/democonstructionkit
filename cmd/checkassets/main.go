// Command checkassets validates every source atlas using CPU-only Go image decoding.
package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/olivierh59500/democonstructionkit/presets"
)

func main() {
	root := flag.String("demos", "../../demos", "local demos directory")
	flag.Parse()
	failed := false
	for _, s := range presets.Fonts() {
		err := check(*root, s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", s.ID, err)
			failed = true
		} else {
			fmt.Printf("OK %s\n", s.ID)
		}
	}
	if failed {
		os.Exit(1)
	}
}
func check(root string, s presets.Font) error {
	f, err := os.Open(filepath.Join(root, filepath.FromSlash(s.Path)))
	if err != nil {
		return err
	}
	defer f.Close()
	c, _, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}
	_, err = s.Build(image.Rect(0, 0, c.Width, c.Height))
	return err
}
