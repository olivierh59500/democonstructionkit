// Command gallery runs the migrated original productions. Historical approximate
// studies remain available under cmd/studies and are never selected here.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/olivierh59500/democonstructionkit/presets"
)

var commands = map[string]string{
	"3d_doc": "threeddoc", "bilizir-demo": "bilizir-demo", "dma-3d": "dma3d", "dma-is-back": "dmaisback",
	"go-cocoisthebest": "cocoisthebest", "go-cuddlymenu": "cuddlymenu", "go-dom-intro": "domintro", "go-fr010": "fr010",
	"go-multiscreen": "multiscreen", "go-secondreality": "secondreality", "go-vectorballs": "vectorballs",
	"grodan-kvack-kvack-demo": "grodan", "megatwist": "megatwist", "nonameno-demo": "nonameno",
	"phenomena-dna-scroll-intro": "phenomena", "tcb-multi-plane-3d-scroller": "tcb-scroller",
	"tcb-replicants-demo": "tcbreplicants", "teamg1-demo": "teamg1demo", "viva_tcb": "vivatcb",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	name := flag.String("demo", "bilizir-demo", "original production name")
	demos := flag.String("demos", "../../demos", "path to the migrated original repositories")
	list := flag.Bool("list", false, "list productions")
	original := flag.Bool("original", false, "run the preserved original instead of its DCK version")
	audio := flag.Bool("audio", true, "original productions keep their original audio and controls")
	flag.Parse()
	if !*audio {
		return fmt.Errorf("the production launcher preserves original audio; use each demo's volume controls or cmd/fidelity for silent captures")
	}
	if *list {
		for _, d := range presets.Demos() {
			fmt.Println(d.Name)
		}
		return nil
	}
	command, ok := commands[*name]
	if !ok {
		return fmt.Errorf("unknown production %q; use -list (the original multiscreen is go-multiscreen)", *name)
	}
	dir, err := filepath.Abs(filepath.Join(*demos, *name))
	if err != nil {
		return err
	}
	mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(mod), "github.com/olivierh59500/democonstructionkit") {
		return fmt.Errorf("%s does not contain the construction-kit migration", dir)
	}
	entry := "./dck/cmd/" + command
	if *original {
		entry = "./cmd/" + command
	}
	c := exec.Command("go", append([]string{"run", entry}, flag.Args()...)...)
	c.Dir = dir
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
