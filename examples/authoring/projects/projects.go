// Package projects embeds saved authoring examples for desktop and Android.
package projects

import (
	"bytes"
	"embed"

	"github.com/olivierh59500/democonstructionkit/authoring"
)

//go:embed composed-effects.json
var files embed.FS

// Composed decodes the same file that the desktop example accepts with -project.
func Composed() (*authoring.Project, error) {
	data, err := files.ReadFile("composed-effects.json")
	if err != nil {
		return nil, err
	}
	return authoring.Decode(bytes.NewReader(data))
}
