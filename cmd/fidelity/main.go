// Command fidelity compares migrated demos with their original Git revisions.
package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/olivierh59500/democonstructionkit/fidelity"
)

type probe struct {
	Width, Height    int
	Factory, Imports string
}

var probes = map[string]probe{
	"bilizir-demo":            {800, 600, `g:=NewGame();if err:=g.loadAssets();err!=nil{return nil,err};g.initScrollText();g.initialized=true;return g,nil`, ""},
	"viva_tcb":                {768, 540, `g:=NewGame();if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, ""},
	"grodan-kvack-kvack-demo": {640, 400, `g:=NewGame();g.audioInitialized=true;return g,nil`, ""},
	"dma-3d":                  {768, 540, `g,err:=NewGame();if err!=nil{return nil,err};g.audioInitAttempted=true;return g,nil`, ""},
	"tcb-replicants-demo":     {768, 540, `g:=NewGame();g.rng=rand.New(rand.NewSource(42));if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, `"math/rand"`},
}

type report struct {
	Demo, Reference, Candidate string
	Frames                     []frameResult
}
type frameResult struct {
	Frame int
	fidelity.Result
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	demos := flag.String("demos", "../../demos", "source demo repositories")
	demo := flag.String("demo", "bilizir-demo", "production to compare")
	kitRoot := flag.String("kit", ".", "construction kit checkout")
	out := flag.String("out", "captures/fidelity", "comparison directory")
	referenceOnly := flag.Bool("reference-only", false, "capture the pinned original only")
	flag.Parse()
	p, ok := probes[*demo]
	if !ok {
		return fmt.Errorf("no fidelity probe for %s", *demo)
	}
	root, err := filepath.Abs(*kitRoot)
	if err != nil {
		return err
	}
	source, err := filepath.Abs(filepath.Join(*demos, *demo))
	if err != nil {
		return err
	}
	output, err := filepath.Abs(filepath.Join(*out, *demo))
	if err != nil {
		return err
	}
	auditData, err := os.ReadFile(filepath.Join(root, "docs/source-audit.json"))
	if err != nil {
		return err
	}
	var audit []struct{ Name, Revision string }
	if err = json.Unmarshal(auditData, &audit); err != nil {
		return err
	}
	revision := ""
	for _, a := range audit {
		if a.Name == *demo {
			revision = a.Revision
		}
	}
	if revision == "" {
		return fmt.Errorf("missing original revision")
	}
	frames := []int{0, 1, 60, 240, 600, 1200, 2400, 4800}
	if err = captureRevision(source, revision, root, filepath.Join(output, "reference"), p, frames); err != nil {
		return err
	}
	if *referenceOnly {
		return nil
	}
	headBytes, err := command(source, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	head := strings.TrimSpace(string(headBytes))
	if err = captureRevision(source, head, root, filepath.Join(output, "candidate"), p, frames); err != nil {
		return err
	}
	r := report{Demo: *demo, Reference: revision, Candidate: head}
	failed := false
	for _, frame := range frames {
		name := fmt.Sprintf("%06d.png", frame)
		ref, err := readImage(filepath.Join(output, "reference", name))
		if err != nil {
			return err
		}
		candidate, err := readImage(filepath.Join(output, "candidate", name))
		if err != nil {
			return err
		}
		result, diff, err := fidelity.Compare(ref, candidate)
		if err != nil {
			return err
		}
		r.Frames = append(r.Frames, frameResult{frame, result})
		if result.DifferentPixels > 0 {
			failed = true
			f, err := os.Create(filepath.Join(output, "diff-"+name))
			if err != nil {
				return err
			}
			err = png.Encode(f, diff)
			closeErr := f.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
		fmt.Printf("%s frame %d: %d/%d pixels differ; max channel error %d\n", *demo, frame, result.DifferentPixels, result.Pixels, result.MaxChannelError)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(output, "report.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	if failed {
		return fmt.Errorf("fidelity mismatch; inspect %s", output)
	}
	return nil
}
func readImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
func command(dir, name string, args ...string) ([]byte, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOCACHE=/private/tmp/dck-go-build")
	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return out, nil
}
func captureRevision(source, revision, root, output string, p probe, frames []int) error {
	tmp, err := os.MkdirTemp("", "dck-fidelity-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	archive, err := command(source, "git", "archive", "--format=tar", revision)
	if err != nil {
		return err
	}
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !filepath.IsLocal(header.Name) {
			return fmt.Errorf("invalid archive path")
		}
		path := filepath.Join(tmp, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, 0755)
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(path), 0755); err == nil {
				var f *os.File
				f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
				if err == nil {
					_, err = io.Copy(f, reader)
					closeErr := f.Close()
					if err == nil {
						err = closeErr
					}
				}
			}
		}
		if err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return err
	}
	pkg := ""
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(tmp, e.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				return err
			}
			pkg = f.Name.Name
			break
		}
	}
	if pkg == "" {
		return fmt.Errorf("cannot find root package")
	}
	frameValues := make([]string, len(frames))
	for i, f := range frames {
		frameValues[i] = fmt.Sprint(f)
	}
	code := fmt.Sprintf(`package %s
import("os";"testing";"fmt";"github.com/hajimehoshi/ebiten/v2";capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten";%s)
func TestMain(m *testing.M){err:=capture.Run(capture.Config{Directory:%q,Width:%d,Height:%d,Frames:[]int{%s}},func()(ebiten.Game,error){%s});if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}
`, pkg, p.Imports, output, p.Width, p.Height, strings.Join(frameValues, ","), p.Factory)
	if err = os.WriteFile(filepath.Join(tmp, "dck_capture_test.go"), []byte(code), 0644); err != nil {
		return err
	}
	if _, err = command(tmp, "go", "mod", "edit", "-go=1.26.0", "-require=github.com/olivierh59500/democonstructionkit@v0.0.0", "-replace=github.com/olivierh59500/democonstructionkit="+root); err != nil {
		return err
	}
	// Only resolve packages in this capture, preserving the original dependency pins.
	if _, err = command(tmp, "go", "test", "-mod=mod", "-count=1", "-timeout=180s", "-run=^$", "."); err != nil {
		return err
	}
	fmt.Printf("captured %s at %s\n", filepath.Base(source), revision[:12])
	return nil
}
