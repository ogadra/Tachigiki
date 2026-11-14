package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/ogadra/Tachigiki/internal/assemble"
	"github.com/ogadra/Tachigiki/internal/config"
	"github.com/ogadra/Tachigiki/internal/fetch"
	_ "github.com/ogadra/Tachigiki/internal/fetch/adapters"
	"github.com/ogadra/Tachigiki/internal/translate"
	"github.com/ogadra/Tachigiki/internal/tts"
)

func main() {
	ctx := context.Background()
	now := time.Now().In(time.FixedZone("JST", 9*60*60))
	day := now.Format(time.DateOnly)
	rawDir := filepath.Join("data", "raw", day)
	scriptDir := filepath.Join("data", "script", day)
	audioDir := filepath.Join("data", "audio", day)
	outDir := filepath.Join("data", "out", day)

	sources, err := config.Load(".")
	if err != nil {
		log.Fatal(err)
	}
	if err := fetch.Run(ctx, now, sources, rawDir); err != nil {
		log.Fatal(err)
	}
	if err := translate.Run(ctx, rawDir, scriptDir); err != nil {
		log.Fatal(err)
	}
	if err := tts.Run(ctx, scriptDir, audioDir); err != nil {
		log.Fatal(err)
	}
	if err := assemble.Run(ctx, day, scriptDir, audioDir, outDir); err != nil {
		log.Fatal(err)
	}
}
