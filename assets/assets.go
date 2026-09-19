// Package assets loads and caches application-owned images from any fs.FS.
package assets

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

type Store struct {
	Files    fs.FS
	images   map[string]image.Image
	textures map[string]*ebiten.Image
}

func New(files fs.FS) *Store {
	return &Store{Files: files, images: map[string]image.Image{}, textures: map[string]*ebiten.Image{}}
}
func (s *Store) Image(name string) (image.Image, error) {
	if img, ok := s.images[name]; ok {
		return img, nil
	}
	if s.Files == nil {
		return nil, fmt.Errorf("assets: nil filesystem")
	}
	data, err := fs.ReadFile(s.Files, name)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("assets: %s: %w", name, err)
	}
	s.images[name] = img
	return img, nil
}
func (s *Store) Texture(name string) (*ebiten.Image, error) {
	if img, ok := s.textures[name]; ok {
		return img, nil
	}
	img, err := s.Image(name)
	if err != nil {
		return nil, err
	}
	texture := ebiten.NewImageFromImage(img)
	s.textures[name] = texture
	return texture, nil
}

// Close releases GPU resources. Effects borrowing them must be closed first.
func (s *Store) Close() error {
	for _, img := range s.textures {
		img.Deallocate()
	}
	clear(s.textures)
	clear(s.images)
	return nil
}
