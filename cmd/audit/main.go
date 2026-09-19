// Command audit inventories local demo repositories without loading Ebitengine.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type declaration struct {
	File string
	Line int
	Name string
}
type asset struct {
	File          string
	Width, Height int
}
type repository struct {
	Name, Revision string
	GoFiles, Lines int
	Declarations   []declaration
	Images         []asset
}

func main() {
	root := flag.String("demos", "../../demos", "local demos directory")
	flag.Parse()
	entries, err := os.ReadDir(*root)
	if err != nil {
		fail(err)
	}
	repos := []repository{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		base := filepath.Join(*root, entry.Name())
		if _, err := os.Stat(filepath.Join(base, "go.mod")); err != nil {
			continue
		}
		r := repository{Name: entry.Name()}
		revision, err := exec.Command("git", "-C", base, "rev-parse", "HEAD").Output()
		if err != nil {
			fail(err)
		}
		r.Revision = strings.TrimSpace(string(revision))
		err = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", "android", "build", "vendor", "node_modules":
					return filepath.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".go":
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				r.GoFiles++
				r.Lines += strings.Count(string(data), "\n")
				set := token.NewFileSet()
				f, err := parser.ParseFile(set, path, data, 0)
				if err != nil {
					return err
				}
				for _, d := range f.Decls {
					if fn, ok := d.(*ast.FuncDecl); ok {
						r.Declarations = append(r.Declarations, declaration{rel, set.Position(fn.Pos()).Line, fn.Name.Name})
					}
				}
			case ".png", ".jpg", ".jpeg":
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				c, _, decodeErr := image.DecodeConfig(f)
				closeErr := f.Close()
				if decodeErr != nil {
					return fmt.Errorf("%s: %w", path, decodeErr)
				}
				if closeErr != nil {
					return closeErr
				}
				r.Images = append(r.Images, asset{rel, c.Width, c.Height})
			}
			return nil
		})
		if err != nil {
			fail(err)
		}
		repos = append(repos, r)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(repos); err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
