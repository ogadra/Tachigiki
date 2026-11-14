package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMergesConfigsAndSkipsIgnoredDirs(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("config.yaml", "sources:\n  - id: a\n    kind: rss\n")
	write("sub/config.yaml", "sources:\n  - id: b\n    kind: rss\n")
	write("data/config.yaml", "sources:\n  - id: ignored\n")
	write(".git/config.yaml", "sources:\n  - id: ignored\n")
	write("worker/node_modules/x/config.yaml", "sources:\n  - id: ignored\n")

	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("got %+v", got)
	}
}
