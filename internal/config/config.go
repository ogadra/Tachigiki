package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type Source struct {
	ID   string `yaml:"id"`
	Kind string `yaml:"kind"`
	URL  string `yaml:"url"`
	Lang string `yaml:"lang"`
}

type file struct {
	Sources []Source `yaml:"sources"`
}

var skippedDirs = map[string]bool{"data": true, "node_modules": true}

func Load(root string) ([]Source, error) {
	var sources []Source
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			hidden := strings.HasPrefix(d.Name(), ".")
			if path != root && (hidden || skippedDirs[d.Name()]) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "config.yaml" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var f file
		if err := yaml.Unmarshal(b, &f); err != nil {
			return err
		}
		sources = append(sources, f.Sources...)
		return nil
	})
	return sources, err
}
