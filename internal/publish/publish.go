// Package publish は当日の音声とscript.jsonを wrangler でR2にアップロードする。
package publish

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/ogadra/Tachigiki/internal/translate"
)

const bucket = "tachigiki"

var files = map[string]string{
	"audio.m4a":   "audio/mp4",
	"script.json": "application/json",
}

// Run は outDir/{level}/ のファイルを R2 の {day}/{level}/ に置く。
func Run(ctx context.Context, day, outDir string) error {
	for _, lv := range translate.Levels {
		for name, contentType := range files {
			key := path.Join(bucket, day, lv, name)
			cmd := exec.CommandContext(ctx, "wrangler", "r2", "object", "put", key, "--remote", "--file", filepath.Join(outDir, lv, name), "--content-type", contentType)
			cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}
