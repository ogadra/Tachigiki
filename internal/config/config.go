// Package config はリポジトリ内の config.yaml をすべて探して読む。
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

// Load は root 以下の config.yaml をすべて読み、sources を1つの一覧にまとめる。
func Load(root string) ([]Source, error) {
	var sources []Source
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// data/ は出力先、node_modules/ は Worker の依存で、どちらにも他人の config.yaml が入りうる
			if path != root && (strings.HasPrefix(d.Name(), ".") || d.Name() == "data" || d.Name() == "node_modules") {
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
