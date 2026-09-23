// Command checkboundaries verifies that demo adapters depend on DCK's public
// music API instead of importing an underlying decoder or carrying local players.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func inspect(root string) ([]string, error) {
	var problems []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "build" || entry.Name() == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, i := range f.Imports {
			name, _ := strconv.Unquote(i.Path.Value)
			for _, backend := range []string{"github.com/olivierh59500/ym-player", "github.com/olivierh59500/go-zikmu"} {
				if name == backend || strings.HasPrefix(name, backend+"/") {
					problems = append(problems, path+": direct decoder import "+name)
				}
			}
		}
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "NewYMPlayer" {
				problems = append(problems, path+": local decoder factory")
			}
			if gen, ok := decl.(*ast.GenDecl); ok {
				for _, s := range gen.Specs {
					if typ, ok := s.(*ast.TypeSpec); ok && typ.Name.Name == "YMPlayer" {
						problems = append(problems, path+": local decoder wrapper")
					}
				}
			}
		}
		return nil
	})
	return problems, err
}

func run(root string) error {
	dirs, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	count, failed := 0, false
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		path := filepath.Join(root, dir.Name(), "dck")
		if dir.Name() == "go-uniondemo" {
			path = filepath.Join(root, dir.Name())
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		problems, err := inspect(path)
		if err != nil {
			return err
		}
		count++
		for _, problem := range problems {
			fmt.Println(problem)
			failed = true
		}
	}
	if count == 0 {
		return fmt.Errorf("no native DCK adapters found")
	}
	if failed {
		return fmt.Errorf("demo audio boundary violations found")
	}
	fmt.Printf("OK: %d native demo adapters use DCK instead of direct decoder imports or local YM players\n", count)
	return nil
}

func main() {
	root := flag.String("demos", "../../demos", "demo repository directory")
	flag.Parse()
	if err := run(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
