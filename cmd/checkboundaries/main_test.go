package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDirectBackendAndLocalWrapper(t *testing.T) {
	root := t.TempDir()
	source := `package demo
import "github.com/olivierh59500/ym-player/pkg/stsound"
type YMPlayer struct{}
func NewYMPlayer(){}
`
	if err := os.WriteFile(filepath.Join(root, "audio.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	issues, err := inspect(root)
	if err != nil || len(issues) != 3 {
		t.Fatal(issues, err)
	}
	if err := os.WriteFile(filepath.Join(root, "audio.go"), []byte("package demo\nimport \"github.com/olivierh59500/democonstructionkit/sound\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	issues, err = inspect(root)
	if err != nil || len(issues) != 0 {
		t.Fatal(issues, err)
	}
}
