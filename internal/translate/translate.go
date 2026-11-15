package translate

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed prompt.md
var prompt string

//go:embed schema.json
var schema string

var Levels = []string{"a2", "b1", "b2"}

type Pair struct {
	En string `json:"en"`
	Ja string `json:"ja"`
}

type Article struct {
	SourceURL string `json:"source_url"`
	Title     Pair   `json:"title"`
	Sentences []Pair `json:"sentences"`
}

type output struct {
	Levels map[string][]Article `json:"levels"`
}

func Run(ctx context.Context, rawDir, outDir string) error {
	files, err := filepath.Glob(filepath.Join(rawDir, "*.json"))
	if err != nil {
		return err
	}
	byLevel := map[string][]Article{}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		var src struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(raw, &src); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		out, err := claude(ctx, raw)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		for _, lv := range Levels {
			for _, a := range out.Levels[lv] {
				a.SourceURL = src.URL
				byLevel[lv] = append(byLevel[lv], a)
			}
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, lv := range Levels {
		b, err := json.MarshalIndent(byLevel[lv], "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(outDir, lv+".json"), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func claude(ctx context.Context, article []byte) (*output, error) {
	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--output-format", "json", "--tools", "", "--json-schema", schema)
	cmd.Stdin = bytes.NewReader(article)
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var res struct {
		IsError          bool    `json:"is_error"`
		Result           string  `json:"result"`
		StructuredOutput *output `json:"structured_output"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	if res.IsError || res.StructuredOutput == nil {
		return nil, fmt.Errorf("claude returned no structured output: %s", res.Result)
	}
	return res.StructuredOutput, nil
}
